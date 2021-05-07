package db

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	migrate "github.com/tcfw/go-migrate/pgx"
)

var migs []migrate.Migration = []migrate.Migration{}

func register(mig migrate.Migration) {
	migs = append(migs, mig)
}

//Migrate runs migrations up
func Migrate(ctx context.Context, conn *pgx.Conn, log *logrus.Logger) error {
	return migrate.Migrate(ctx, conn, log, migs)
}
