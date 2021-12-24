package tradovate

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/alympus-engineering/tradovate-go/http"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type RequestMetadata struct {
	Endpoint string
	SocketId string
}

type AsyncClient struct {
	AppName    string
	AppVersion string

	HttpHost       string
	WebsocketHost  string
	MarketDataHost string

	Username string
	Password string
	ClientId string
	ApiKey   string
	DeviceId string

	WebsocketConnected  bool
	WebsocketAuthorized  bool
	WebsocketConnection *websocket.Conn

	MarketDataConnected  bool
	MarketDataAuthorized bool
	MarketDataConnection *websocket.Conn

	Events chan Message

	Connected bool

	AccessToken     string
	TokenExpiration time.Time

	HeartbeatTicker       *time.Ticker
	LastHeartbeatResponse time.Time

	PenaltyTicket string

	RequestPool   map[int64]RequestMetadata
	LastRequestId int64

	MarketData map[string][]Tick
}

func NewAsyncClient(
	environment string, appName string, appVersion string,
	username string, password string, clientId string, apiKey string) *AsyncClient {
	var httpHost string
	var wsHost string
	var mdHost string

	if strings.ToLower(environment) == "live" {
		httpHost = "https://live.tradovateapi.com/v1"
		wsHost = "wss://live.tradovateapi.com/v1"
	} else if strings.ToLower(environment) == "demo" {
		httpHost = "https://demo.tradovateapi.com/v1"
		wsHost = "wss://demo.tradovateapi.com/v1"
	}

	mdHost = "wss://md.tradovateapi.com/v1"

	var deviceId uuid.UUID

	b, err := os.ReadFile("/tmp/tradovate-device-id")
	if err != nil {
		deviceId = uuid.New()
		os.WriteFile("/tmp/tradovate-device-id", []byte(deviceId.String()), os.ModePerm)

	} else {
		deviceId = uuid.MustParse(string(b))
	}

	return &AsyncClient{
		AppName:        appName,
		AppVersion:     appVersion,
		HttpHost:       httpHost,
		WebsocketHost:  wsHost,
		MarketDataHost: mdHost,
		Username:       username,
		Password:       password,
		ClientId:       clientId,
		ApiKey:         apiKey,
		DeviceId:       deviceId.String(),
		RequestPool:    make(map[int64]RequestMetadata),
	}
}

func (c *AsyncClient) Connect() (err error) {
	resp, err := http.GetAccessToken(
		c.HttpHost, c.Username, c.Password, c.AppName, c.AppVersion,
		c.ClientId, c.DeviceId, c.ApiKey, c.PenaltyTicket)

	if err != nil {
		switch e := err.(type) {
		case *RateLimitError:
			log.Println("ERROR: rate limit exceeded getting access token, retrying and attaching penalty ticket")
			time.Sleep(time.Duration(e.PenaltyTime) * time.Second)
		}

		resp, err = http.GetAccessToken(
			c.HttpHost, c.Username, c.Password, c.AppName, c.AppVersion,
			c.ClientId, c.DeviceId, c.ApiKey, c.PenaltyTicket)
	}

	c.Events = make(chan Message)

	wsUrl := c.WebsocketHost + CONNECT_WEBSOCKET
	mdUrl := c.MarketDataHost + CONNECT_WEBSOCKET

	ws, err := c.CreateWebsocket(wsUrl, c.Events)
	if err != nil {
		return err
	}

	mdWs, err := c.CreateWebsocket(mdUrl, c.Events)
	if err != nil {
		return err
	}

	c.WebsocketConnection = ws
	c.MarketDataConnection = mdWs

	err = c.SendAuthorization("async-md", resp.AccessToken)
	if err != nil {
		return err
	}

	err = c.SendAuthorization("async-ws", resp.AccessToken)
	if err != nil {
		return err
	}


	return err
}

func (c *AsyncClient) CreateWebsocket(url string, events chan Message) (*websocket.Conn, error) {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	go AsyncListener(ws, events, c.onConnect, c.disconnect, c.onEvent)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

connectionLoop:
	for {
		select {
		case _ = <-ctx.Done():
			return nil, errors.New("timeout while waiting for websocket to connect")
		default:

			if strings.Contains(url, "md") && c.MarketDataConnected{
				break connectionLoop
			} else if c.WebsocketConnected {
				break connectionLoop
			}
			time.Sleep(1 * time.Second)
		}
	}

	return ws, nil
}

