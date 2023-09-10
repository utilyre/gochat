package hub

import (
	"bytes"
	"container/list"
	"context"
	"html/template"
	"log/slog"

	"github.com/gorilla/websocket"
	"go.uber.org/fx"
)

type Message struct {
	// TODO: Sender
	Payload string
}

type Hub struct {
	logger      *slog.Logger
	tmpl        *template.Template
	subscribers *list.List
	messages    chan Message
}

func New(lc fx.Lifecycle, logger *slog.Logger, tmpl *template.Template) *Hub {
	h := &Hub{
		logger:      logger,
		tmpl:        tmpl,
		subscribers: list.New(),
		messages:    make(chan Message),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go h.Start()
			return nil
		},
	})

	return h
}

func (h *Hub) Start() {
	for msg := range h.messages {
		data := map[string]any{
			"Name":    "TODO",
			"Message": msg.Payload,
		}

		buf := new(bytes.Buffer)
		if err := h.tmpl.ExecuteTemplate(buf, "message", data); err != nil {
			h.logger.Warn("failed to execute template 'message'", "data", data)
			continue
		}
		for cur := h.subscribers.Front(); cur != nil; cur = cur.Next() {
			conn := cur.Value.(*websocket.Conn)

			if err := conn.WriteMessage(websocket.TextMessage, buf.Bytes()); err != nil {
				h.logger.Warn("failed to write message to connection", "err", err)
				continue
			}
		}
	}
}

func (h *Hub) Subscribe(conn *websocket.Conn) *list.Element {
	return h.subscribers.PushBack(conn)
}

func (h *Hub) Unsubscribe(e *list.Element) {
	conn := h.subscribers.Remove(e).(*websocket.Conn)
	conn.Close()
}

func (h *Hub) Broadcast(msg Message) {
	h.messages <- msg
}
