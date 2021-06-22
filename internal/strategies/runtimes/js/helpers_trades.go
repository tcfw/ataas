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

	trades, err := svc.TradesRange(context.TODO(), &ticksAPI.RangeRequest{
		Market:     exchange,
		Instrument: symbol,
		Since:      duration,
	})
	if err != nil {
		panic(err)
	}

	return trades.Data
}

func GetCandles(exchange, symbol, duration string, depth int) []*ticksAPI.OHLCV {
	svc, err := tradesClient()
	if err != nil {
		panic(err)
	}

	trades, err := svc.Candles(context.TODO(), &ticksAPI.CandlesRequest{
		Market:     exchange,
		Instrument: symbol,
		Interval:   duration,
		Depth:      int32(depth),
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

func GetTestCandles(call goja.FunctionCall) goja.Value {
	v := goja.New().ToValue([]*ticksAPI.OHLCV{
		{Open: 445.59, High: 451.02, Low: 445, Close: 448.93, Volume: 230.55168},
		{Open: 449.26, High: 452.94, Low: 448.65, Close: 451.71, Volume: 172.53421},
		{Open: 451.61, High: 453.58, Low: 449.19, Close: 450.26, Volume: 133.16269},
		{Open: 450.1, High: 453.24, Low: 449.69, Close: 449.99, Volume: 68.9112},
		{Open: 450.61, High: 452.51, Low: 446.75, Close: 449.92, Volume: 96.6359},
		{Open: 449.93, High: 451.78, Low: 446.53, Close: 448.88, Volume: 212.23628},
		{Open: 449.32, High: 450.76, Low: 445.26, Close: 446, Volume: 89.62449},
		{Open: 446.8, High: 449.37, Low: 442.8, Close: 449.37, Volume: 305.3545},
		{Open: 448.73, High: 452.39, Low: 448.66, Close: 451.83, Volume: 237.36652},
		{Open: 452.25, High: 453.08, Low: 446.28, Close: 447.04, Volume: 178.77116},
		{Open: 446.4, High: 446.64, Low: 439.07, Close: 441.07, Volume: 723.7745},
		{Open: 440.63, High: 443.35, Low: 439, Close: 441.9, Volume: 319.9392},
		{Open: 441.96, High: 442.15, Low: 439.2, Close: 439.65, Volume: 228.94815},
		{Open: 438.3, High: 441.09, Low: 436.74, Close: 441.09, Volume: 356.13446},
		{Open: 440.67, High: 440.67, Low: 436.24, Close: 436.58, Volume: 112.941864},
		{Open: 436, High: 441.04, Low: 435.27, Close: 439.12, Volume: 224.3942},
		{Open: 438.92, High: 439.23, Low: 436.29, Close: 438.92, Volume: 183.984},
		{Open: 439, High: 442.09, Low: 438.01, Close: 442.04, Volume: 327.68677},
	})

	return v
}
