package tradovate

import (
	"fmt"
)

func (c *Client) GetFilledOrders() error {
	msg, err := c.Send(GET_FILLED_ORDERS, "", "", 10)

	if err == nil {
		fmt.Printf("%s\n", msg.Data)
	}
	return err
}
