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
		"create_strategy_table",
		time.Date(2021, 5, 1, 12, 53, 34, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS strategies (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					market STRING NOT NULL,
					instrument STRING NOT NULL,
					strategy INT NOT NULL,
					params JSONB,
					duration INT NOT NULL,
					next TIMESTAMPTZ NOT NULL,
					INDEX strategy_next (next ASC)
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
