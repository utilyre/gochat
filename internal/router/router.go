package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

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
				err := srv.ListenAndServe()
				logger.Info("started server", "address", srv.Addr)

				if err != nil {
					logger.Error("failed to listen and serve", "err", err)
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
