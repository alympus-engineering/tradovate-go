package tradovate

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type Message struct {
	Id int64 `json:"i"`
	Status int `json:"s"`
	Data json.RawMessage `json:"d"`
}

func Listener(c *Client) {
	c.HeartbeatTicker = time.NewTicker(1 * time.Second)

	for {
		select {
		case _ = <-c.HeartbeatTicker.C:
			if c.Connected {
				err := c.Connection.WriteMessage(websocket.TextMessage, []byte("[]"))
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

			// Determine Message Frame Type
			// https://api.tradovate.com/#section/Connecting-to-the-WebSocket-Server
			frameType, body := msg[0], msg[1:]

			switch frameType {
			case 'o':
				c.Connected = true
			case 'h':
				println("heartbeat")

			case 'a':
				var frame []Message
				err = json.Unmarshal(body, &frame)

				if err != nil {
					return
				}

				for _, m := range frame {
					if req, ok := c.RequestPool[m.Id]; ok {
						req <- m
					}
				}

			case 'c':
				print("")
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
