package paypal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	clientId     string
	clientSecret string
	endpoint     string

	accessToken          string
	accessTokenExpiresAt int64
}

const (
	ApiSandbox = "https://api-m.sandbox.paypal.com"
)

func NewClient(clientId, clientSecret, endpoint string) Client {
	return Client{
		clientId:     clientId,
		clientSecret: clientSecret,
		endpoint:     endpoint,
	}
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (pp *Client) getAccessToken() (string, error) {
	now := time.Now().UnixMilli()
	if pp.accessTokenExpiresAt > now {
		return pp.accessToken, nil
	}

	payloadBuf := new(bytes.Buffer)
	payloadBuf.Write([]byte("grant_type=client_credentials"))
	req, _ := http.NewRequest("POST", pp.endpoint+"/v1/oauth2/token", payloadBuf)
	req.SetBasicAuth(pp.clientId, pp.clientSecret)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var resp accessTokenResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", err
	}

	pp.accessToken = resp.AccessToken
	pp.accessTokenExpiresAt = now + resp.ExpiresIn

	return pp.accessToken, nil
}

type purchaseUnitAmount struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type purchaseUnit struct {
	Amount purchaseUnitAmount `json:"amount"`
}

type orderApplicationContext struct {
	ReturnUrl string `json:"return_url"`
}

type createOrderRequest struct {
	Intent             string                  `json:"intent"`
	PurchaseUnits      []purchaseUnit          `json:"purchase_units"`
	ApplicationContext orderApplicationContext `json:"application_context"`
}

type createOrderResponse struct {
	Id string `json:"id"`
}

func (pp *Client) CreateOrder(internalOrderId string, currency string, amount float64) (string, error) {
	accessToken, err := pp.getAccessToken()
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(createOrderRequest{
		Intent: "CAPTURE",
		PurchaseUnits: []purchaseUnit{
			{
				Amount: purchaseUnitAmount{
					CurrencyCode: currency,
					Value:        fmt.Sprintf("%.2f", amount),
				},
			},
		},
		ApplicationContext: orderApplicationContext{
			ReturnUrl: "http://127.0.0.1:8081/orders/" + internalOrderId + "/finish-payment",
		},
	})
	if err != nil {
		return "", err
	}

	payloadBuf := new(bytes.Buffer)
	payloadBuf.Write(payload)
	req, _ := http.NewRequest("POST", pp.endpoint+"/v2/checkout/orders", payloadBuf)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var resp createOrderResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", err
	}

	return resp.Id, nil
}

type captureOrderResponse struct {
	Status string `json:"status"`
}

func (pp *Client) CheckOrderCompleted(orderId string) (bool, error) {
	accessToken, err := pp.getAccessToken()
	if err != nil {
		return false, err
	}

	payloadBuf := new(bytes.Buffer)
	payloadBuf.Write([]byte("{}"))
	req, _ := http.NewRequest("POST", pp.endpoint+"/v2/checkout/orders/"+orderId+"/capture", payloadBuf)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	var resp captureOrderResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return false, err
	}

	return resp.Status == "COMPLETED", nil
}
