package template

import (
	"html/template"
	"log/slog"
	"os"
)

func New(logger *slog.Logger) *template.Template {
	tmpl, err := template.ParseGlob("views/*.html")
	if err != nil {
		logger.Error("failed to parse templates", "error", err)
		os.Exit(1)
	}

	return tmpl
}
