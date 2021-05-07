package db

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	pool *pgxpool.Pool
)

func Conn(ctx context.Context) (*pgxpool.Conn, error) {
	return pool.Acquire(ctx)
}

func Init(ctx context.Context, config string) (err error) {
	pool, err = pgxpool.Connect(ctx, config)
	return err
}

func Build() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func Exec(ctx context.Context, tx pgx.Tx, stmt squirrel.Sqlizer) (pgconn.CommandTag, error) {
	sql, args, err := stmt.ToSql()
	if err != nil {
		return nil, err
	}

	return tx.Exec(ctx, sql, args...)
}

func Query(ctx context.Context, tx pgx.Tx, stmt squirrel.Sqlizer) (pgx.Rows, error) {
	sql, args, err := stmt.ToSql()
	if err != nil {
		return nil, err
	}

	return tx.Query(ctx, sql, args...)
}

func SimpleExec(ctx context.Context, stmt squirrel.Sqlizer) error {
	conn, err := Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	_, err = Exec(ctx, tx, stmt)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func SimpleQuery(ctx context.Context, stmt squirrel.Sqlizer) (pgx.Rows, func(), error) {
	conn, err := Conn(ctx)
	if err != nil {
		return nil, nil, err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Commit(ctx)

	rows, err := Query(ctx, tx, stmt)
	if err != nil {
		return nil, nil, err
	}

	closer := func() {
		rows.Close()
		conn.Release()
	}

	return rows, closer, nil
}
