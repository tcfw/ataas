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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dur string
	var ok bool
	dur, ok = job.Params["duration"]
	if ok {
		dur = "5m"
	}

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

		s := math.Log(float64(a / b))
		if s < 0 {
			sum += -1
		} else {
			sum += 1
		}
	}

	avg := sum / n

	if avg == 0 {
		return strategy.Action_STAY
	} else if avg < 0 {
		return strategy.Action_SELL
	}

	return strategy.Action_BUY
}
