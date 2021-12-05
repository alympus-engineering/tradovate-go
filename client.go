package tradovate

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"strings"
	"time"
)

type Connection struct {
	Connected  bool
	Connection *websocket.Conn
}

type Client struct {
	AppName    string
	AppVersion string

	HttpHost      string
	WebsocketHost string
	Username      string
	Password      string
	ClientId      string
	ApiKey        string
	DeviceId      string

	Connection *websocket.Conn
	Connected  bool

	AccessToken     string
	TokenExpiration time.Time

	HeartbeatTicker *time.Ticker

	PenaltyTicket  string

	RequestPool map[int64]chan Message
}

func NewClient(
	environment string, appName string, appVersion string,
	username string, password string, clientId string, apiKey string) *Client {
	var httpHost string
	var wsHost string

	if strings.ToLower(environment) == "live" {
		httpHost = "https://live.tradovateapi.com/v1"
		wsHost = "wss://live.tradovateapi.com/v1"
	} else if strings.ToLower(environment) == "demo" {
		httpHost = "https://demo.tradovateapi.com/v1"
		wsHost = "wss://demo.tradovateapi.com/v1"
	}

	deviceId := uuid.New()

	return &Client{
		AppName:       appName,
		AppVersion:    appVersion,
		HttpHost:      httpHost,
		WebsocketHost: wsHost,
		Username:      username,
		Password:      password,
		ClientId:      clientId,
		ApiKey:        apiKey,
		DeviceId:      deviceId.String(),
		RequestPool:   make(map[int64]chan Message),
	}
}

func (c *Client) ConnectWebsocket() (err error) {
	err = c.GetAccessToken()

	if err != nil {
		switch e := err.(type) {
		case *RateLimitError:
			log.Println("ERROR: rate limit exceeded getting access token, retrying and attaching penalty ticket")
			time.Sleep(time.Duration(e.PenaltyTime) * time.Second)
		}

		err = c.GetAccessToken()
	}


	url := c.WebsocketHost + CONNECT_WEBSOCKET

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	c.Connection = ws

	go Listener(c)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connectionLoop:
		for {
			select {
			case _ = <-ctx.Done():
				return errors.New("timeout while waiting for websocket to connect")
			default:
				if c.Connected {
					break connectionLoop
				}
				time.Sleep(1 * time.Second)
			}
		}



	err = c.SendAuthorization()

	return err
}

func (c *Client) SendAuthorization() error {
	msg, err := c.Send(AUTHORIZE, "", c.AccessToken, 30)

	if err != nil {
		return err
	}

	if msg.Status != 200 {
		return errors.New("authorization unsuccessful, code: " + strconv.Itoa(msg.Status))
	}

	return err
}

func (c *Client) Send(endpoint string, queryParams string, body string, timeout int) (*Message, error) {
	var sb strings.Builder

	id := c.getRequestId()

	sb.WriteString(endpoint)
	sb.WriteString("\n")
	sb.WriteString(strconv.FormatInt(id, 10))
	sb.WriteString("\n")
	sb.WriteString(queryParams)
	sb.WriteString("\n")
	sb.WriteString(body)

	channel := make(chan Message)
	c.RequestPool[id] = channel

	err := c.Connection.WriteMessage(websocket.TextMessage, []byte(sb.String()))

	if err != nil {
		return nil, err
	}

	select {
	case response := <-channel:
		return &response, nil
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil, errors.New("timeout exceeded")
	}
}

func (c *Client) SendAsync(endpoint string, queryParams string, body string) (int64, error) {
	var sb strings.Builder

	id := c.getRequestId()

	sb.WriteString(endpoint)
	sb.WriteString("\n")
	sb.WriteString(strconv.FormatInt(id, 10))
	sb.WriteString("\n")
	sb.WriteString(queryParams)
	sb.WriteString("\n")
	sb.WriteString(body)

	err := c.Connection.WriteMessage(websocket.TextMessage, []byte(sb.String()))

	if err != nil {
		return id, err
	}

	return id, nil
}

func (c *Client) getRequestId() int64 {
	return time.Now().Unix()
}
