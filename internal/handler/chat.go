package handler

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/utilyre/gochat/internal/hub"
	"github.com/utilyre/gochat/pkg/notifier"
)

type observer struct {
	*websocket.Conn

	logger *slog.Logger
	tmpl   *template.Template
}

var _ notifier.Observer[hub.Message] = observer{}

func (o observer) OnNotify(msg hub.Message) {
	buf := new(bytes.Buffer)
	if err := o.tmpl.ExecuteTemplate(buf, "message", msg); err != nil {
		o.logger.Warn("failed to execute template 'message'", "data", msg)
		return
	}

	if err := o.WriteMessage(websocket.TextMessage, buf.Bytes()); err != nil {
		o.logger.Warn("failed to write message to connection", "error", err)
	}
}

type chatHandler struct {
	logger   *slog.Logger
	tmpl     *template.Template
	upgrader *websocket.Upgrader
	hub      *hub.Hub
}

func Chat(
	r *mux.Router,
	logger *slog.Logger,
	tmpl *template.Template,
	upgrader *websocket.Upgrader,
	hub *hub.Hub,
) {
	s := r.PathPrefix("/api/chat").Subrouter()
	h := chatHandler{
		logger:   logger,
		tmpl:     tmpl,
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

	e := h.hub.Register(observer{
		Conn:   conn,
		logger: h.logger,
		tmpl:   h.tmpl,
	})
	defer h.hub.Deregister(e)

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

		msg := new(hub.Message)
		if err := json.Unmarshal(data, msg); err != nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("Invalid JSON Payload"))
			return
		}

		h.hub.Notify(hub.Message{
			Sender: "TODO",
			Payload: msg.Payload,
		})
	}
}
