package binance

var (
	subSymbols = []string{
		"adaaud@trade",
		"bnbaud@trade",
		"btcaud@trade",
		"dogeaud@trade",
		"ethaud@trade",
		"linkaud@trade",
		"sxpaud@trade",
		"trxaud@trade",
		"xrpaud@trade",
	}
)

var (
	stepScale = map[string]int{
		"adaaud":  1,
		"bnbaud":  3,
		"btcaud":  5,
		"dogeaud": 0,
		"ethaud":  4,
		"linkaud": 2,
		"sxpaud":  1,
		"trxaud":  0,
		"xrpaud":  0,
	}
)
