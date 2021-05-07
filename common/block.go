package common

import (
	"fmt"
	"sync"
	"time"
)

type Action uint8

const (
	BUY Action = 1 + iota
	SELL
	STAY
)

func ActionToString(a Action) string {
	switch a {
	case BUY:
		return "BUY"
	case SELL:
		return "SELL"
	case STAY:
		return "STAY"
	default:
		return "UNKNOWN"
	}
}

type Strategy interface {
	CurrentAsk() float64
	BuySell() Action
	TakeAction() <-chan struct{}
}

type OrderController interface {
	Buy(units float64, costPerUnit float64) (float64, error)
	Sell(units float64) error
}

type BlockState uint16

const (
	NOTHING BlockState = iota
	PURCHASED
	SOLD
	BACKOUT
)

type UnitBlock struct {
	Market            string        `json:"market"`
	Instrument        string        `json:"instrument"`
	BaseUnits         float64       `json:"base_units"`
	CurrentUnits      float64       `json:"units"`
	Purchase          float64       `json:"price"`
	LastSell          float64       `json:"sell"`
	Initial           float64       `json:"initial"`
	State             BlockState    `json:"state"`
	WatchDuration     time.Duration `json:"duration"`
	ShortSellAllowed  bool          `json:"short_allowed"`
	BackoutPercentage float64       `json:"backout_percentage"`

	GSStrategy Strategy
	Controller OrderController

	events chan *Event

	mu sync.Mutex
}

func (b *UnitBlock) StartWatch() {
	go func() {
		t := time.NewTicker(b.WatchDuration)
		for range t.C {
			if b.shouldBackOut() {
				b.BackOut()
			}
		}
	}()

	go func() {
		for range b.GSStrategy.TakeAction() {
			b.Execute()
		}
	}()
}

func (b *UnitBlock) Events() <-chan *Event {
	if b.events == nil {
		b.events = make(chan *Event)
	}

	return b.events
}

func (b *UnitBlock) Execute() {
	bsState := b.GSStrategy.BuySell()

	switch bsState {
	case BUY:
		if b.State == NOTHING {
			b.Buy(1)
		} else if b.State == SOLD {
			n := int8(1)
			if b.ShortSellAllowed {
				n = 2
			}
			b.Buy(n)
		}
	case SELL:
		if b.State == PURCHASED {
			n := int8(1)
			if b.ShortSellAllowed {
				n = 2
			}
			b.Sell(n)
		} else if b.State == NOTHING {
			if b.ShortSellAllowed {
				b.Sell(1)
			} else {
				b.sendEvent(&Event{Action: STAY, Price: b.GSStrategy.CurrentAsk()})
			}
		}
	case STAY:
		//
	}

	fmt.Printf("BS: %v, S: %+v\n", bsState, b)
}

func (b *UnitBlock) sendEvent(e *Event) {
	e.Market = b.Market
	e.Instrument = b.Instrument

	select {
	case b.events <- e:
	default:
	}
}

func (b *UnitBlock) shouldBackOut() bool {
	return (b.GSStrategy.CurrentAsk()*b.CurrentUnits) < (b.Initial*b.BackoutPercentage) && b.State == PURCHASED
}

func (b *UnitBlock) BackOut() {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := b.CurrentUnits
	price := b.GSStrategy.CurrentAsk()

	b.Controller.Sell(n)
	b.LastSell = n * price
	b.CurrentUnits = 0
	b.State = NOTHING
	b.sendEvent(&Event{Action: SELL, Units: n, Price: price})

}

func (b *UnitBlock) Buy(n int8) {
	b.mu.Lock()
	defer b.mu.Unlock()

	nUnits := b.BaseUnits * float64(n)
	price := b.GSStrategy.CurrentAsk()

	purchasedUnits, err := b.Controller.Buy(nUnits, price)
	if err != nil {
		fmt.Println("[error-order] failed to order", err)
	}
	b.CurrentUnits += purchasedUnits
	b.Purchase = purchasedUnits * price
	b.State = PURCHASED

	if b.Initial == 0 {
		b.Initial = b.Purchase
	}

	b.sendEvent(&Event{Action: BUY, Units: purchasedUnits, Price: price})
}

func (b *UnitBlock) Sell(n int8) {
	b.mu.Lock()
	defer b.mu.Unlock()

	nUnits := b.CurrentUnits * float64(n)
	price := b.GSStrategy.CurrentAsk()

	b.Controller.Sell(nUnits)
	b.LastSell = nUnits * price
	b.CurrentUnits -= nUnits
	b.State = SOLD
	b.sendEvent(&Event{Action: SELL, Units: nUnits, Price: price})
}
