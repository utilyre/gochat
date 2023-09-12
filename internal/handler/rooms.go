package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/utilyre/gochat/internal/storage"
)

type Room struct {
	ID   int64  `json:"id" validate:"isdefault"`
	Name string `json:"name" validate:"required,min=3,max=50"`
}

type roomsHandler struct {
	validate *validator.Validate
	logger   *slog.Logger
	tmpl     *template.Template
	storage  storage.RoomsStorage
}

func Rooms(
	r *mux.Router,
	validate *validator.Validate,
	logger *slog.Logger,
	tmpl *template.Template,
	storage storage.RoomsStorage,
) {
	s := r.PathPrefix("/api/rooms").Subrouter()
	h := roomsHandler{
		validate: validate,
		logger:   logger,
		tmpl:     tmpl,
		storage:  storage,
	}

	s.HandleFunc("", h.create).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	s.HandleFunc("", h.readAll).
		Methods(http.MethodGet)

	/*
		s.HandleFunc("/{id:[0-9]+}", h.read).
			Methods(http.MethodGet)

		s.HandleFunc("/{id:[0-9]+}/chat", h.chat).
			Methods(http.MethodGet)
	*/
}

func (h roomsHandler) create(w http.ResponseWriter, r *http.Request) {
	room := new(Room)
	if err := json.NewDecoder(r.Body).Decode(room); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(room); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbRoom := &storage.Room{
		Name: room.Name,
	}

	if err := h.storage.Create(dbRoom); err != nil {
		switch {
		case errors.Is(err, storage.ErrDuplicateKey):
			http.Error(w, "room already exists", http.StatusConflict)
		default:
			h.logger.Warn("failed to create room in rooms table", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	room.ID = dbRoom.ID

	body, err := json.Marshal(room)
	if err != nil {
		h.logger.Warn("failed to marshal response body", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(body); err != nil {
		h.logger.Warn("failed to write body to response", "error", err)
	}
}

func (h roomsHandler) readAll(w http.ResponseWriter, r *http.Request) {
	dbRooms := []storage.Room{}
	if err := h.storage.ReadAll(&dbRooms); err != nil {
		h.logger.Warn("failed to read all rooms from database", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	if err := h.tmpl.ExecuteTemplate(buf, "rooms", dbRooms); err != nil {
		h.logger.Warn("failed to execute 'rooms' template", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
