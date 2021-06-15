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
		"update_blocks_col_types",
		time.Date(2021, 6, 17, 15, 32, 0, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			//Break out of tx as of v21.1.1 altering column types cannot be executed in transactions
			conn := tx.Conn()
			tx.Commit(ctx)
			_, err := conn.Exec(ctx, `
			SET enable_experimental_alter_column_type_general = true;
			ALTER TABLE blocks ALTER current_units TYPE DECIMAL(10,6);
			`)
			tx, _ = tx.Begin(ctx)
			return err
		},

		//Down
		func(ctx context.Context, tx pgx.Tx) error {
			return fmt.Errorf("down not supported in this migration")
		},
	))
}
