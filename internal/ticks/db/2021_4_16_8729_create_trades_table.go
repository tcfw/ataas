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
		"create_trades_table",
		time.Date(2021, 4, 16, 8, 7, 29, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `
				SET experimental_enable_hash_sharded_indexes=on;
				CREATE TABLE IF NOT EXISTS trades (
					market STRING NOT NULL,
					instrument STRING NOT NULL,
					tradeid STRING NOT NULL,
					direction BOOL NOT NULL,
					amount FLOAT8 NOT NULL,
					units FLOAT8 NOT NULL,
					ts TIMESTAMPTZ NOT NULL,
					PRIMARY KEY (market, instrument, tradeid),
					INDEX trades_ts_idx (ts DESC) USING HASH WITH BUCKET_COUNT=24
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
