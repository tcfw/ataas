package strategies

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gogo/status"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"pm.tcfw.com.au/source/ataas/api/pb/blocks"
	blocksAPI "pm.tcfw.com.au/source/ataas/api/pb/blocks"
	ordersAPI "pm.tcfw.com.au/source/ataas/api/pb/orders"
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
	"pm.tcfw.com.au/source/ataas/db"
	migrate "pm.tcfw.com.au/source/ataas/internal/strategies/db"
)

const (
	checkT         = 1 * time.Second
	tblName        = "strategies"
	historyTblName = "strategy_history"
)

type Server struct {
	strategy.UnimplementedStrategyServiceServer

	Jobs chan *strategy.Strategy

	nWorkers int
	log      *logrus.Logger
	stop     chan struct{}
	running  bool
}

func NewServer(ctx context.Context) (*Server, error) {
	return NewServerNWorkers(ctx, 5)
}

func NewServerNWorkers(ctx context.Context, n int) (*Server, error) {
	s := &Server{
		Jobs:     make(chan *strategy.Strategy, 10),
		log:      logrus.New(),
		stop:     make(chan struct{}),
		nWorkers: n,
	}

	err := s.Migrate(ctx)
	if err != nil {
		return nil, err
	}

	for i := 0; i < n; i++ {
		go s.Work(i)
	}

	return s, nil
}

func (s *Server) Migrate(ctx context.Context) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return migrate.Migrate(ctx, conn.Conn(), s.log)
}

func (s *Server) List(ctx context.Context, req *strategy.ListRequest) (*strategy.ListResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}

	q := db.Build().Select("id", "market", "instrument", "strategy", "params", "duration", "next").
		From(tblName).OrderBy("id ASC").Limit(uint64(req.Limit))

	if req.Page != "" {
		q.Where(sq.Gt{"id": req.Page})
	}

	res, close, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	defer close()

	strategies := []*strategy.Strategy{}

	for res.Next() {
		s := &strategy.Strategy{}
		var next time.Time

		err := res.Scan(&s.Id, &s.Market, &s.Instrument, &s.Strategy, &s.Params, &s.Duration, &next)
		if err != nil {
			return nil, err
		}

		s.Next = next.Format(time.RFC3339)

		strategies = append(strategies, s)
	}

	return &strategy.ListResponse{Strategies: strategies}, nil
}

func (s *Server) History(ctx context.Context, req *strategy.HistoryRequest) (*strategy.HistoryResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}

	q := db.Build().Select("id", "action", "ts").
		From(historyTblName).OrderBy("ts DESC").Where(sq.Eq{"strategy_id": req.Id}).Limit(uint64(req.Limit))

	if req.Page != "" {
		q.Where(sq.Lt{"ts": req.Page})
	}

	res, close, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	defer close()

	events := []*strategy.HistoryAction{}

	for res.Next() {
		ev := &strategy.HistoryAction{}
		var ts time.Time

		err := res.Scan(&ev.Id, &ev.Action, &ts)
		if err != nil {
			return nil, err
		}

		ev.Timestamp = ts.Format(time.RFC3339)

		events = append(events, ev)
	}

	return &strategy.HistoryResponse{Events: events}, nil
}

func (s *Server) Create(ctx context.Context, req *strategy.CreateRequest) (*strategy.CreateResponse, error) {
	existQ := db.Build().Select("id").From(tblName).Where(sq.Eq{
		"market":     req.Strategy.Market,
		"instrument": req.Strategy.Instrument,
		"strategy":   req.Strategy.Strategy,
		"params":     req.Strategy.Params,
		"duration":   req.Strategy.Duration,
	}).Limit(1)

	exRes, done, err := db.SimpleQuery(ctx, existQ)
	if err != nil {
		return nil, err
	}

	if exRes.Next() {
		var existing string
		if err := exRes.Scan(&existing); err != nil {
			return nil, err
		}
		done()

		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("use %s", existing))
	}

	id, _ := uuid.NewRandom()
	req.Strategy.Id = id.String()

	q := db.Build().Insert(tblName).Columns("id", "market", "instrument", "strategy", "params", "duration", "next").Values(
		req.Strategy.Id,
		req.Strategy.Market,
		req.Strategy.Instrument,
		req.Strategy.Strategy,
		req.Strategy.Params,
		req.Strategy.Duration,
		time.Now().Add(5*time.Minute),
	)

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(ctx, tx, q)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &strategy.CreateResponse{Strategy: req.Strategy}, nil
}

func (s *Server) Get(ctx context.Context, req *strategy.GetRequest) (*strategy.Strategy, error) {
	q := db.Build().Select("id", "market", "instrument", "strategy", "params", "duration", "next").From(tblName).Where(sq.Eq{"id": req.Id})
	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		s.log.Errorf("failed to find blocks: %s", err)
		return nil, err
	}
	defer done()

	if !res.Next() {
		return nil, status.Error(codes.NotFound, "strategy not found")
	}

	var next time.Time

	strategy := &strategy.Strategy{}
	err = res.Scan(
		&strategy.Id,
		&strategy.Market,
		&strategy.Instrument,
		&strategy.Strategy,
		&strategy.Params,
		&strategy.Duration,
		&next,
	)

	strategy.Next = next.Format(time.RFC3339)

	if err != nil {
		s.log.Errorf("failed to scan block: %s", err)
		return nil, err
	}

	return strategy, nil
}

func (s *Server) Update(ctx context.Context, req *strategy.UpdateRequest) (*strategy.Strategy, error) {
	strategy, err := s.Get(ctx, &strategy.GetRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}

	strategy.Strategy = req.Strategy.Strategy
	strategy.Params = req.Strategy.Params
	strategy.Duration = req.Strategy.Duration
	strategy.Next = req.Strategy.Next

	q := db.Build().Update(tblName).SetMap(sq.Eq{
		"strategy": strategy.Strategy,
		"params":   strategy.Params,
		"duration": strategy.Duration,
		"next":     strategy.Next,
	}).Where(sq.Eq{"id": req.Id}).Limit(1)

	err = db.SimpleExec(ctx, q)
	if err != nil {
		return nil, err
	}

	return strategy, nil
}

func (s *Server) Delete(ctx context.Context, req *strategy.DeleteRequest) (*strategy.DeleteResponse, error) {
	q := db.Build().Delete(tblName).Where(sq.Eq{"id": req.Id}).Limit(1)
	err := db.SimpleExec(ctx, q)
	if err != nil {
		return nil, err
	}

	return &strategy.DeleteResponse{}, nil
}

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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
				if resp.State == blocks.BlockState_PURCHASED {
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
		case blocks.BlockState_PURCHASED:
			return &ordersAPI.Order{
				Action:    ordersAPI.Action_BUY,
				Price:     marketPrice,
				Units:     block.BaseUnits,
				Timestamp: ts.Format(time.RFC3339),
			}, nil
		case blocks.BlockState_SOLD:
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
