package orders

import (
	"context"

	"pm.tcfw.com.au/source/ataas/api/pb/excreds"
	"pm.tcfw.com.au/source/ataas/internal/exchanges"
	binance "pm.tcfw.com.au/source/ataas/internal/exchanges/binance-client"
)

type MarketList map[string]exchanges.Exchange

// const (
// 	binanceKey    = "6eYb30dkibfkdNI7FasPBOWIU85GywLktejWaY4PtkefS4KFiGwilbUNygMoP3wp"
// 	binanceSecret = "QJPQEhHO6Cfz7uLdUv6lcPSoLBTtovuhgEb4Vf4LsuxmyEGcDBWt9swlbVanhsQx"
// )

func initForUser(ctx context.Context, account string) (MarketList, error) {
	ml := MarketList{}

	// ml["crypto.com"] = cryptoCom.NewClientWithEndpoint(viper.GetString(""), cryptocomSecret, "https://uat-api.3ona.co/v2/")

	creds, err := getCreds(ctx, account, "binance.com")
	if err != nil {
		return nil, err
	}

	ml["binance.com"] = binance.NewClient(creds.Key, creds.Secret)
	return ml, nil
}

func getCreds(ctx context.Context, account, exchange string) (*excreds.ExchangeCreds, error) {
	creds, err := excredsSvc()
	if err != nil {
		return nil, err
	}

	exCreds, err := creds.Get(ctx, &excreds.GetRequest{Account: account, Exchange: exchange, Decrypt: true})
	if err != nil {
		return nil, err
	}

	return exCreds, nil
}
