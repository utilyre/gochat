package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Static(r *mux.Router) {
	fs := http.FileServer(http.Dir("public"))
	r.PathPrefix("/").Handler(fs)
}
