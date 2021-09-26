package exchanges

type Exchange interface {
	Buy(instrument string, price float32, units float64) (OrderResponse, error)
	Sell(instrument string, price float32, units float64) (OrderResponse, error)
}

type OrderResponse interface {
	Price() string
	Units() string
}
