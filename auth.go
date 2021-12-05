package tradovate

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type accessTokenRequest struct {
	Name string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
	AppId string `json:"appId,omitempty"`
	AppVersion string `json:"appVersion,omitempty"`
	ClientId string `json:"cid,omitempty"`
	DeviceId string `json:"deviceId,omitempty"`
	ApiKey string `json:"sec,omitempty"`

	PenaltyTicket string `json:"p-ticket,omitempty"`
}

type accessTokenResponse struct {
	AccessToken string `json:"accessToken"`
	MarketDataAccessToken string `json:"mdAccessToken"`
	ExpirationTime time.Time `json:"expirationTime"`
	UserStatus string `json:"userStatus"`
	UserId int `json:"userId"`
	Name string `json:"name"`
	HasLive bool `json:"hasLive"`
	OutdatedTaC bool `json:"outdatedTaC"`
	HasFunded bool `json:"hasFunded"`
	HasMarketData bool `json:"hasMarketData"`
	OutdatedLiquidationPolicy bool `json:"outdatedLiquidationPolicy"`
}

type AuthorizationResponse struct {
	Id int64 `json:"i"`
	StatusCode int `json:"s"`
}

func (c *Client) GetAccessToken() error {
	if c.AccessToken != "" && c.TokenExpiration.After(time.Now()) {
		return nil
	} else if c.AccessToken != "" && c.TokenExpiration.Before(time.Now()) {
		// TODO: Refresh Token
	}


	reqBody := accessTokenRequest{
		Name: 		c.Username,
		Password: 	c.Password,
		AppId: 		c.AppName,
		AppVersion: c.AppVersion,
		ClientId: 	c.ClientId,
		DeviceId: 	c.DeviceId,
		ApiKey: 	c.ApiKey,
	}

	if c.PenaltyTicket == "" {
		reqBody.PenaltyTicket = c.PenaltyTicket
	}

	b, err := json.Marshal(reqBody)

	if err != nil {
		return err
	}

	url := c.HttpHost + GET_ACCESS_TOKEN

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var parsedResponse map[string]interface{}
	err = json.Unmarshal(respBody, &parsedResponse)
	if err != nil {
		return err
	}

	if _, ok := parsedResponse["p-ticket"]; ok {
		var rateLimitResponse map[string]interface{}
		err = json.Unmarshal(respBody, &rateLimitResponse)

		return &RateLimitError{
			PenaltyTicket: parsedResponse["p-ticket"].(string),
			PenaltyTime: int64(parsedResponse["p-time"].(float64)),
			PenaltyExpiration: time.Unix(int64(parsedResponse["p-time"].(float64)), 0),
			PenaltyCaptcha: parsedResponse["p-captcha"].(bool),
		}
	}

	c.TokenExpiration, err = time.Parse("2006-01-02T15:04:05Z", parsedResponse["expirationTime"].(string))

	if err != nil {
		return err
	}
	c.AccessToken = parsedResponse["accessToken"].(string)

	return nil
}