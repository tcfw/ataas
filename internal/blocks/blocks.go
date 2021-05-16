package blocks

import (
	"context"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	blocksAPI "pm.tcfw.com.au/source/trader/api/pb/blocks"
	"pm.tcfw.com.au/source/trader/api/pb/orders"
	"pm.tcfw.com.au/source/trader/api/pb/strategy"
	"pm.tcfw.com.au/source/trader/db"
	migrate "pm.tcfw.com.au/source/trader/internal/blocks/db"
	"pm.tcfw.com.au/source/trader/internal/broadcast"
	"pm.tcfw.com.au/source/trader/internal/strategies"
)

const (
	tblName                  = "blocks"
	defaultBackoutPercentage = 0.05
)

var (
	allColumns = []string{
		"id",
		"strategy_id",
		"state",
		"base_units",
		"current_units",
		"purchase",
		"watch_duration",
		"short_sell_allowed",
		"backout_percentage",
		"market",
		"instrument",
	}
)

func NewServer(ctx context.Context) (*Server, error) {
	return NewServerNWorkers(ctx, 5)
}

func NewServerNWorkers(ctx context.Context, n int) (*Server, error) {
	s := &Server{
		log:     logrus.New(),
		applyCh: make(chan *apply, 5),
	}

	err := s.Migrate(ctx)
	if err != nil {
		return nil, err
	}

	for i := 0; i < n; i++ {
		go s.work(i)
	}

	return s, nil
}

type Server struct {
	blocksAPI.UnimplementedBlocksServiceServer

	log     *logrus.Logger
	unsub   func() error
	applyCh chan *apply

	workWg sync.WaitGroup
}

type apply struct {
	action strategy.Action
	block  *blocksAPI.Block
}

func (s *Server) Migrate(ctx context.Context) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return migrate.Migrate(ctx, conn.Conn(), s.log)
}

func (s *Server) Stop() {
	close(s.applyCh)

	//Wait for all processing orders to complete
	s.workWg.Wait()
}

func (s *Server) Listen() error {
	b, err := broadcast.Driver()
	if err != nil {
		return err
	}

	unsub, err := b.Subscribe("STRAT.action", s.handleAction)
	if err != nil {
		return err
	}

	s.unsub = unsub

	return nil
}

func (s *Server) handleAction(data *strategies.ActionEvent) {
	s.log.Infoln("Handling STRAT action")

	q := db.Build().Select(allColumns...).From(tblName).Where(sq.Eq{"strategy_id": data.StrategyID})
	res, done, err := db.SimpleQuery(context.Background(), q)
	if err != nil {
		s.log.Errorf("failed to find blocks: %s", err)
		return
	}
	defer done()

	n := 0

	for res.Next() {
		block := &blocksAPI.Block{}
		err := res.Scan(
			&block.Id,
			&block.StrategyId,
			&block.State,
			&block.BaseUnits,
			&block.CurrentUnits,
			&block.Purchase,
			&block.WatchDuration,
			&block.ShortSellAllowed,
			&block.BackoutPercentage,
			&block.Market,
			&block.Instrument,
		)
		if err != nil {
			s.log.Errorf("failed to scan block: %s", err)
			continue
		}
		s.applyCh <- &apply{data.Action, block}
		n++
	}

	s.log.Infof("Applied to %d blocks", n)
}

func (s *Server) New(ctx context.Context, req *blocksAPI.Block) (*blocksAPI.Block, error) {
	if req.BackoutPercentage == 0 {
		req.BackoutPercentage = defaultBackoutPercentage
	}

	req.Id = uuid.New().String()

	err := s.validateBlock(req)
	if err != nil {
		return nil, err
	}

	q := db.Build().Insert(tblName).Columns(allColumns...).Values(
		req.Id,
		req.StrategyId,
		blocksAPI.BlockState_NOTHING,
		req.BaseUnits,
		0,
		0,
		req.WatchDuration,
		false,
		req.BackoutPercentage,
		req.Market,
		req.Instrument,
	)

	if err := db.SimpleExec(ctx, q); err != nil {
		return nil, err
	}

	return req, nil
}

