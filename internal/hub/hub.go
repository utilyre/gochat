package hub

import (
	"context"
	"html/template"
	"log/slog"
	"os"

	"github.com/utilyre/gochat/pkg/notifier"
	"go.uber.org/fx"
)

type Message struct {
	Sender  string `json:"-"`
	Payload string `json:"payload"`
}

type Hub struct {
	notifier.Notifier[*Message]

	logger *slog.Logger
	tmpl   *template.Template
}

func New(lc fx.Lifecycle, logger *slog.Logger, tmpl *template.Template) *Hub {
	h := &Hub{
		Notifier: notifier.New[*Message](),
		logger:   logger,
		tmpl:     tmpl,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := h.Listen(); err != nil {
					logger.Error("failed to listen for messages", "error", err)
					os.Exit(1)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			h.Shutdown()
			return nil
		},
	})

	return h
}
