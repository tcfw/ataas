package ticks

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
	"pm.tcfw.com.au/source/ataas/db"
	migrate "pm.tcfw.com.au/source/ataas/internal/ticks/db"
)

var (
	ErrUnknownInterval = errors.New("unknown interval")
	ErrDurationTooLong = errors.New("duration too long")
)

func NewServer(ctx context.Context) (*Server, error) {
	log := logrus.New()

	libDir := viper.GetString("collector.library")
	if libDir == "" {
		return nil, fmt.Errorf("no library loc provided")
	}
	log.Infof("Loading trades library: %s", libDir)

	lib, err := NewLibrary(libDir, log)
	if err != nil {
		return nil, err
	}

	s := &Server{
		log:     log,
		library: lib,
	}

	err = s.Migrate(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

type Server struct {
	ticks.UnimplementedHistoryServiceServer

	log *logrus.Logger

	library *TradeLibrary
}

func (s *Server) Migrate(ctx context.Context) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return migrate.Migrate(ctx, conn.Conn(), s.log)
}

func (s *Server) Close() error {
	return s.library.Close()
}

//TradesRange provides trades of a specific instrument/market within a specific time span from now
func (s *Server) TradesRange(ctx context.Context, req *ticks.RangeRequest) (*ticks.TradesResponse, error) {
	if req.Instrument == "" || req.Market == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required arguments")
	}

	tsUntil := time.Now()

	tsFrom, wasDuration, err := parseTime(req.Since)
	if err != nil {
		return nil, err
	}

	if req.Until != "" {
		t, _, err := parseTime(req.Until)
		if err != nil {
			return nil, err
		}
		tsUntil = t
	}

	if wasDuration && req.Until != "" {
		ts, err := time.ParseDuration(req.Since)
		if err != nil {
			return nil, err
		}
		tsFrom = tsUntil.Add(-ts)
	}

	trades, err := s.library.GetSince(req.Market, req.Instrument, tsFrom, tsUntil)
	if err != nil {
		return nil, err
	}

	for _, trade := range trades {
		trade.Market = ""
		trade.Instrument = ""
		trade.TradeID = ""
	}

	return &ticks.TradesResponse{Data: trades}, nil
}

//Ticks provides a history of transactions for a particular instrument
func (s *Server) Trades(ctx context.Context, req *ticks.GetRequest) (*ticks.TradesResponse, error) {
	if req.Instrument == "" || req.Market == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required arguments")
	}

	if req.Depth == 0 {
		req.Depth = 100
	}

	trades, err := s.library.GetSince(req.Market, req.Instrument, time.Now().Add(-time.Duration(req.Depth+60)*time.Second), time.Now())
	if err != nil {
		return nil, err
	}

	c := int32(len(trades)) - req.Depth
	if c < 0 {
		c = 0
	}

	for _, trade := range trades[c:] {
		trade.Market = ""
		trade.Instrument = ""
		trade.TradeID = ""
	}

	return &ticks.TradesResponse{Data: trades[c:]}, nil
}

func (s *Server) TradesRangeStream(req *ticks.RangeRequest, stream ticks.HistoryService_TradesRangeStreamServer) error {
	if req.Instrument == "" || req.Market == "" {
		return status.Error(codes.InvalidArgument, "missing required arguments")
	}

	tsUntil := time.Now()

	tsFrom, wasDuration, err := parseTime(req.Since)
	if err != nil {
		return err
	}

	if req.Until != "" {
		t, _, err := parseTime(req.Until)
		if err != nil {
			return err
		}
		tsUntil = t
	}

	if wasDuration && req.Until != "" {
		ts, err := time.ParseDuration(req.Since)
		if err != nil {
			return err
		}
		tsFrom = tsUntil.Add(-ts)
	}

	tradesCh, err := s.library.GetSinceStream(req.Market, req.Instrument, tsFrom, tsUntil)
	if err != nil {
		return err
	}

	for t := range tradesCh {
		err := stream.Send(t)
		if err != nil {
			return err
		}
	}

	return nil
}

//Candles provides a means of sumarising trades into OHLCV formats for a particular instrument
func (s *Server) Candles(ctx context.Context, req *ticks.CandlesRequest) (*ticks.CandlesResponse, error) {
	if req.Instrument == "" || req.Market == "" || req.Interval == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required arguments")
	}

	interval, err := time.ParseDuration(req.Interval)
	if err != nil {
		return nil, err
	}

	startingPoint := time.Now().Add(-time.Duration(req.Depth) * interval).Round(interval)

	trades, err := s.library.GetSinceStream(req.Market, req.Instrument, startingPoint, time.Now())
	if err != nil {
		return nil, err
	}

	data := []*ticks.OHLCV{}

	var current *ticks.OHLCV
	var currentTs time.Time

	for trade := range trades {
		tts := trade.Timestamp
		if tts > 9999999999 {
			tts = tts / 1000
		}
		ts := time.Unix(tts, 0).Truncate(interval)
		if current == nil || ts != currentTs {
			currentTs = ts
			current = &ticks.OHLCV{
				// Market:     trade.Market,
				// Instrument: trade.Instrument,
				Open:      trade.Amount,
				High:      0,
				Low:       math.MaxFloat32,
				Timestamp: ts.Unix(),
			}
			data = append(data, current)
		}
		if trade.Amount < current.Low {
			current.Low = trade.Amount
		}
		if trade.Amount > current.High {
			current.High = trade.Amount
		}
		current.Close = trade.Amount
		current.Volume += trade.Units
	}

	return &ticks.CandlesResponse{Data: data}, nil
}

func parseTime(ts string) (time.Time, bool, error) {
	if strings.ContainsAny(ts, ":/.+") {
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return time.Time{}, true, err
		}
		return t, true, nil
	} else {
		ts, err := time.ParseDuration(ts)
		if err != nil {
			return time.Time{}, false, err
		}

		if ts > 336*time.Hour {
			return time.Time{}, false, ErrDurationTooLong
		}

		return time.Now().Add(-ts), false, nil
	}
}
