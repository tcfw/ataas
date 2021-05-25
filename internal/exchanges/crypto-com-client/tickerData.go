package client

type TickerData struct {
	Instrument          string  `json:"i"`
	BestBid             float64 `json:"b"`
	BestAsk             float64 `json:"k"`
	Last                float64 `json:"a"`
	Timestamp           float64 `json:"t"`
	Volume24h           float64 `json:"v"`
	Highest24h          float64 `json:"h"`
	Lowest24h           float64 `json:"l"`
	ClosePriceChange24h float64 `json:"c"`
	Open                float64 `json:"o"`
}
