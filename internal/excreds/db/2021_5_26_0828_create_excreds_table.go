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
		"create_excreds_table",
		time.Date(2021, 5, 28, 8, 28, 0, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS excreds (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					account UUID NOT NULL,
					exchange STRING NOT NULL,
					key STRING NOT NULL,
					secret STRING NOT NULL,
					createdAt TIMESTAMPTZ NOT NULL DEFAULT now(),
					INDEX accountidx (account)
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
