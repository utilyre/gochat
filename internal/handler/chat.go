package handler

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Message struct {
	Name    string `json:"name"`
	Payload string `json:"payload"`
}

type chatHandler struct {
	logger   *slog.Logger
	tmpl     *template.Template
	upgrader *websocket.Upgrader
}

func Chat(r *mux.Router, logger *slog.Logger, tmpl *template.Template, upgrader *websocket.Upgrader) {
	h := chatHandler{
		logger:   logger,
		tmpl:     tmpl,
		upgrader: upgrader,
	}

	r.HandleFunc("/chat", h.chat)
}

func (h chatHandler) chat(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				return
			}

			h.logger.Warn("failed to read message from connection", "err", err)
			return
		}
		if mt != websocket.TextMessage {
			conn.WriteMessage(websocket.TextMessage, []byte("Unsupported Message Type"))
			return
		}

		msg := new(Message)
		if err := json.Unmarshal(data, msg); err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid JSON Payload"))
			return
		}

		buf := new(bytes.Buffer)
		if err := h.tmpl.ExecuteTemplate(buf, "message", msg); err != nil {
			h.logger.Warn("failed to execute template", "name", "message", "data", msg)
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, buf.Bytes()); err != nil {
			h.logger.Warn("failed to write message to connection", "err", err)
			return
		}
	}
}
