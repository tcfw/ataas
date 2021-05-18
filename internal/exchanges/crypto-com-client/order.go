package client

import (
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"pm.tcfw.com.au/source/ataas/internal/exchanges"
)

type OrderType string

const (
	OrderTypeLimit             OrderType = "LIMIT"
	OrderTypeMarket            OrderType = "MARKET"
	OrderTypeStop_loss         OrderType = "STOP_LOSS"
	OrderTypeStop_limit        OrderType = "STOP_LIMIT"
	OrderTypeTake_profit       OrderType = "TAKE_PROFIT"
	OrderTypeTake_profit_limit OrderType = "TAKE_PROFIT_LIMIT"
)

type OrderResponse struct {
	price float32
	units float64
}

func (or *OrderResponse) Price() float32 { return or.price }
func (or *OrderResponse) Units() float64 { return or.units }

func (c *Client) Buy(instrument string, price float32, units float64) (exchanges.OrderResponse, error) {
	return c.createImmediateOrder(instrument, true, OrderTypeMarket, price, units)
}

func (c *Client) Sell(instrument string, price float32, units float64) (exchanges.OrderResponse, error) {
	return c.createImmediateOrder(instrument, false, OrderTypeMarket, 0, units)
}

func (c *Client) getOrderHistory() (*CryptoComResponse, error) {
	return c.doReq(getOrderHistory, map[string]interface{}{})
}

func (c *Client) getTrades() (*CryptoComResponse, error) {
	return c.doReq(getUserTrades, map[string]interface{}{})
}

//createImmediateOrder creates a new signed order
//side: true for buy, false for sell
func (c *Client) createImmediateOrder(instrument string, side bool, orderType OrderType, price float32, quantity float64) (exchanges.OrderResponse, error) {
	sideStr := "SELL"
	if side {
		sideStr = "BUY"
	}

	params := map[string]interface{}{
		"instrument_name": instrument,
		"side":            sideStr,
		"type":            orderType,
	}

	if side { //buy
		switch orderType {
		case OrderTypeMarket:
			if price < -1 {
				return nil, status.Error(codes.FailedPrecondition, "price must be set")
			}
			params["notional"] = price
		}
	} else { //sell
		switch orderType {
		case OrderTypeMarket:
			if quantity < 0 {
				return nil, status.Error(codes.FailedPrecondition, "quantity must be set")
			}
			params["quantity"] = quantity
		}
	}

	resp, err := c.doReq(createOrder, params)
	if err != nil {
		return nil, err
	}

	conf := &OrderConfirmation{}
	if err := json.Unmarshal(resp.Result, conf); err != nil {
		return nil, err
	}

	order, err := c.orderDetails(conf.OrderID)
	if err != nil {
		return nil, err
	}

	fnResp := &OrderResponse{
		price: order.TradeList[0].TradedPrice,
		units: float64(order.TradeList[0].TradedQuantity),
	}

	return fnResp, nil
}

func (c *Client) orderDetails(orderID string) (*OrderDetails, error) {
	res, err := c.doReq(getOrderDetails, map[string]interface{}{
		"order_id": orderID,
	})
	if err != nil {
		return nil, err
	}

	details := &OrderDetails{}
	if err := json.Unmarshal(res.Result, details); err != nil {
		return nil, err
	}

	return details, nil
}
