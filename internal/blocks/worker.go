package blocks

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"pm.tcfw.com.au/source/ataas/api/pb/blocks"
	"pm.tcfw.com.au/source/ataas/api/pb/notify"
	"pm.tcfw.com.au/source/ataas/api/pb/orders"
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
	"pm.tcfw.com.au/source/ataas/db"
)

var (
	ErrSameState = errors.New("same state")
)

func (s *Server) work(id int) {
	s.workWg.Add(1)
	defer s.workWg.Done()

	for a := range s.applyCh {
		err := s.handleApply(id, a)
		if err != nil {
			s.log.Errorf("failed to apply: %s", err)
		}
	}
}

func (s *Server) handleApply(wid int, ap *apply) error {

	if ap.block.State == blocks.BlockState_ENDED {
		return nil
	}

	desiredState, n := s.calcState(ap.block, ap.action)

	_, err := s.applyState(ap.block, desiredState, n)
	if err == ErrSameState {
		return nil
	}

	return err
}

func (s *Server) calcState(block *blocks.Block, action strategy.Action) (blocks.BlockState, int) {
	desiredState := block.State
	n := 1

	switch action {
	case strategy.Action_BUY:
		if block.State == blocks.BlockState_NOTHING {
			desiredState = blocks.BlockState_PURCHASED
		} else if block.State == blocks.BlockState_SOLD {
			if block.ShortSellAllowed {
				n = 2
			}
			desiredState = blocks.BlockState_PURCHASED
		}
	case strategy.Action_SELL:
		if block.State == blocks.BlockState_PURCHASED {
			if block.ShortSellAllowed {
				n = 2
			}
			desiredState = blocks.BlockState_SOLD
		} else if block.State == blocks.BlockState_NOTHING {
			if block.ShortSellAllowed {
				desiredState = blocks.BlockState_SOLD
			}
		}
	case strategy.Action_STAY:
		//noop
	}

	return desiredState, n
}

func (s *Server) applyState(b *blocks.Block, ns blocks.BlockState, n int) (*orders.Order, error) {
	if b.State == ns {
		//no change
		return nil, ErrSameState
	}

	s.log.Warnf("Applying state to block %s: %s x%d", b.Id, ns, n)

	ordersSvc, err := ordersSvc()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	nUnits := b.CurrentUnits
	// unitDiff := b.BaseUnits * float64(n)
	// if b.CurrentUnits < unitDiff {
	unitDiff := b.CurrentUnits
	// }

	var price float32 = -1
	var order *orders.Order

	switch ns {
	case blocks.BlockState_PURCHASED:
		//buy
		if b.Purchase > 0 {
			price = b.Purchase
		}

		resp, err := ordersSvc.Create(ctx, &orders.CreateRequest{
			BlockID: b.Id,
			Action:  orders.Action_BUY,
			Units:   unitDiff,
			Price:   price,
		})
		if err != nil {
			notifyFail(ctx, b, ns, err)
			return nil, err
		}
		order = resp.Order
		nUnits += order.Units

	case blocks.BlockState_SOLD:
		if b.Purchase > 0 {
			unitDiff = b.CurrentUnits
		}

		//sell
		resp, err := ordersSvc.Create(ctx, &orders.CreateRequest{
			BlockID: b.Id,
			Action:  orders.Action_SELL,
			Units:   unitDiff,
			Price:   -1, //market
		})
		if err != nil {
			notifyFail(ctx, b, ns, err)
			return nil, err
		}
		order = resp.Order
		nUnits -= order.Units
		if nUnits < 0 && !b.ShortSellAllowed {
			nUnits = 0
		}
		if b.Purchase > 0 {
			b.Purchase = float32(resp.Order.Units * float64(resp.Order.Price))
		} else {
			//This is mainly to account for fees
			b.BaseUnits = resp.Order.Units
		}

	case blocks.BlockState_ENDED:
		if b.State == blocks.BlockState_PURCHASED {
			resp, err := ordersSvc.Create(ctx, &orders.CreateRequest{
				BlockID: b.Id,
				Action:  orders.Action_SELL,
				Units:   unitDiff,
				Price:   -1, //market
			})
			if err != nil {
				notifyFail(ctx, b, ns, err)
				return nil, err
			}
			order = resp.Order

			nUnits = 0
		}
	default:
		return nil, fmt.Errorf("unknown desired state")
	}

	//Store state
	q := db.Build().Update(tblName).SetMap(sq.Eq{
		"state":         ns,
		"current_units": int(nUnits * 1000000),
		"purchase":      b.Purchase,
		"base_units":    b.BaseUnits,
	}).Where(sq.Eq{"id": b.Id}).Limit(1)

	if err := db.SimpleExec(ctx, q); err != nil {
		return nil, err
	}

	notifyOrder(ctx, b, ns, order)

	return order, nil
}

func notifyFail(ctx context.Context, block *blocks.Block, state blocks.BlockState, errStr error) {
	nSvc, err := notifySvc()
	if err != nil {
		return
	}

	nSvc.Send(ctx, &notify.SendRequest{
		Uid:   block.Account,
		Type:  notify.SendRequest_BLOCK,
		Title: fmt.Sprintf("Error attempting order on %s", block.Instrument),
		Body:  fmt.Sprintf("Failed to create order on %s - %s<br/><br/>%s", block.Instrument, state, errStr.Error()),
	})
}

func notifyOrder(ctx context.Context, block *blocks.Block, state blocks.BlockState, order *orders.Order) {
	nSvc, err := notifySvc()
	if err != nil {
		return
	}
	nSvc.Send(ctx, &notify.SendRequest{
		Uid:   block.Account,
		Type:  notify.SendRequest_BLOCK,
		Title: fmt.Sprintf("New Order - %s %s", block.Instrument, state),
		Body:  fmt.Sprintf("%v %v", order.Price, order.Units),
	})
}
