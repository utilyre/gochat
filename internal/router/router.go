package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/utilyre/gochat/internal/env"
	"go.uber.org/fx"
)

func New(lc fx.Lifecycle, env env.Env, logger *slog.Logger) *mux.Router {
	r := mux.NewRouter()
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", env.BEPort),
		Handler: r,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil {
					logger.Error("failed to listen and serve", "address", srv.Addr, "error", err)
					os.Exit(1)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return r
}
