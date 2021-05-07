package blocks

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"pm.tcfw.com.au/source/trader/api/pb/blocks"
	"pm.tcfw.com.au/source/trader/api/pb/orders"
	"pm.tcfw.com.au/source/trader/api/pb/strategy"
	"pm.tcfw.com.au/source/trader/db"
)

var (
	ErrSameState = errors.New("same state")
)

func (s *Server) work(id int) {
	for a := range s.applyCh {
		err := s.handleApply(id, a)
		if err != nil {
			s.log.Errorf("failed to apply: %s", err)
		}
	}
}

func (s *Server) handleApply(wid int, ap *apply) error {
	desiredState := ap.block.State
	n := 1

	if ap.block.State == blocks.BlockState_ENDED {
		return nil
	}

	switch ap.action {
	case strategy.Action_BUY:
		if ap.block.State == blocks.BlockState_NOTHING {
			desiredState = blocks.BlockState_PURCHASED
		} else if ap.block.State == blocks.BlockState_SOLD {
			if ap.block.ShortSellAllowed {
				n = 2
			}
			desiredState = blocks.BlockState_PURCHASED
		}
	case strategy.Action_SELL:
		if ap.block.State == blocks.BlockState_PURCHASED {
			if ap.block.ShortSellAllowed {
				n = 2
			}
			desiredState = blocks.BlockState_SOLD
		} else if ap.block.State == blocks.BlockState_NOTHING {
			if ap.block.ShortSellAllowed {
				desiredState = blocks.BlockState_SOLD
			}
		}
	case strategy.Action_STAY:
		//noop
	}

	_, err := s.applyState(ap.block, desiredState, n)
	if err == ErrSameState {
		return nil
	}

	return err
}

func (s *Server) applyState(b *blocks.Block, ns blocks.BlockState, n int) (*orders.Order, error) {
	if b.State == ns {
		//no change
		return nil, ErrSameState
	}

	ordersSvc, err := ordersSvc()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	nUnits := b.CurrentUnits
	unitDiff := b.BaseUnits * float32(n)
	var price float32
	var order *orders.Order

	switch ns {
	case blocks.BlockState_PURCHASED:
		//buy
		resp, err := ordersSvc.Create(ctx, &orders.CreateRequest{
			BlockID: b.Id,
			Action:  orders.Action_BUY,
			Units:   unitDiff,
			Price:   -1, //market
		})
		if err != nil {
			return nil, err
		}
		order = resp.Order

		nUnits += order.Units
		price = order.Price
	case blocks.BlockState_SOLD:
		//sell
		resp, err := ordersSvc.Create(ctx, &orders.CreateRequest{
			BlockID: b.Id,
			Action:  orders.Action_SELL,
			Units:   unitDiff,
			Price:   -1, //market
		})
		if err != nil {
			return nil, err
		}
		order = resp.Order

		nUnits -= order.Units
		price = 0.0
	case blocks.BlockState_ENDED:
		if b.State == blocks.BlockState_PURCHASED {
			resp, err := ordersSvc.Create(ctx, &orders.CreateRequest{
				BlockID: b.Id,
				Action:  orders.Action_SELL,
				Units:   b.CurrentUnits,
				Price:   -1, //market
			})
			if err != nil {
				return nil, err
			}
			order = resp.Order

			nUnits = 0
			price = 0.0
		}
	default:
		return nil, fmt.Errorf("unknown desired state")
	}

	//Store state
	q := db.Build().Update(tblName).SetMap(sq.Eq{
		"state":         ns,
		"current_units": nUnits,
		"purchase":      price,
	}).Where(sq.Eq{"id": b.Id}).Limit(1)

	if err := db.SimpleExec(ctx, q); err != nil {
		return nil, err
	}

	return order, nil
}
