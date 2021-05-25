package client

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
	"pm.tcfw.com.au/source/ataas/internal/exchanges"
)

type Client struct {
	c  *http.Client
	ws *wsManager

	key          string
	secret       string
	httpEndpoint string
}

var _ exchanges.Exchange = (*Client)(nil)

func NewClient(key, secret string) *Client {
	return NewClientWithEndpoint(key, secret, restEndpoint)
}

func NewClientWithEndpoint(key, secret, endpoint string) *Client {
	return &Client{
		c:            newHttpClent(),
		ws:           newWsManager(),
		key:          key,
		secret:       secret,
		httpEndpoint: endpoint,
	}
}

func (c *Client) GetTicker(instrument string) (*TickerData, error) {
	params := map[string]interface{}{
		"instrument_name": instrument,
	}

	resp, err := c.doReq(getTicker, params)
	if err != nil {
		return nil, err
	}

	tickerData := &SingleTick{}
	if err := json.Unmarshal(resp.Result, tickerData); err != nil {
		return nil, err
	}

	return tickerData.Data, nil
}

func (c *Client) SubscribeTradesAll() (<-chan *ticks.Trade, error) {
	return c.ws.SubscribeTradesAll()
}

func (c *Client) SubscribeTicker(instrument string, getHistory bool) (<-chan *TickerSubscriptionEvent, error) {
	return c.ws.SubscribeTicker(instrument, getHistory)
}

func (c *Client) SubscribeTickerAll() (<-chan *TickerSubscriptionEvent, error) {
	return c.ws.SubscribeTickerAll()
}

func (c *Client) GetTickerAll() (*TickerResponse, error) {
	resp, err := c.doReq(getTicker, nil)
	if err != nil {
		return nil, err
	}

	tickerData := &TickerResponse{}
	if err := json.Unmarshal(resp.Result, tickerData); err != nil {
		return nil, err
	}

	return tickerData, nil
}

func (c *Client) doReq(method apiMethod, params map[string]interface{}) (*CryptoComResponse, error) {
	req, err := c.newHttpRequest(method, params)
	if err != nil {
		return nil, err
	}

	logrus.New().Infof("CRYPTO.com REQ: %+v\n", req)

	httpResp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	r := io.LimitReader(httpResp.Body, 50<<20) //limit response to 50MB

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	logrus.New().Infof("CRYPTO.com: %s\n", body)

	resp := &CryptoComResponse{}

	if err := json.Unmarshal(body, resp); err != nil {
		return nil, err
	}

	if resp.Code != SUCCESS {
		err := &ResponseError{
			Code:     resp.Code,
			Response: resp,
		}

		return nil, err
	}

	return resp, nil
}
