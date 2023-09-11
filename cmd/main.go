package main

import (
	"github.com/utilyre/gochat/internal/database"
	"github.com/utilyre/gochat/internal/env"
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
			env.New,
			logger.New,
			template.New,
			database.New,
			validator.New,
			hub.New,
			router.New,
			websocket.NewUpgrader,
		),
		fx.Invoke(
			handler.Chat,
			handler.Static,
		),
	).Run()
}
