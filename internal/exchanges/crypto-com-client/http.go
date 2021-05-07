package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

func newHttpClent() *http.Client {
	return &http.Client{}
}

const (
	restEndpoint     = "https://api.crypto.com/v2/"
	wsMarketEndpoint = "wss://stream.crypto.com/v2/market"
)

func (c *Client) newHttpRequest(method apiMethod, params map[string]interface{}) (*http.Request, error) {
	endpoint := fmt.Sprintf("%s%s", c.httpEndpoint, method)

	httpMethod, found := methodToHttpMethod[method]
	if !found {
		return nil, ErrUnknownMethod
	}

	var body string

	if httpMethod == http.MethodGet && len(params) > 0 {
		q := url.Values{}

		for k, v := range params {
			if vString, ok := v.(string); ok {
				q.Add(k, vString)
			}
		}

		endpoint = fmt.Sprintf("%s?%s", endpoint, q.Encode())
	} else if httpMethod == http.MethodPost {
		req := &CryptoComRequest{
			Id:     uint64(rand.Uint64() / 100000),
			Nonce:  uint64(time.Now().UnixNano() / 1000000),
			Method: string(method),
			Params: params,
		}

		sign, exists := requiresSigning[method]
		if exists && sign {
			req.ApiKey = c.key
			req.Signature = c.sign(method, params, req.Id, req.Nonce)
		}

		b, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		body = string(b)
	}

	logrus.New().Infof("CRYPTO.com BODY: %s\n", body)

	req, err := http.NewRequest(httpMethod, endpoint, bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}

	if httpMethod == http.MethodPost {
		req.Header.Add("Content-Type", "application/json")
	}

	return req, nil
}
