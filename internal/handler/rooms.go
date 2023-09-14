package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/utilyre/gochat/internal/auth"
	"github.com/utilyre/gochat/internal/env"
	"github.com/utilyre/gochat/internal/hub"
	"github.com/utilyre/gochat/internal/storage"
	"github.com/utilyre/gochat/pkg/notifier"
)

type Room struct {
	ID   int64  `json:"id" validate:"isdefault"`
	Name string `json:"name" validate:"required,min=3,max=50"`
}

type messageObserver struct {
	conn   *websocket.Conn
	logger *slog.Logger
	tmpl   *template.Template
}

var _ notifier.Observer[*hub.Message] = messageObserver{}

func (o messageObserver) OnNotify(msg *hub.Message) {
	buf := new(bytes.Buffer)
	if err := o.tmpl.ExecuteTemplate(buf, "message", msg); err != nil {
		o.logger.Warn("failed to execute template 'message'", "data", msg)
		return
	}

	if err := o.conn.WriteMessage(websocket.TextMessage, buf.Bytes()); err != nil {
		o.logger.Warn("failed to write message to connection", "error", err)
	}
}

type roomsHandler struct {
	env      env.Env
	validate *validator.Validate
	logger   *slog.Logger
	tmpl     *template.Template
	storage  storage.RoomsStorage
	upgrader *websocket.Upgrader
	hub      *hub.Hub
}

// create new room on demand
// remove any room that has no participants

func Rooms(
	r *mux.Router,
	env env.Env,
	validate *validator.Validate,
	logger *slog.Logger,
	tmpl *template.Template,
	storage storage.RoomsStorage,
	upgrader *websocket.Upgrader,
	hub *hub.Hub,
) {
	s := r.PathPrefix("/api/rooms").Subrouter()
	h := roomsHandler{
		env:      env,
		validate: validate,
		logger:   logger,
		tmpl:     tmpl,
		storage:  storage,
		upgrader: upgrader,
		hub:      hub,
	}

	s.HandleFunc("", h.create).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	s.HandleFunc("", h.readAll).
		Methods(http.MethodGet)

	/*
		s.HandleFunc("/{id:[0-9]+}", h.read).
			Methods(http.MethodGet)
	*/

	s.HandleFunc("/{id:[0-9]+}/chat", h.chat).
		Methods(http.MethodGet)
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

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if err := h.tmpl.ExecuteTemplate(w, "rooms", dbRooms); err != nil {
		h.logger.Warn("failed to write body to response", "error", err)
	}
}

func (h roomsHandler) chat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Warn("failed to convert id URL parameter to int64", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("failed to upgrade protocol", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	o := messageObserver{
		conn:   conn,
		logger: h.logger,
		tmpl:   h.tmpl,
	}

	e := h.hub.Join(id, o)
	defer func() { _ = h.hub.Leave(id, e) }()

	data := readMessage(conn, h.logger)
	claims, err := auth.Verify(h.env.BESecret, string(data))
	if err != nil {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
			h.logger.Warn("failed to write message to connection", "error", err)
		}

		return
	}

	h.logger.Info("user entered chat room", "room-id", id, "user-email", claims.Email)

	for {
		data := readMessage(conn, h.logger)

		msg := new(hub.Message)
		if err := json.Unmarshal(data, msg); err != nil {
			if err := conn.WriteMessage(websocket.TextMessage, []byte("Invalid JSON Payload")); err != nil {
				h.logger.Warn("failed to write message to connection", "error", err)
			}

			return
		}

		msg.Sender = claims.Email
		if err := h.hub.Send(id, msg); err != nil {
			h.logger.Warn("failed to send message to room", "error", err)
		}
	}
}

func readMessage(conn *websocket.Conn, logger *slog.Logger) []byte {
	mt, data, err := conn.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return nil
		}

		logger.Warn("failed to read message from connection", "error", err)
		return nil
	}
	if mt != websocket.TextMessage {
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Unsupported Message Type")); err != nil {
			logger.Warn("failed to write message to connection", "error", err)
		}

		return nil
	}

	return data
}
