package orders

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gogo/status"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"pm.tcfw.com.au/source/ataas/api/pb/blocks"
	ordersAPI "pm.tcfw.com.au/source/ataas/api/pb/orders"
	ticksAPI "pm.tcfw.com.au/source/ataas/api/pb/ticks"
	"pm.tcfw.com.au/source/ataas/db"
	"pm.tcfw.com.au/source/ataas/internal/exchanges"
	migrate "pm.tcfw.com.au/source/ataas/internal/orders/db"
)

const (
	tblName = "orders"
)

var (
	allColumns = []string{
		"id",
		"block_id",
		"side",
		"price",
		"quantity",
		"ts",
	}
)

type Server struct {
	ordersAPI.UnimplementedOrdersServiceServer

	log *logrus.Logger
}

func NewServer(ctx context.Context) (*Server, error) {
	s := &Server{
		log: logrus.New(),
	}

	err := s.Migrate(ctx)
	if err != nil {
		return nil, err
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

func (s *Server) Create(ctx context.Context, req *ordersAPI.CreateRequest) (*ordersAPI.CreateResponse, error) {
	blocksSvc, err := blocksSvc()
	if err != nil {
		return nil, err
	}

	block, err := blocksSvc.Get(ctx, &blocks.GetRequest{Id: req.BlockID})
	if err != nil {
		return nil, err
	}

	markets := initForUser(ctx)
	market, exists := markets[block.Market]
	if !exists {
		return nil, status.Error(codes.FailedPrecondition, "market not supported")
	}

	var exchangeRes exchanges.OrderResponse

	currentMarketPrice, err := s.getMarketPrice(ctx, block.Market, block.Instrument)
	if err != nil {
		return nil, err
	}
	if req.Price <= 0 {
		req.Price = float32(float64(currentMarketPrice) * block.BaseUnits)
	}

	switch req.Action {
	case ordersAPI.Action_BUY:
		exchangeRes, err = market.Buy(block.Instrument, req.Price, req.Units)
	case ordersAPI.Action_SELL:
		exchangeRes, err = market.Sell(block.Instrument, req.Price, req.Units)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execute order: %s", err)
	}

	id := uuid.New().String()
	t := time.Now()

	price := exchangeRes.Price()
	if price == 0 {
		price = currentMarketPrice
	}

	q := db.Build().Insert(tblName).Columns(allColumns...).Values(
		id,
		req.BlockID,
		req.Action == ordersAPI.Action_BUY,
		price,
		exchangeRes.Units(),
		t,
	)

	err = db.SimpleExec(ctx, q)
	if err != nil {
		return nil, err
	}

	order := &ordersAPI.Order{
		Id:        id,
		BlockID:   req.BlockID,
		Action:    req.Action,
		Units:     exchangeRes.Units(),
		Price:     exchangeRes.Price(),
		Timestamp: t.Format(time.RFC3339),
	}

	return &ordersAPI.CreateResponse{Order: order}, nil
}

func (s *Server) Get(ctx context.Context, req *ordersAPI.GetRequest) (*ordersAPI.GetResponse, error) {
	q := db.Build().Select(allColumns...).From(tblName).Where(sq.Eq{"block_id": req.BlockID}).OrderBy("ts")

	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	defer done()

	orders := []*ordersAPI.Order{}

	for res.Next() {
		order := &ordersAPI.Order{}

		var side bool
		var t time.Time

		err := res.Scan(
			&order.Id,
			&order.BlockID,
			&side,
			&order.Price,
			&order.Units,
			&t,
		)
		if err != nil {
			return nil, err
		}

		order.Action = ordersAPI.Action_SELL
		if side {
			order.Action = ordersAPI.Action_BUY
		}

		order.Timestamp = t.Format(time.RFC3339)

		orders = append(orders, order)
	}

	return &ordersAPI.GetResponse{Orders: orders}, nil
}

func (s *Server) getMarketPrice(ctx context.Context, market, instrument string) (float32, error) {
	ticks, err := ticksSvc()
	if err != nil {
		return 0, err
	}

	trades, err := ticks.Trades(ctx, &ticksAPI.GetRequest{
		Market:     market,
		Instrument: instrument,
		Depth:      1,
	})
	if err != nil {
		return 0, err
	}

	if len(trades.Data) == 0 {
		return 0, fmt.Errorf("no data")
	}

	return trades.Data[0].Amount, nil
}
