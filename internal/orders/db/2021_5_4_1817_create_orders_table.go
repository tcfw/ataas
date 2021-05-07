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
		"create_orders_table",
		time.Date(2021, 5, 5, 18, 17, 43, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS orders (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					block_id UUID NOT NULL,
					side BOOL NOT NULL,
					price FLOAT NOT NULL,
					quantity FLOAT NOT NULL,
					ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					INDEX blocks (block_id)
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
