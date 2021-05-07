package exchanges

type Exchange interface {
	Buy(instrument string, price float32, units float32) (OrderResponse, error)
	Sell(instrument string, price float32, units float32) (OrderResponse, error)
}

type OrderResponse interface {
	Price() float32
	Units() float32
}
