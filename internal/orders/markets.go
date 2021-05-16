package orders

import (
	"context"

	"pm.tcfw.com.au/source/trader/internal/exchanges"
	binance "pm.tcfw.com.au/source/trader/internal/exchanges/binance-client"
	cryptoCom "pm.tcfw.com.au/source/trader/internal/exchanges/crypto-com-client"
)

type MarketList map[string]exchanges.Exchange

const (
	cryptocomKey    = "w3UzTQv1nhXCtXtSHTm1FW"
	cryptocomSecret = "UVD8aG5L7FX3asR6qeQD7p"

	binanceKey    = "6eYb30dkibfkdNI7FasPBOWIU85GywLktejWaY4PtkefS4KFiGwilbUNygMoP3wp"
	binanceSecret = "QJPQEhHO6Cfz7uLdUv6lcPSoLBTtovuhgEb4Vf4LsuxmyEGcDBWt9swlbVanhsQx"
)

func initForUser(ctx context.Context) MarketList {
	ml := MarketList{}
	ml["crypto.com"] = cryptoCom.NewClientWithEndpoint(cryptocomKey, cryptocomSecret, "https://uat-api.3ona.co/v2/")
	ml["binance.com"] = binance.NewClient(binanceKey, binanceSecret)
	return ml
}
