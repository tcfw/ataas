package strategy

import (
	"math"
	"sync"
	"time"

	"pm.tcfw.com.au/source/trader/common"
	client "pm.tcfw.com.au/source/trader/internal/ticks/crypto-com-client"
)

type SlidingWindow struct {
	maxN    int
	maxDiff time.Duration

	ticks []*common.TickerData

	minTs time.Time
	maxTs time.Time

	mu sync.RWMutex

	onTick func()
}

var _ common.Strategy = &SlidingWindow{}

func NewSWMaxN(n int) *SlidingWindow {
	return &SlidingWindow{
		maxN:  n,
		ticks: []*common.TickerData{},
	}
}

func NewSWDiff(t time.Duration) *SlidingWindow {
	return &SlidingWindow{
		maxN:    0,
		maxDiff: t,
		ticks:   []*common.TickerData{},
	}
}

func (sw *SlidingWindow) CurrentAsk() float64 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	return sw.ticks[len(sw.ticks)-1].BestAsk
}

func (sw *SlidingWindow) Pipe(inCh <-chan *client.TickerSubscriptionEvent) <-chan *client.TickerSubscriptionEvent {
	outCh := make(chan *client.TickerSubscriptionEvent)

	go func() {
		defer close(outCh)

		for tick := range inCh {
			for _, tickD := range tick.Data {
				sw.Injest(tickD)
			}

			outCh <- tick
		}
	}()

	return outCh
}

func (sw *SlidingWindow) EndPipe(inCh <-chan *client.TickerSubscriptionEvent) {
	for tick := range inCh {
		for _, tickD := range tick.Data {
			sw.Injest(tickD)
		}
	}
}

func (sw *SlidingWindow) Injest(t *common.TickerData) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	tip := 0
	if sw.atMax(t) {
		tip = 1
	}

	sw.ticks = append(sw.ticks[tip:], t)

	sw.minTs = time.Unix(int64(sw.ticks[0].Timestamp/1000), 0)
	sw.maxTs = time.Unix(int64(sw.ticks[len(sw.ticks)-1].Timestamp/1000), 0)

	if sw.onTick != nil {
		sw.onTick()
	}
}

func (sw *SlidingWindow) atMax(new *common.TickerData) bool {
	if sw.maxN > 0 {
		return len(sw.ticks) == sw.maxN
	} else {
		t := time.Unix(int64(new.Timestamp)/1000, 0)
		return sw.tsDiffOf(t) >= sw.maxDiff
	}
}

func (sw *SlidingWindow) Get() []*common.TickerData {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	return sw.ticks
}

func (sw *SlidingWindow) TsDiff() time.Duration {
	return sw.tsDiffOf(sw.maxTs)
}

func (sw *SlidingWindow) tsDiffOf(o time.Time) time.Duration {
	if sw.minTs.IsZero() {
		return 0
	}
	return o.Sub(sw.minTs)
}

func (sw *SlidingWindow) Reset() {
	sw.ticks = []*common.TickerData{}
}

//buySell general recommendation;
//1 = buy, 0 = unknown/stay, -1 = sell
func (sw *SlidingWindow) BuySell() common.Action {
	ml := sw.meanLog()

	if ml > 0 {
		return common.BUY
	} else if ml == 0 {
		return common.STAY
	}
	return common.SELL
}

func (sw *SlidingWindow) TakeAction() <-chan struct{} {
	ch := make(chan struct{})

	if sw.maxN == 0 {
		go func() {
			t := time.NewTicker(sw.maxDiff)

			for range t.C {
				ch <- struct{}{}
			}
		}()
	} else {
		go func() {
			n := 0
			sw.onTick = func() {
				n++
				if n%sw.maxN == 0 {
					n = 0
					ch <- struct{}{}
				}
			}
		}()
	}

	return ch
}

func (sw *SlidingWindow) meanLog() float64 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	sum := 0.0
	n := 0.0

	for i := 1; i < len(sw.ticks); i++ {
		a := sw.ticks[i].BestAsk
		b := sw.ticks[i-1].BestAsk

		if a == 0 || b == 0 {
			a = sw.ticks[i].ClosePriceChange24h
			b = sw.ticks[i-1].ClosePriceChange24h
		}

		if a == b {
			continue
		}

		s := math.Log(a / b)
		n++
		if s < 0 {
			sum += -1
		} else {
			sum += 1
		}
	}

	return sum / n
}
