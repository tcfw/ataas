package orders

import (
	"context"

	"pm.tcfw.com.au/source/trader/internal/exchanges"
	cryptoCom "pm.tcfw.com.au/source/trader/internal/exchanges/crypto-com-client"
)

type MarketList map[string]exchanges.Exchange

const (
	key    = "w3UzTQv1nhXCtXtSHTm1FW"
	secret = "UVD8aG5L7FX3asR6qeQD7p"
)

func initForUser(ctx context.Context) MarketList {
	ml := MarketList{}
	ml["crypto.com"] = cryptoCom.NewClientWithEndpoint(key, secret, "https://uat-api.3ona.co/v2/")
	return ml
}
