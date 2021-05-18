package strategies

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
)

func (w *Worker) handleMeanLog(job *strategy.Strategy) error {
	t, err := ticksSvc()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var dur string
	var ok bool
	dur, ok = job.Params["duration"]
	if ok {
		dur = "5m"
	}

	// tradesResp, err := t.TradesRangeStream(ctx, &ticks.RangeRequest{Market: job.Market, Instrument: job.Instrument, Since: dur})
	// if err != nil {
	// 	return err
	// }

	// trades := make([]*ticks.Trade, 0, 1000)

	// for {
	// 	trade, err := tradesResp.Recv()
	// 	if err == io.EOF {
	// 		break
	// 	} else if err != nil {
	// 		return err
	// 	}

	// 	trades = append(trades, trade)
	// }

	tradesResp, err := t.TradesRange(ctx, &ticks.RangeRequest{Market: job.Market, Instrument: job.Instrument, Since: dur})
	if err != nil {
		return err
	}

	action := meanLog(tradesResp.Data)

	err = w.storeSuggestedAction(action, job)
	if err != nil {
		return err
	}

	return w.broadcastSuggestedAction(action, job)
}

type SortableTrades []*ticks.Trade

func (a SortableTrades) Len() int           { return len(a) }
func (a SortableTrades) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortableTrades) Less(i, j int) bool { return a[i].Timestamp < a[j].Timestamp }

func meanLog(trades []*ticks.Trade) strategy.Action {
	if len(trades) < 2 {
		return strategy.Action_STAY
	}

	//Ensure is sorted in ascending timestamp
	sort.Sort(SortableTrades(trades))

	sum := 0.0

	for i := 1; i < len(trades); i++ {
		a := float64(trades[i].Amount)
		b := float64(trades[i-1].Amount)
		if a == b {
			continue
		}

		sum += math.Log(a / b)
	}

	fmt.Printf("ML: %+v\n", sum)

	if sum > 0.004 {
		return strategy.Action_BUY
	} else if sum > 0.001 {
		return strategy.Action_STAY
	}

	return strategy.Action_SELL
}
