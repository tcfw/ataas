package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testKey    = "w3UzTQv1nhXCtXtSHTm1FW"
	testSecret = "UVD8aG5L7FX3asR6qeQD7p"
)

func TestCreateOrder(t *testing.T) {
	c := NewClientWithEndpoint(testKey, testSecret, "https://uat-api.3ona.co/v2/")

	res, err := c.createImmediateOrder("BTC_USDT", true, OrderTypeMarket, 12345.23, 1)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, res.Price())
		assert.Equal(t, 1, res.Price())
	}
}
