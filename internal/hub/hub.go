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

type Client struct {
	UserID int64
	Conn   *websocket.Conn
}

type Message struct {
	// TODO: Sender
	Payload string
}

type Hub struct {
	logger   *slog.Logger
	tmpl     *template.Template
	clients  *list.List
	messages chan Message
}

func New(lc fx.Lifecycle, logger *slog.Logger, tmpl *template.Template) *Hub {
	h := &Hub{
		logger:   logger,
		tmpl:     tmpl,
		clients:  list.New(),
		messages: make(chan Message),
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
		for cur := h.clients.Front(); cur != nil; cur = cur.Next() {
			client := cur.Value.(*Client)

			if err := client.Conn.WriteMessage(websocket.TextMessage, buf.Bytes()); err != nil {
				h.logger.Warn("failed to write message to connection", "err", err)
				continue
			}
		}
	}
}

func (h *Hub) Join(client *Client) *list.Element {
	return h.clients.PushBack(client)
}

func (h *Hub) Leave(e *list.Element) {
	client := h.clients.Remove(e).(*Client)
	client.Conn.Close()
}

func (h *Hub) Broadcast(msg Message) {
	h.messages <- msg
}
