package binance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gogo/status"
	"google.golang.org/grpc/codes"
	"pm.tcfw.com.au/source/ataas/internal/exchanges"
)

const (
	txFee = 0.001
)

type OrderResponse struct {
	price float32
	units float64
}

func (or *OrderResponse) Price() float32 { return or.price }
func (or *OrderResponse) Units() float64 { return or.units }

type OrderType string

const (
	OrderTypeLimit             OrderType = "LIMIT"
	OrderTypeMarket            OrderType = "MARKET"
	OrderTypeStop_loss         OrderType = "STOP_LOSS"
	OrderTypeStop_limit        OrderType = "STOP_LIMIT"
	OrderTypeTake_profit       OrderType = "TAKE_PROFIT"
	OrderTypeTake_profit_limit OrderType = "TAKE_PROFIT_LIMIT"
	OrderTypeLimitMaker        OrderType = "LIMIT_MAKER"
)

func (c *Client) Buy(instrument string, price float32, units float64) (exchanges.OrderResponse, error) {
	return c.createOrder(instrument, true, OrderTypeMarket, price, units)
}

func (c *Client) Sell(instrument string, price float32, units float64) (exchanges.OrderResponse, error) {
	return c.createOrder(instrument, false, OrderTypeMarket, price, units)

}

type OrderResp struct {
	Symbol              string `json:"symbol"`              // "BTCUSDT",
	OrderId             int64  `json:"orderId"`             // 28,
	OrderListId         int    `json:"orderListId"`         // -1, //Unless OCO, value will be -1
	ClientOrderId       string `json:"clientOrderId"`       // "6gCrw2kRUAF9CvJDGP16IP",
	TransactTime        int64  `json:"transactTime"`        // 1507725176595,
	Price               string `json:"price"`               // "0.00000000",
	OrigQty             string `json:"origQty"`             // "10.00000000",
	ExecutedQty         string `json:"executedQty"`         // "10.00000000",
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"` // "10.00000000",
	Status              string `json:"status"`              // "FILLED",
	TimeInForce         string `json:"timeInForce"`         // "GTC",
	Type                string `json:"type"`                // "MARKET",
	Side                string `json:"side"`                // "SELL"
}

type ErrResp struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (e ErrResp) Error() string {
	return fmt.Sprintf("code=%d: %s", e.Code, e.Message)
}

func (c *Client) createOrder(symbol string, side bool, orderType OrderType, price float32, quantity float64) (exchanges.OrderResponse, error) {

	fmt.Printf("BINANCE ORDER: SYM:%s SIDE:%t TYPE:%s PRICE:%+v QUANT:%+v\n", symbol, side, orderType, price, quantity)

	vals := url.Values{
		"symbol":           {symbol},
		"side":             {"SELL"},
		"type":             {string(orderType)},
		"timestamp":        {strconv.FormatInt(time.Now().UnixNano()/1000000, 10)},
		"newOrderRespType": {"RESULT"},
	}

	if side { //buy
		switch orderType {
		case OrderTypeMarket:
			if price < -1 {
				return nil, status.Error(codes.FailedPrecondition, "price must be set")
			}
			vals["quoteOrderQty"] = []string{strconv.FormatFloat(float64(price*(1+txFee)), 'f', -1, 32)}
			// vals["quantity"] = []string{strconv.FormatFloat(float64(quantity), 'f', -1, 32)}
			vals["side"] = []string{"BUY"}
		}
	} else { //sell
		switch orderType {
		case OrderTypeMarket:
			if quantity <= 0 {
				return nil, status.Error(codes.FailedPrecondition, "quantity must be set")
			}
			vals["quantity"] = []string{strconv.FormatFloat(float64(quantity), 'f', -1, 32)}
		case OrderTypeLimit:
			if quantity <= 0 {
				return nil, status.Error(codes.FailedPrecondition, "quantity must be set")
			}
			if price <= 0 {
				return nil, status.Error(codes.FailedPrecondition, "quantity must be set")
			}
			vals["quantity"] = []string{strconv.FormatFloat(float64(quantity), 'f', -1, 32)}
			vals["price"] = []string{strconv.FormatFloat(float64(price), 'f', -1, 32)}
			vals["timeInForce"] = []string{"IOC"}
		}
	}

	pl := c.sign(vals, []byte(c.secret))

	req, err := http.NewRequest(http.MethodPost, c.httpEndpoint+"/api/v3/order", bytes.NewReader([]byte(pl)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-MBX-APIKEY", c.key)

	rResp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)

	r := io.TeeReader(rResp.Body, buf)

	bResp := &OrderResp{}
	err = json.NewDecoder(io.LimitReader(r, 10<<20)).Decode(bResp)
	if err != nil {
		return nil, err
	}

	fmt.Printf("RESP: %+v :: %s\n", bResp, buf.String())

	if bResp.Symbol == "" {
		//Assume error or empty resp
		err = &ErrResp{}
		json.Unmarshal(buf.Bytes(), err)
		return nil, err
	}

	respPrice, err := strconv.ParseFloat(bResp.Price, 32)
	if err != nil {
		return nil, err
	}

	respQuantity, err := strconv.ParseFloat(bResp.ExecutedQty, 32)
	if err != nil {
		return nil, err
	}

	if side { //buy
		respQuantity = respQuantity * (1 - txFee)
	} else {
		respQuantity = respQuantity * (1 + txFee)
	}

	respQStepScale, ok := stepScale[bResp.Symbol]
	if !ok {
		respQStepScale = 2
	}

	res := &OrderResponse{
		price: float32(respPrice),
		units: truncatePrecision(respQuantity, respQStepScale),
	}

	return res, nil
}

func truncatePrecision(f float64, pres int) float64 {
	i := math.Pow10(pres)
	return float64(int(f*i)) / i
}
