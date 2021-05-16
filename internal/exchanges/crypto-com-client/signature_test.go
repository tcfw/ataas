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
		"notional":        12345.23,
	}

	hmac := c.sign("private/create-order", params, 5577006791947779410, 1620090103445)

	assert.Equal(t, "887ef25120ceef609ba05da4109c288d234b5839843fdd9404a2ad151727039c", hmac)
}
