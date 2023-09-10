package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Static(r *mux.Router) {
	r.Handle("/", http.FileServer(http.Dir("public")))
}
