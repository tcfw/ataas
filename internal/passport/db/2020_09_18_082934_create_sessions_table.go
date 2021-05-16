package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	migrate "github.com/tcfw/go-migrate/pgx"
)

func init() {
	register(migrate.NewSimpleMigration(
		"create_sessions_tables",
		time.Date(2020, 9, 18, 8, 29, 34, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `CREATE TABLE sessions (
				jti UUID,
				sub UUID,
				exp TIMESTAMPTZ NOT NULL,
				revoked bool DEFAULT false,

				PRIMARY KEY (jti, sub)
			)`)
			return err
		},

		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `DROP TABLE sessions`)
			return err
		},
	))
}
