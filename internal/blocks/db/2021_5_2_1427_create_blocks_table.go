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
		"create_blocks_table",
		time.Date(2021, 5, 1, 12, 53, 34, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS blocks (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					strategy_id UUID NOT NULL,
					state int8 NOT NULL DEFAULT 0,
					base_units float NOT NULL,
					current_units float NOT NULL,
					purchase float NOT NULL,
					watch_duration int64 NOT NULL DEFAULT 0,
					short_sell_allowed bool NOT NULL DEFAULT false,
					backout_percentage float NOT NULL DEFAULT 0.05,
					INDEX strategy (strategy_id)
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
