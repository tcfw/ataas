package client

import (
	"encoding/json"
)

type CryptoComResponse struct {
	Id       uint64          `json:"id"`
	Method   string          `json:"method"`
	Result   json.RawMessage `json:"result"`
	Code     ResponseCode    `json:"code"`
	Message  string          `json:"message,omitempty"`
	Original string          `json:"original,omitempty"`
}

type SubscriptionEvent struct {
	InstrumentName string `json:"instrument_name,omitempty"`
	Channel        string `json:"channel,omitempty"`
	Subscription   string `json:"subscription,omitempty"`
}

type TickerSubscriptionEvent struct {
	SubscriptionEvent
	Data []*TickerData `json:"data"`
}

type TradeSubscriptionEvent struct {
	SubscriptionEvent
	Data []*TradeEvent `json:"data"`
}

type TradeEvent struct {
	ID        int64   `json:"d"`
	Price     float32 `json:"p"`
	Quantity  float32 `json:"q"`
	Side      string  `json:"s"` // ("buy" or "sell")
	Timestamp int64   `json:"t"`
}

type OrderConfirmation struct {
	OrderID   string `json:"order_id"`   //	Newly created order ID
	ClientOID string `json:"client_oid"` //	(Optional) if a Client order ID was provided in the request
}

type OrderDetails struct {
	TradeList []*OrderDetailsTrade `json:"trade_list"`
	Info      *OrderDetailsInfo    `json:"order_info"`
}

type OrderDetailsTrade struct {
	Side           string  `json:"side"`            //	BUY, SELL
	InstrumentName string  `json:"instrument_name"` //	e.g. ETH_CRO, BTC_USDT
	Fee            float32 `json:"fee"`             //	Trade fee
	TradeID        string  `json:"trade_id"`        //	Trade ID
	CreateTime     int64   `json:"create_time"`     //	Trade creation time
	TradedPrice    float32 `json:"traded_price"`    //	Executed trade price
	TradedQuantity float32 `json:"traded_quantity"` //	Executed trade quantity
	FeeCurrency    string  `json:"fee_currency"`    //	Currency used for the fees (e.g. CRO)
	OrderID        string  `json:"order_id"`        //	Order ID
}

type OrderDetailsInfo struct {
	Status             string  `json:"status"`              //ACTIVE, CANCELED, FILLED, REJECTED or EXPIRED
	Reason             string  `json:"reason"`              //Reason code (see "Response and Reason Codes") -- only for REJECTED orders
	Side               string  `json:"side"`                //BUY, SELL
	Price              float32 `json:"price"`               //Price specified in the order
	Quantity           float32 `json:"quantity"`            //Quantity specified in the order
	OrderID            string  `json:"order_id"`            //Order ID
	ClientOID          string  `json:"client_oid"`          //(Optional) Client order ID if included in request
	CreateTime         float32 `json:"create_time"`         //Order creation time (Unix timestamp)
	UpdateTime         float32 `json:"update_time"`         //Order update time (Unix timestamp)
	Type               string  `json:"type"`                //LIMIT, MARKET, STOP_LOSS, STOP_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	InstrumentName     string  `json:"instrument_name"`     //e.g. ETH_CRO, BTC_USDT
	CumulativeQuantity float32 `json:"cumulative_quantity"` //Cumulative executed quantity (for partially filled orders)
	CumulativeValue    float32 `json:"cumulative_value"`    //Cumulative executed value (for partially filled orders)
	AvgPrice           float32 `json:"avg_price"`           //Average filled price. If none is filled, returns 0
	FeeCurrency        string  `json:"fee_currency"`        //Currency used for the fees (e.g. CRO)
	TimeInForce        string  `json:"time_in_force"`       //- GOOD_TILL_CANCEL, FILL_OR_KILL, IMMEDIATE_OR_CANCEL
	ExecInst           string  `json:"exec_inst"`           //Empty or POST_ONLY (Limit Orders Only)
	TriggerPrice       float32 `json:"trigger_price"`       //Used for trigger-related orders
}

type TickerResponse struct {
	Data []*TickerData `json:"data"`
}

type SingleTick struct {
	InstrumentName string      `json:"instrument_name,omitempty"`
	Data           *TickerData `json:"data"`
}

func (tr *TickerResponse) Instruments() []string {
	instruments := make([]string, len(tr.Data))

	for i, tick := range tr.Data {
		instruments[i] = tick.Instrument
	}

	return instruments
}

func (tr *TickerResponse) Instrument(in string) *TickerData {
	for _, tick := range tr.Data {
		if tick.Instrument == in {
			return tick
		}
	}

	return nil
}
