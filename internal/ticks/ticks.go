package ticks

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pm.tcfw.com.au/source/trader/api/pb/ticks"
	"pm.tcfw.com.au/source/trader/db"
	migrate "pm.tcfw.com.au/source/trader/internal/ticks/db"
)

var (
	ErrUnknownInterval = errors.New("unknown interval")
	ErrDurationTooLong = errors.New("duration too long")
)

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

type Server struct {
	ticks.UnimplementedHistoryServiceServer

	log *logrus.Logger
}

func (s *Server) Migrate(ctx context.Context) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return migrate.Migrate(ctx, conn.Conn(), s.log)
}

//TradesRange provides trades of a specific instrument/market within a specific time span from now
func (s *Server) TradesRange(ctx context.Context, req *ticks.RangeRequest) (*ticks.TradesResponse, error) {
	if req.Instrument == "" || req.Market == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required arguments")
	}

	ts, err := time.ParseDuration(req.Since)
	if err != nil {
		return nil, err
	}

	if ts > 48*time.Hour {
		return nil, ErrDurationTooLong
	}

	query := db.Build().Select("tradeid", "ts", "direction", "amount", "units").
		From("trades").
		Where(sq.And{sq.Eq{"market": req.Market, "instrument": req.Instrument}, sq.GtOrEq{"ts": time.Now().Add(-1 * ts)}}).
		OrderBy("ts DESC")

	return s.readTrades(ctx, query, req.Market, req.Instrument)
}

//Ticks provides a history of transactions for a particular instrument
func (s *Server) Trades(ctx context.Context, req *ticks.GetRequest) (*ticks.TradesResponse, error) {
	if req.Instrument == "" || req.Market == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required arguments")
	}

	if req.Depth == 0 {
		req.Depth = 100
	}

	query := db.Build().Select("tradeid", "ts", "direction", "amount", "units").
		From("trades").
		Where(sq.Eq{"market": req.Market, "instrument": req.Instrument}).
		OrderBy("ts DESC").
		Limit(uint64(req.Depth))

	return s.readTrades(ctx, query, req.Market, req.Instrument)
}

func (s *Server) readTrades(ctx context.Context, query sq.Sqlizer, market, instrument string) (*ticks.TradesResponse, error) {
	res, done, err := db.SimpleQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	defer done()

	data := []*ticks.Trade{}

	for res.Next() {
		trade := &ticks.Trade{
			Market:     market,
			Instrument: instrument,
		}

		ts := &time.Time{}
		direction := false

		err := res.Scan(&trade.TradeID, &ts, &direction, &trade.Amount, &trade.Units)
		if err != nil {
			return nil, err
		}

		trade.Timestamp = ts.Unix()

		if direction {
			trade.Direction = ticks.TradeDirection_SELL
		}

		data = append(data, trade)
	}

	return &ticks.TradesResponse{Data: data}, nil
}

//Candles provides a means of sumarising trades into OHLCV formats for a particular instrument
func (s *Server) Candles(ctx context.Context, req *ticks.CandlesRequest) (*ticks.CandlesResponse, error) {
	if req.Instrument == "" || req.Market == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required arguments")
	}

	tsExpr, err := s.timeframeExpr(req.Interval)
	if err != nil {
		return nil, err
	}

	if req.Depth == 0 {
		req.Depth = 100
	}

	query := db.Build().Select(
		"array_agg(amount)[1]",
		"array_agg(amount)[array_length(array_agg(amount), 1)]",
		"max(amount)",
		"min(amount)",
		"count(*)",
		tsExpr,
	).From("trades").
		Where(sq.Eq{"market": req.Market, "instrument": req.Instrument}).
		GroupBy("timestamp").OrderBy("timestamp DESC").Limit(uint64(req.Depth))

	res, done, err := db.SimpleQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	defer done()

	data := []*ticks.OHLCV{}

	for res.Next() {
		tick := &ticks.OHLCV{}
		ts := time.Time{}
		err := res.Scan(&tick.Open, &tick.Close, &tick.High, &tick.Low, &tick.Volume, &ts)
		if err != nil {
			return nil, err
		}
		tick.Timestamp = ts.Unix()
		data = append(data, tick)
	}

	return &ticks.CandlesResponse{Data: data}, nil
}

func (s *Server) timeframeExpr(interval string) (string, error) {
	suffix := ""
	largeUnit := ""
	smallUnit := ""
	multipler := 0
	switch interval {
	case "1m", "5m", "15m", "30m":
		suffix = "m"
		largeUnit = "hour"
		smallUnit = "minute"
		multipler = 60
	case "1h", "4h", "6h", "12h":
		suffix = "h"
		largeUnit = "day"
		smallUnit = "hour"
		multipler = 60 * 60
	case "1d":
		return "date_trunc('day', ts)", nil
	default:
		return "", ErrUnknownInterval
	}

	d, err := strconv.Atoi(strings.TrimSuffix(interval, suffix))
	if err != nil {
		return "", nil
	}

	expr := "date_trunc('%s', ts) + ((extract('%s', ts) / %d)::int * %d)::interval as timestamp"

	return fmt.Sprintf(expr, largeUnit, smallUnit, d, d*multipler), nil
}

func (s *Server) RangeCompare(ctx context.Context, req *ticks.CompareRequest) (*ticks.CompareResponse, error) {
	if req.Instrument == "" || req.Market == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required arguments")
	}

	tsExpr, err := s.timeframeExpr(req.Interval)
	if err != nil {
		return nil, err
	}

	query := db.Build().Select(
		"array_agg(amount)[1]",
		"array_agg(amount)[array_length(array_agg(amount), 1)]",
		tsExpr,
	).From("trades").
		Where(sq.Eq{"market": req.Market, "instrument": req.Instrument}).
		GroupBy("timestamp").OrderBy("timestamp DESC").Limit(2)

	res, done, err := db.SimpleQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	defer done()

	openClose := []struct{ o, c float32 }{}

	for res.Next() {
		tick := struct{ o, c float32 }{}
		ts := &time.Time{}

		if err := res.Scan(&tick.o, &tick.c, &ts); err != nil {
			return nil, err
		}

		openClose = append(openClose, tick)
	}

	diff := (openClose[0].c / (openClose[1].o + 0.000000000000001)) - 1

	return &ticks.CompareResponse{Difference: diff}, nil
}
