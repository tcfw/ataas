package strategies

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sirupsen/logrus"
	"pm.tcfw.com.au/source/ataas/api/pb/strategy"
	"pm.tcfw.com.au/source/ataas/db"
	"pm.tcfw.com.au/source/ataas/internal/broadcast"
)

type Worker struct {
	id  int
	log *logrus.Logger
}

type ActionEvent struct {
	Action     strategy.Action `json:"action"`
	StrategyID string          `json:"strategy"`
}

func (s *Server) Start(ctx context.Context) error {
	s.running = true
	defer func() {
		s.running = false
	}()

	t := time.NewTicker(checkT)

	for {
		select {
		case <-t.C:
			err := s.RunOnce()
			if err != nil {
				return err
			}
		case <-s.stop:
			return nil
		}
	}
}

func (s *Server) RunOnce() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	q := db.Build().Select("id", "market", "instrument", "strategy", "params", "duration").
		From(tblName).
		Where(sq.LtOrEq{"next": time.Now()})

	res, done, err := db.SimpleQuery(ctx, q)
	if err != nil {
		return err
	}
	defer done()

	var duration time.Duration

	ts := time.Now()

	for res.Next() {
		job := &strategy.Strategy{}
		err := res.Scan(&job.Id, &job.Market, &job.Instrument, &job.Strategy, &job.Params, &duration)
		if err != nil {
			return err
		}

		q := db.Build().Update(tblName).
			SetMap(sq.Eq{"next": ts.Add(duration).Round(2 * time.Second)}).
			Where(sq.Eq{"id": job.Id}).
			Limit(1)

		err = db.SimpleExec(ctx, q)
		if err != nil {
			return err
		}

		s.Jobs <- job
	}

	return nil
}

func (s *Server) Stop() {
	if s.running {
		s.stop <- struct{}{}
	}
}

func (s *Server) Work(id int) {
	w := &Worker{
		id:  id,
		log: s.log,
	}

	for job := range s.Jobs {
		err := w.HandleJob(job)
		if err != nil {
			s.log.Errorf("worker[%d] error: %s", w.id, err)
		}
	}
}

func (w *Worker) HandleJob(job *strategy.Strategy) error {
	w.log.Debugf("job(%d): %s %s %s", w.id, job.Market, job.Instrument, job.Strategy)

	switch strat := job.Strategy; strat {
	case strategy.StrategyAlgo_MeanLog:
		return w.handleMeanLog(job)
	default:
		return fmt.Errorf("unknown strategy %s", strat)
	}
}

func (w *Worker) storeSuggestedAction(action strategy.Action, job *strategy.Strategy) error {
	q := db.Build().Insert(historyTblName).Columns("strategy_id", "action").
		SetMap(sq.Eq{"strategy_id": job.Id, "action": action})

	return db.SimpleExec(context.Background(), q)
}

func (w *Worker) broadcastSuggestedAction(action strategy.Action, job *strategy.Strategy) error {
	b, err := broadcast.Driver()
	if err != nil {
		return err
	}

	ev := &ActionEvent{Action: action, StrategyID: job.Id}

	return b.Publish("STRAT.action", ev)
}
