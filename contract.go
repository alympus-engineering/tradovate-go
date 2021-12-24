package tradovate

import (
	"encoding/json"
	"net/url"
)

type Contract struct {
	Id 					int64 	`json:"id"`
	Name 				string 	`json:"name"`
	ContractMaturityId 	int64 	`json:"contractMaturityId"`
	Status				string 	`json:"status"`
	ProviderTickSize    float64 	`json:"providerTickSize"`
}

func (c *Client) FindContract(name string) (Contract, error) {
	params := url.Values{}
	params.Add("name", name)

	msg, err := c.Send(FIND_CONTRACT, params.Encode(), "", 10)

	if msg.Status != 200 {

	}

	var contract Contract
	err = json.Unmarshal(msg.Data, &contract)

	if err != nil {
		return Contract{}, err
	}

	return contract, nil
}

func (c *AsyncClient) FindContract(name string) (int64, error) {
	p, b, err := findContract(name)

	if err != nil {
		return -1, err
	}

	return c.SendAsync("async-ws", FIND_CONTRACT, p, b)
}


func findContract(name string) (p string, b string, err error) {
	params := url.Values{}
	params.Add("name", name)

	return params.Encode(), b, nil

	//msg, err := c.Send(FIND_CONTRACT, params.Encode(), "", 10)
	//
	//if msg.Status != 200 {
	//
	//}
	//
	//var contract Contract
	//err = json.Unmarshal(msg.Data, &contract)
	//
	//if err != nil {
	//	return Contract{}, err
	//}
	//
	//return contract, nil
}