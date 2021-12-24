package http

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

type AccessTokenResponse struct {
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

func GetAccessToken(host string, user string, pass string,
					appName string, appVersion string, clientId string,
					deviceId string, apiKey string, penaltyTicket string) (AccessTokenResponse, error) {
	var accessTokenResponse AccessTokenResponse

	//if c.AccessToken != "" && c.TokenExpiration.After(time.Now()) {
	//	return nil
	//} else if c.AccessToken != "" && c.TokenExpiration.Before(time.Now()) {
	//	// TODO: Refresh Token
	//}

	reqBody := accessTokenRequest{
		Name:       user,
		Password:   pass,
		AppId:      appName,
		AppVersion: appVersion,
		ClientId:   clientId,
		DeviceId:   deviceId,
		ApiKey:     apiKey,
	}

	if penaltyTicket != "" {
		reqBody.PenaltyTicket = penaltyTicket
	}

	b, err := json.Marshal(reqBody)

	if err != nil {
		return accessTokenResponse, err
	}

	url := host + GET_ACCESS_TOKEN

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return accessTokenResponse, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var parsedResponse map[string]interface{}
	err = json.Unmarshal(respBody, &parsedResponse)
	if err != nil {
		return accessTokenResponse, err
	}

	if _, ok := parsedResponse["p-ticket"]; ok {
		var rateLimitResponse map[string]interface{}
		err = json.Unmarshal(respBody, &rateLimitResponse)

		return accessTokenResponse, &RateLimitError{
			PenaltyTicket:     parsedResponse["p-ticket"].(string),
			PenaltyTime:       int64(parsedResponse["p-time"].(float64)),
			PenaltyExpiration: time.Unix(int64(parsedResponse["p-time"].(float64)), 0),
			PenaltyCaptcha:    parsedResponse["p-captcha"].(bool),
		}
	}

	err = json.Unmarshal(respBody, &accessTokenResponse)

	//c.TokenExpiration, err = time.Parse("2006-01-02T15:04:05Z", parsedResponse["expirationTime"].(string))
	//
	//if err != nil {
	//	return accessTokenResponse, err
	//}
	//c.AccessToken = parsedResponse["accessToken"].(string)

	return accessTokenResponse, nil
}