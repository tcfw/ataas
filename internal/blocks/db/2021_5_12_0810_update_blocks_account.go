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
		"update_blocks_account",
		time.Date(2021, 5, 12, 8, 10, 0, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `
				ALTER TABLE blocks ADD COLUMN account UUID NOT NULL;
				CREATE INDEX ON blocks (account);
			`)
			return err
		},

		//Down
		func(ctx context.Context, tx pgx.Tx) error {
			return fmt.Errorf("down not supported in this migration")
		},
	))
}
