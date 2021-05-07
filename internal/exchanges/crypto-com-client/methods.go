package client

import "net/http"

type apiMethod string

const (
	getTicker       apiMethod = "public/get-ticker"
	getTrades       apiMethod = "public/get-trades"
	createOrder     apiMethod = "private/create-order"
	getOrderDetails apiMethod = "private/get-order-details"
)

var (
	methodToHttpMethod = map[apiMethod]string{
		getTicker:       http.MethodGet,
		getTrades:       http.MethodGet,
		createOrder:     http.MethodPost,
		getOrderDetails: http.MethodGet,
	}
)

var (
	requiresSigning = map[apiMethod]bool{
		getTicker:       false,
		getTrades:       false,
		createOrder:     true,
		getOrderDetails: true,
	}
)
