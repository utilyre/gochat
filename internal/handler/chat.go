package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/utilyre/gochat/internal/hub"
)

type Message struct {
	Payload string `json:"payload"`
}

type chatHandler struct {
	logger   *slog.Logger
	upgrader *websocket.Upgrader
	hub      *hub.Hub
}

func Chat(r *mux.Router, logger *slog.Logger, upgrader *websocket.Upgrader, hub *hub.Hub) {
	s := r.PathPrefix("/api/chat").Subrouter()
	h := chatHandler{
		logger:   logger,
		upgrader: upgrader,
		hub:      hub,
	}

	s.HandleFunc("", h.chat).Methods(http.MethodGet)
}

func (h chatHandler) chat(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e := h.hub.Subscribe(conn)
	defer h.hub.Unsubscribe(e)

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				return
			}

			h.logger.Warn("failed to read message from connection", "error", err)
			return
		}
		if mt != websocket.TextMessage {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("Unsupported Message Type"))
			return
		}

		msg := new(Message)
		if err := json.Unmarshal(data, msg); err != nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("Invalid JSON Payload"))
			return
		}

		h.hub.Broadcast(hub.Message{
			// TODO: Sender
			Payload: msg.Payload,
		})
	}
}
