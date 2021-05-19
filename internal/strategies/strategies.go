package strategies

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gogo/status"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
	"pm.tcfw.com.au/source/ataas/db"
	migrate "pm.tcfw.com.au/source/ataas/internal/strategies/db"
)

type Action int8

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

	strategy := &strategy.Strategy{}
	err = res.Scan(
		&strategy.Id,
		&strategy.Market,
		&strategy.Instrument,
		&strategy.Strategy,
		&strategy.Params,
		&strategy.Duration,
		&strategy.Next,
	)
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
