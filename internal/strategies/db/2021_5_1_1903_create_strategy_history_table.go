package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	migrate "github.com/tcfw/go-migrate/pgx"
)

func init() {
	register(migrate.NewSimpleMigration(
		"create_strategy_history_table",
		time.Date(2021, 5, 1, 12, 53, 34, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS strategy_history (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					strategy_id UUID NOT NULL,
					action INT NOT NULL,
					ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					INDEX strategy_id (strategy_id ASC)
				)
			`)
			return err
		},

		//Down
		func(ctx context.Context, tx pgx.Tx) error {
			return fmt.Errorf("down not supported in this migration")
		},
	))
}
