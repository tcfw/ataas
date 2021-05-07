package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignature(t *testing.T) {
	c := NewClient("API_KEY", "SECRET_KEY")

	params := map[string]interface{}{
		"instrument_name": "BTC_USDT",
		"side":            "BUY",
		"type":            "MARKET",
		"time_in_force":   "IMMEDIATE_OR_CANCEL",
		"client_oid":      "Trader",
		"notional":        12345.23,
	}

	hmac := c.sign("private/get-order-detail", params, 5577006791947779410, 1620090103445)

	assert.Equal(t, "95453ee99107c53e142a546f0a275a40209eaaefbccc1e875fd7bee0cb3ecd90", hmac)
}