func (s *Server) validateBlock(b *blocksAPI.Block) error {
	//TODO(tcfw): check if strategy exists
	if b.StrategyId == "" {
		return status.Error(codes.FailedPrecondition, "strategy required")
	}

	if b.Market == "" {
		return status.Error(codes.FailedPrecondition, "market required")
	}

	if b.Instrument == "" {
		return status.Error(codes.FailedPrecondition, "instrument required")
	}

	if b.Market == "" {
		return status.Error(codes.FailedPrecondition, "market required")
	}

	if b.BaseUnits <= 0 {
		return status.Error(codes.FailedPrecondition, "base units required")
	}

	if b.BackoutPercentage <= 0 {
		return status.Error(codes.FailedPrecondition, "backout percentage required")
	}

	return nil
}

func (s *Server) List(ctx context.Context, req *blocksAPI.ListRequest) (*blocksAPI.ListResponse, error) {
	q := db.Build().Select(allColumns...).From(tblName)
	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		s.log.Errorf("failed to find blocks: %s", err)
		return nil, err
	}
	defer done()

	blocks := []*blocksAPI.Block{}

	for res.Next() {
		block := &blocksAPI.Block{}

		err := res.Scan(
			&block.Id,
			&block.StrategyId,
			&block.State,
			&block.BaseUnits,
			&block.CurrentUnits,
			&block.Purchase,
			&block.WatchDuration,
			&block.ShortSellAllowed,
			&block.BackoutPercentage,
			&block.Market,
			&block.Instrument,
		)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	return &blocksAPI.ListResponse{Blocks: blocks}, nil
}

func (s *Server) Get(ctx context.Context, req *blocksAPI.GetRequest) (*blocksAPI.Block, error) {
	q := db.Build().Select(allColumns...).From(tblName).Where(sq.Eq{"id": req.Id})
	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		s.log.Errorf("failed to find blocks: %s", err)
		return nil, err
	}
	defer done()

	if !res.Next() {
		return nil, status.Error(codes.NotFound, "block not found")
	}

	block := &blocksAPI.Block{}
	err = res.Scan(
		&block.Id,
		&block.StrategyId,
		&block.State,
		&block.BaseUnits,
		&block.CurrentUnits,
		&block.Purchase,
		&block.WatchDuration,
		&block.ShortSellAllowed,
		&block.BackoutPercentage,
		&block.Market,
		&block.Instrument,
	)
	if err != nil {
		s.log.Errorf("failed to scan block: %s", err)
		return nil, err
	}

	return block, nil
}

func (s *Server) ManualAction(ctx context.Context, req *blocksAPI.ManualRequest) (*blocksAPI.ManualResponse, error) {
	block, err := s.Get(ctx, &blocksAPI.GetRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}

	alreadyPurchased := block.State == blocksAPI.BlockState_PURCHASED && req.Action == orders.Action_BUY
	alreadySold := block.State == blocksAPI.BlockState_SOLD && req.Action == orders.Action_SELL
	if alreadyPurchased || alreadySold {
		return nil, status.Errorf(codes.FailedPrecondition, "invalid future state %t %t", alreadyPurchased, alreadySold)
	}

	var order *orders.Order
	if req.Action == orders.Action_BUY {
		order, err = s.applyState(block, blocksAPI.BlockState_PURCHASED, 1)
	} else {
		order, err = s.applyState(block, blocksAPI.BlockState_SOLD, 1)
	}
	if err != nil {
		return nil, err
	}

	return &blocksAPI.ManualResponse{Order: order}, nil
}

func (s *Server) Delete(ctx context.Context, req *blocksAPI.DeleteRequest) (*blocksAPI.DeleteResponse, error) {
	block, err := s.Get(ctx, &blocksAPI.GetRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}

	if block.State == blocksAPI.BlockState_PURCHASED {
		_, err = s.applyState(block, blocksAPI.BlockState_ENDED, 1)
		if err != nil {
			return nil, fmt.Errorf("failed to sell before delete: %s", err)
		}
	}

	q := db.Build().Delete(tblName).Where(sq.Eq{"id": block.Id}).Limit(1)
	err = db.SimpleExec(ctx, q)
	if err != nil {
		return nil, err
	}

	return &blocksAPI.DeleteResponse{}, nil
}
