package strategies

import (
	"context"
	"math"
	"sort"
	"time"

	"pm.tcfw.com.au/source/trader/api/pb/strategy"
	"pm.tcfw.com.au/source/trader/api/pb/ticks"
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
	n := 0.0

	for i := 1; i < len(trades); i++ {
		a := trades[i].Amount
		b := trades[i-1].Amount
		if a == b {
			continue
		}

		n++

		sum += math.Log(float64(a / b))
		// s := math.Log(float64(a / b))
		// if s < 0 {
		// 	sum += -1
		// } else {
		// 	sum += 1
		// }
	}

	// avg := sum / n

	if sum == 0 {
		return strategy.Action_STAY
	} else if sum < 0 {
		return strategy.Action_SELL
	}

	return strategy.Action_BUY
}
