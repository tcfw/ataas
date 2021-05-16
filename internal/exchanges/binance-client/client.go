package binance

import (
	"net/http"

	"pm.tcfw.com.au/source/trader/api/pb/ticks"
)

const (
	defaultRestEndpoint = "https://api.binance.com"
)

type Client struct {
	c  *http.Client
	ws *wsManager

	key          string
	secret       string
	httpEndpoint string
}

func NewClient(key, secret string) *Client {
	return NewClientWithEndpoint(key, secret, defaultRestEndpoint)
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

func newHttpClent() *http.Client {
	return &http.Client{}
}

func (c *Client) SubscribeTradesAll() (<-chan *ticks.Trade, error) {
	return c.ws.SubscribeTradesAll()
}

