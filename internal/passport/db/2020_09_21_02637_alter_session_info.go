package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	migrate "github.com/tcfw/go-migrate/pgx"
)

func init() {
	register(migrate.NewSimpleMigration(
		"alter_session_info",
		time.Date(2020, 9, 21, 0, 26, 37, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `ALTER TABLE sessions
			ADD COLUMN ip STRING,
			ADD COLUMN ua STRING
			`)
			return err
		},

		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `ALTER TABLE sessions DROP COLUMN ip, DROP COLUMN ua`)
			return err
		},
	))
}
