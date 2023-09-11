package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/utilyre/gochat/internal/env"
	"go.uber.org/fx"
)

func New(lc fx.Lifecycle, env env.Env, logger *slog.Logger) *sqlx.DB {
	dsn := fmt.Sprintf(
		"user='%s' password='%s' host='%s' port='%s' sslmode=disable",
		env.DBUser, env.DBPass, env.DBHost, env.DBPort,
	)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		logger.Error("failed to open connection to database", "error", err)
		os.Exit(1)
	}
	logger.Info("database connection has been opened")

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return db.PingContext(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return db.Close()
		},
	})

	return db
}
