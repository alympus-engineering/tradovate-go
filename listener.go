package tradovate

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type Message struct {
	Id        int64           `json:"i,omitempty"`
	Status    int             `json:"s,omitempty"`
	EventType string          `json:"e,omitempty"`
	Data      json.RawMessage `json:"d,omitempty"`
}

func Listener(c *Client) {
	c.HeartbeatTicker = time.NewTicker(1 * time.Second)

	for {
		select {
		case _ = <-c.HeartbeatTicker.C:
			if c.Connected {
				c.ConnectionMutex.Lock()
				err := c.Connection.WriteMessage(websocket.TextMessage, []byte("[]"))
				c.ConnectionMutex.Unlock()
				if err != nil {
					log.Println("ERROR: listener exiting ", err)
					return
				}
			}

		default:
			_, msg, err := c.Connection.ReadMessage()

			if err != nil {
				log.Println("ERROR: listener exiting ", err)
				return
			}
			fmt.Printf("%s \n\n", msg)
			// Determine Message Frame Type
			// https://api.tradovate.com/#section/Connecting-to-the-WebSocket-Server
			frameType, body := msg[0], msg[1:]

			switch frameType {
			case 'o':
				c.Connected = true
			case 'h':
				c.LastHeartbeatResponse = time.Now()
			case 'a':
				var frame []Message
				err = json.Unmarshal(body, &frame)

				if err != nil {
					return
				}

				for _, m := range frame {

					if m.EventType == "chart" {
						reqStatus := handleChart(m)

						for id, s := range reqStatus {
							if s {
								c.RequestPool[id] <- Message{
									Id:        id,
									Status:    200,
									EventType: "done",
									Data:      nil,
								}
							}
						}

					} else {
						if req, ok := c.RequestPool[m.Id]; ok {
							req <- m
						}
					}

				}

			case 'c':
				c.Connected = false
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

func handleChart(m Message) map[int64]bool {
	var d map[string]interface{}

	err := json.Unmarshal(m.Data, &d)

	if err != nil {

	}

	println(string(m.Data))

	charts := d["charts"].([]interface{})
	requestStatus := map[int64]bool{}

	for _, c := range charts {
		chart := c.(map[string]interface{})

		id := int64(chart["id"].(float64))

		if eoh, ok := chart["eoh"]; ok {
			requestStatus[id] = eoh.(bool)
		}

	}

	return requestStatus
}
