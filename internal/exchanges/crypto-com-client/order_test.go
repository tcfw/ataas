package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testKey    = "V1kXAcQrWKqdNeozwfrUgW"
	testSecret = "FnoYVQEGs3fn7pVwXbRckd"
)

func TestCreateOrder(t *testing.T) {
	c := NewClientWithEndpoint(testKey, testSecret, "https://uat-api.3ona.co/v2/")

	res, err := c.createImmediateOrder("ETH_USDT", true, OrderTypeMarket, 10000.01, 1)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, res.Price())
		assert.Equal(t, 1, res.Price())
	}
}

func TestOrderHistory(t *testing.T) {
	c := NewClient(testKey, testSecret)
	r, err := c.getOrderHistory()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, r.Result)
}

func TestTrades(t *testing.T) {
	c := NewClient(testKey, testSecret)
	r, err := c.getTrades()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, r.Result)
}