func (c *AsyncClient) SendAuthorization(socketId string, accessToken string) error {
	id, err := c.SendAsync(socketId, AUTHORIZE, "", accessToken)

	if err != nil {
		return err
	}

	c.RequestPool[id] = RequestMetadata{
		Endpoint: AUTHORIZE,
		SocketId: socketId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

authorizationLoop:
	for {
		select {
		case _ = <-ctx.Done():
			return errors.New("timeout while waiting for websocket to authorize")
		default:

			if socketId == "async-md" && c.MarketDataAuthorized {
				break authorizationLoop
			} else if socketId == "async-ws" && c.WebsocketAuthorized {
				break authorizationLoop
			}
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

func (c *AsyncClient) SendAsync(socketId string, endpoint string, queryParams string, body string) (id int64, err error) {
	var sb strings.Builder

	id = c.getRequestId()

	sb.WriteString(endpoint)
	sb.WriteString("\n")
	sb.WriteString(strconv.FormatInt(id, 10))
	sb.WriteString("\n")
	sb.WriteString(queryParams)
	sb.WriteString("\n")
	sb.WriteString(body)

	if socketId == "async-md" {
		err = c.MarketDataConnection.WriteMessage(websocket.TextMessage, []byte(sb.String()))
	} else if socketId == "async-ws" {
		err = c.WebsocketConnection.WriteMessage(websocket.TextMessage, []byte(sb.String()))
	}

	if err != nil {
		return id, err
	}

	return id, nil
}

func (c *AsyncClient) onConnect(m Message) {
	if string(m.Data) == "async-md" {
		c.MarketDataConnected = true
	} else if string(m.Data) == "async-ws" {
		c.WebsocketConnected = true
	}
}

func (c *AsyncClient) disconnect(m Message) {
	println(m.EventType)

}

func (c *AsyncClient) onAuthorize() {
}

func (c *AsyncClient) onEvent(m Message) {

	if metadata, ok := c.RequestPool[m.Id]; ok {
		if metadata.Endpoint == AUTHORIZE {

			if metadata.SocketId == "async-md" {
				c.MarketDataAuthorized = true
			} else if metadata.SocketId == "async-ws" {
				c.WebsocketAuthorized = true
			}

		}
	}

	println(string(m.Data))

}

func (c *AsyncClient) getRequestId() int64 {
	id := time.Now().Unix()
	if id == c.LastRequestId {
		c.LastRequestId = id + 1
		return id + 1
	}
	c.LastRequestId = id
	return id
}

func AsyncListener(c *websocket.Conn, out chan Message, onConnect func(m Message), onDisconnect func(m Message), onEvent func(m Message)) {
	heartbeatTicker := time.NewTicker(1 * time.Second)

	var socketId string

	if strings.Contains(c.RemoteAddr().String(), "md") {
		socketId = "async-md"
	} else {
		socketId = "async-ws"
	}

	for {
		select {
		case _ = <-heartbeatTicker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("[]"))
			if err != nil {
				log.Println("ERROR: listener exiting ", err)
				return
			}

		default:
			_, msg, err := c.ReadMessage()

			if err != nil {
				log.Println("ERROR: listener exiting ", err)
				return
			}

			// Determine Message Frame Type
			// https://api.tradovate.com/#section/Connecting-to-the-WebSocket-Server
			frameType, body := msg[0], msg[1:]

			switch frameType {
			case 'o':
				//out <- Message{Status: 200, EventType: "connected", Data: []byte(socketId)}
				onConnect(Message{Status: 200, EventType: "connected", Data: []byte(socketId)})
			case 'h':
				//unixTime := strconv.FormatInt(time.Now().Unix(), 10)
				//out <- Message{Status: 200, EventType: "heartbeat", Data: []byte(unixTime)}
			case 'a':
				var frame []Message
				err = json.Unmarshal(body, &frame)
				//
				if err != nil {
					return
				}
				//
				for _, m := range frame {
					onEvent(m)
					//
					//	if m.EventType == "chart" {
					//		reqStatus := handleChart(m)
					//
					//		for id, s := range reqStatus {
					//			if s {
					//				c.RequestPool[id] <- Message{
					//					Id:        id,
					//					Status:    200,
					//					EventType: "done",
					//					Data:      nil,
					//				}
					//			}
					//		}
					//
					//	} else {
					//		if req, ok := c.RequestPool[m.Id]; ok {
					//			req <- m
					//		}
					//	}
					//
				}

			case 'c':
				//out <- Message{Status: 200, EventType: "disconnected", Data: []byte(socketId)}
				onDisconnect(Message{Status: 200, EventType: "disconnected", Data: []byte(socketId)})
				heartbeatTicker.Stop()

				return
			default:
				return
			}

			//if err == nil {
			//	err = handleResponse(resp, c, listenerType)
			//} else {
			//	err = handleEvent(msg, c)
			//}

			//if err != nil {
			//	log.Error("error handling message", zap.ByteString("message", msg), zap.Error(err))
			//	break
			//}
		}
	}

}
