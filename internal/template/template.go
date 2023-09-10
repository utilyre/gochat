package template

import (
	"html/template"
	"log/slog"
)

func New(logger *slog.Logger) *template.Template {
	tmpl, err := template.ParseGlob("views/*.html")
	if err != nil {
		logger.Error("failed to parse templates", "err", err)
	}

	return tmpl
}
