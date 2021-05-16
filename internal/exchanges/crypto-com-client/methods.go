package client

import "net/http"

type apiMethod string

const (
	getTicker       apiMethod = "public/get-ticker"
	getTrades       apiMethod = "public/get-trades"
	createOrder     apiMethod = "private/create-order"
	getOrderDetails apiMethod = "private/get-order-details"
	getOrderHistory apiMethod = "private/get-order-history"
	getUserTrades   apiMethod = "private/get-trades"
)

var (
	methodToHttpMethod = map[apiMethod]string{
		getTicker:       http.MethodGet,
		getTrades:       http.MethodGet,
		createOrder:     http.MethodPost,
		getOrderDetails: http.MethodPost,
		getOrderHistory: http.MethodPost,
		getUserTrades:   http.MethodPost,
	}
)

var (
	requiresSigning = map[apiMethod]bool{
		getTicker:       false,
		getTrades:       false,
		createOrder:     true,
		getOrderDetails: true,
		getOrderHistory: true,
		getUserTrades:   true,
	}
)
