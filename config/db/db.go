package db

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustConnect(ctx context.Context, url string) *pgxpool.Pool {
	conn, err := pgxpool.New(ctx, url)
	if err != nil {
		slog.Error("cannot connect to database", "err", err.Error())
		os.Exit(1)
	}
	slog.Info("connected to database")
	return conn
}
