package tradovate

import (
	"fmt"
)

func (c *Client) GetOrders() error {
	msg, err := c.Send(GET_ORDERS, "", "", 10)

	if err == nil {
		fmt.Printf("%s\n", msg.Data)
	}
	return err
}
