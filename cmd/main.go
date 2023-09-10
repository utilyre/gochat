package main

import (
	"github.com/utilyre/gochat/internal/handler"
	"github.com/utilyre/gochat/internal/hub"
	"github.com/utilyre/gochat/internal/logger"
	"github.com/utilyre/gochat/internal/router"
	"github.com/utilyre/gochat/internal/template"
	"github.com/utilyre/gochat/internal/validator"
	"github.com/utilyre/gochat/internal/websocket"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			logger.New,
			template.New,
			validator.New,
			hub.New,
			router.New,
			websocket.NewUpgrader,
		),
		fx.Invoke(
			handler.Static,
			handler.Chat,
		),
	).Run()
}
