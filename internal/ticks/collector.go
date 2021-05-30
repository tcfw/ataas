package ticks

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"pm.tcfw.com.au/source/ataas/api/pb/ticks"
	"pm.tcfw.com.au/source/ataas/db"
	"pm.tcfw.com.au/source/ataas/internal/broadcast"
	binance_client "pm.tcfw.com.au/source/ataas/internal/exchanges/binance-client"
	crypto_com_client "pm.tcfw.com.au/source/ataas/internal/exchanges/crypto-com-client"
)

func (s *Server) Collect(ctx context.Context) {
	ch := make(chan *ticks.Trade, 100)

	go s.collectFromCh(ctx, ch)

	//crypto.com
	// go func() {
	// 	err := s.collectCryptoDotCom(ctx, ch)
	// 	if err != nil {
	// 		s.log.Fatalf("Disconnected from crypto.com: %s", err)
	// 		os.Exit(1)
	// 	}
	// }()

	//binance.com
	go func() {
		err := s.collectBinanceDotCom(ctx, ch)
		if err != nil {
			s.log.Fatalf("Disconnected from binance.com: %s", err)
			os.Exit(1)
		}
	}()

	s.gcTrades()
}

func (s *Server) collectFromCh(ctx context.Context, ch <-chan *ticks.Trade) {
	br, err := broadcast.Driver()
	if err != nil {
		panic(err)
	}

	for trade := range ch {
		br.Publish(fmt.Sprintf("TRADE.%s.%s", trade.Market, trade.Instrument), trade)

		if err := s.library.Add(trade); err != nil {
			s.log.Errorf("failed to record in library: %s", err)
		}
	}
}

func (s *Server) gcTrades() {
	ctx := context.Background()

	conn, err := db.Conn(ctx)
	if err != nil {
		panic(err)
	}

	t := time.NewTicker(10 * time.Minute)

	for range t.C {
		_, err := conn.Exec(ctx, "DELETE FROM trades WHERE ts < $1", time.Now().Add(-48*time.Hour))
		if err != nil {
			panic(err)
		}
	}
}

func (s *Server) collectCryptoDotCom(ctx context.Context, ch chan *ticks.Trade) error {
	key := viper.GetString("collector.crypto_com.key")
	secret := viper.GetString("collector.crypto_com.secret")

	c := crypto_com_client.NewClient(key, secret)

	tch, err := c.SubscribeTradesAll()
	if err != nil {
		return err
	}

	s.log.Info("Collecting trades from crypto.com")

	for t := range tch {
		ch <- t
	}

	return nil
}

func (s *Server) collectBinanceDotCom(ctx context.Context, ch chan *ticks.Trade) error {
	key := viper.GetString("collector.binance.key")
	secret := viper.GetString("collector.binance.secret")

	c := binance_client.NewClient(key, secret)

	tch, err := c.SubscribeTradesAll()
	if err != nil {
		return err
	}

	s.log.Info("Collecting trades from binance.com")

	for t := range tch {
		ch <- t
	}

	return nil
}
