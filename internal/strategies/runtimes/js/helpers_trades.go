package js

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dop251/goja"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	ticksAPI "pm.tcfw.com.au/source/ataas/api/pb/ticks"
	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

type LimitedGetTrades struct {
	Until time.Time
}

func (lgt *LimitedGetTrades) GetTrades(exchange, symbol, duration string) []*ticksAPI.Trade {
	svc, err := tradesClient()
	if err != nil {
		panic(err)
	}

	fmt.Printf("TSF: %+v %+v", duration, lgt.Until.Format(time.RFC3339))

	trades, err := svc.TradesRange(context.Background(), &ticksAPI.RangeRequest{
		Market:     exchange,
		Instrument: symbol,
		Since:      duration,
		Until:      lgt.Until.Format(time.RFC3339),
	})
	if err != nil {
		panic(err)
	}

	return trades.Data
}

func GetTrades(exchange, symbol, duration string) []*ticksAPI.Trade {
	svc, err := tradesClient()
	if err != nil {
		panic(err)
	}

	trades, err := svc.TradesRange(context.Background(), &ticksAPI.RangeRequest{
		Market:     exchange,
		Instrument: symbol,
		Since:      duration,
	})
	if err != nil {
		panic(err)
	}

	return trades.Data
}

func tradesClient() (ticksAPI.HistoryServiceClient, error) {
	ticksEndpoint, envExists := os.LookupEnv("TICKS_HOST")
	if !envExists {
		ticksEndpoint = viper.GetString("grpc.addr")
	}

	conn, err := grpc.Dial(ticksEndpoint, rpcUtils.InternalClientOptions()...)
	if err != nil {
		return nil, err
	}

	ticksSvc := ticksAPI.NewHistoryServiceClient(conn)

	return ticksSvc, nil
}

func GetTestTrades(call goja.FunctionCall) goja.Value {
	v := goja.New().ToValue([]*ticksAPI.Trade{
		{TradeID: "1", Amount: 1, Units: 1, Timestamp: 1},
		{TradeID: "2", Amount: 2, Units: 2, Timestamp: 2},
		{TradeID: "3", Amount: 3, Units: 3, Timestamp: 3},
		{TradeID: "4", Amount: 4, Units: 4, Timestamp: 4},
		{TradeID: "5", Amount: 5, Units: 5, Timestamp: 5},
	})

	return v
}
