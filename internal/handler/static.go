package handler

import (
	"errors"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type HTMLDir struct {
	dir http.Dir
}

func (d HTMLDir) Open(name string) (http.File, error) {
	f, err := d.dir.Open(name)
	if err != nil {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return d.dir.Open(name + ".html")
		default:
			return nil, err
		}
	}

	return f, nil
}

type staticHandler struct {
	logger *slog.Logger
	tmpl   *template.Template
}

func Static(
	r *mux.Router,
	logger *slog.Logger,
	tmpl *template.Template,
) {
	h := staticHandler{
		logger: logger,
		tmpl:   tmpl,
	}

	r.HandleFunc("/chat/{id:[0-9]+}", h.chat).
		Methods(http.MethodGet)

	r.PathPrefix("/").
		Handler(http.FileServer(HTMLDir{dir: http.Dir("public")}))
}

func (h staticHandler) chat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Warn("failed to convert id URL parameter to int64", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if err := h.tmpl.ExecuteTemplate(w, "chat", id); err != nil {
		h.logger.Warn("failed to write body to response", "error", err)
	}
}
