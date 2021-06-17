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
		"update_order_col_types",
		time.Date(2021, 6, 15, 15, 23, 3, 0, time.Local),

		//Up
		func(ctx context.Context, tx pgx.Tx) error {
			//Break out of tx as of v21.1.1 altering column types cannot be executed in transactions
			conn := tx.Conn()
			tx.Commit(ctx)
			_, err := conn.Exec(ctx, `
			SET enable_experimental_alter_column_type_general = true;
			ALTER TABLE orders ALTER price TYPE INT64 USING (price*1000000)::INT64;
			ALTER TABLE orders ALTER quantity TYPE INT64 USING (quantity*1000000)::INT64;
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
