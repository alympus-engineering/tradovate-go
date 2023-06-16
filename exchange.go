package tradovate

import "fmt"

func (c *Client) GetExchanges() error {
	msg, err := c.Send(GET_EXCHANGE_LIST, "", "", 10)

	if err == nil {
		fmt.Printf("%s\n", msg.Data)
	}
	return err
}
