package strategies

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	blocksAPI "pm.tcfw.com.au/source/ataas/api/pb/blocks"
	ordersAPI "pm.tcfw.com.au/source/ataas/api/pb/orders"
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
)

func (s *Server) BackTest(ctx context.Context, req *strategy.BacktestRequest) (*strategy.BacktestResponse, error) {
	t, err := ticksSvc()
	if err != nil {
		return nil, err
	}

	b, err := blocksSvc()
	if err != nil {
		return nil, err
	}

	var tsFrom time.Time

	if strings.ContainsAny(req.FromTimestamp, ":/.+") {
		t, err := time.Parse(time.RFC3339, req.FromTimestamp)
		if err != nil {
			return nil, err
		}
		tsFrom = t
	} else {
		ts, err := time.ParseDuration(req.FromTimestamp)
		if err != nil {
			return nil, err
		}

		tsFrom = time.Now().Add(-ts)
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	nextLook := tsFrom.Add(time.Duration(req.Strategy.Duration))

	block := &blocksAPI.Block{
		Purchase:  req.Amount,
		BaseUnits: 1,
	}

	orders := []*ordersAPI.Order{}

	curBlock := []*ticks.Trade{}

	var calc func([]*ticks.Trade, map[string]string) strategy.Action

	switch req.Strategy.Strategy {
	case strategy.StrategyAlgo_MeanLog:
		calc = meanLog
	default:
		return nil, fmt.Errorf("unknown strategy")
	}

	dur, ok := req.Strategy.Params["duration"]
	if !ok {
		dur = "5m"
	}

	duration, err := time.ParseDuration(dur)
	if err != nil {
		return nil, err
	}

	tradesResp, err := t.TradesRangeStream(ctx, &ticks.RangeRequest{Market: req.Strategy.Market, Instrument: req.Strategy.Instrument, Since: req.FromTimestamp})
	if err != nil {
		return nil, err
	}

	var marketPrice float32

	for {
		trade, err := tradesResp.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		tradeTs := trade.Timestamp
		if tradeTs > 9999999999 {
			tradeTs = tradeTs / 1000
		}
		ts := time.Unix(tradeTs, 0)

		curBlock = append(curBlock, trade)

		if len(curBlock) > 10000 {
			curBlock = curBlock[1:]
		}

		if ts.After(nextLook) {
			//execute order

			iTh := len(curBlock) - 1
			afterTs := nextLook.Add(-duration)

			for iTh > 0 {
				tradeTs := curBlock[iTh].Timestamp
				if tradeTs > 9999999999 {
					tradeTs = tradeTs / 1000
				}

				if time.Unix(tradeTs, 0).Before(afterTs) {
					break
				}

				iTh--
			}

			marketPrice = curBlock[len(curBlock)-1].Amount

			sugg := calc(curBlock[iTh:], req.Strategy.Params)

			resp, err := b.CalcState(ctx, &blocksAPI.CalcRequest{Block: block, Action: sugg})
			if err != nil {
				return nil, err
			}

			if resp.State != block.State {
				if resp.State == blocksAPI.BlockState_PURCHASED {
					block.BaseUnits = float64(req.Amount / marketPrice)
				}
				order, err := s.backtestChange(ctx, block, resp.State, marketPrice, nextLook)
				if err != nil {
					return nil, err
				}
				if order != nil {
					orders = append(orders, order)
				}
			}

			nextLook = nextLook.Add(time.Duration(req.Strategy.Duration))
		}
	}

	if len(orders) > 0 && orders[len(orders)-1].Action == 0 {
		orders = append(orders, &ordersAPI.Order{
			Action:    ordersAPI.Action_SELL,
			Price:     marketPrice,
			Units:     block.BaseUnits,
			Timestamp: nextLook.Format(time.RFC3339),
		})
	}

	var purOrder *ordersAPI.Order
	var pnl float32
	var fees float32

	for _, order := range orders {
		fees += order.Price * float32(order.Units) * 0.001
		if order.Action == ordersAPI.Action_BUY {
			purOrder = order
		} else {
			pnl += (purOrder.Price - order.Price) * float32(order.Units)
		}
	}

	resp := &strategy.BacktestResponse{
		Pnl:  pnl,
		Fees: fees,
	}

	if req.ShowOrders {
		resp.Orders = orders
	}

	return resp, nil
}

func (s *Server) backtestChange(ctx context.Context, block *blocksAPI.Block, ns blocksAPI.BlockState, marketPrice float32, ts time.Time) (*ordersAPI.Order, error) {
	if block.State != ns {
		fmt.Printf("BTO: %+v %+v %+v", ns, marketPrice, ts.Format(time.RFC3339))
		if math.IsNaN(float64(marketPrice)) {
			return nil, nil
		}
		block.State = ns
		switch ns {
		case blocksAPI.BlockState_PURCHASED:
			return &ordersAPI.Order{
				Action:    ordersAPI.Action_BUY,
				Price:     marketPrice,
				Units:     block.BaseUnits,
				Timestamp: ts.Format(time.RFC3339),
			}, nil
		case blocksAPI.BlockState_SOLD:
			return &ordersAPI.Order{
				Action:    ordersAPI.Action_SELL,
				Price:     marketPrice,
				Units:     block.BaseUnits,
				Timestamp: ts.Format(time.RFC3339),
			}, nil
		}
	}

	return nil, nil
}
