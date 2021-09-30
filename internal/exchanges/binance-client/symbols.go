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
		"adaaud":  3,
		"bnbaud":  3,
		"btcaud":  6,
		"dogeaud": 1,
		"ethaud":  5,
		"linkaud": 3,
		"sxpaud":  3,
		"trxaud":  0,
		"xrpaud":  2,
	}
)
