package handler

import (
	"errors"
	"io/fs"
	"net/http"

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

func Static(r *mux.Router) {
	fs := http.FileServer(HTMLDir{
		dir: http.Dir("public"),
	})

	r.PathPrefix("/").Handler(fs)
}
