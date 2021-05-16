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
		"create_users_table",
		time.Date(2021, 5, 12, 8, 24, 0, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS users (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					status INT NOT NULL DEFAULT 0,
					email STRING NOT NULL,
					firstName STRING NOT NULL,
					lastName STRING,
					createdAt TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					updatedAt TIMESTAMPTZ NOT NULL DEFAULT NOW(),
					deletedAt TIMESTAMPTZ,
					mfa JSONB,
					password STRING,
					metadata JSONB,
					account UUID NOT NULL
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
