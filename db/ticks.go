package db

import (
	"context"
	"fmt"
	"time"

	"pm.tcfw.com.au/source/trader/common"
)

func StoreTick(ctx context.Context, t *common.TickerData) error {
	return fmt.Errorf("not implemented")
}

func GetTicks(ctx context.Context, tsRange time.Duration, market string, instrument string) ([]*common.TickerData, error) {
	return nil, fmt.Errorf("not implemented")
}

func GetCandles(ctx context.Context, tsRange time.Duration, interval time.Duration, market string, instrument string) ([]*common.TickerData, error) {
	return nil, fmt.Errorf("not implemented")
}
