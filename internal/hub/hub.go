package hub

import (
	"container/list"
	"errors"
	"html/template"
	"log/slog"

	"github.com/utilyre/gochat/pkg/notifier"
	"go.uber.org/fx"
)

var (
	ErrRoomNotFound = errors.New("room not found")
)

type Message struct {
	Sender  string `json:"-"`
	Payload string `json:"payload"`
}

type Hub struct {
	logger *slog.Logger
	tmpl   *template.Template

	rooms map[int64]notifier.Notifier[*Message]
}

func New(lc fx.Lifecycle, logger *slog.Logger, tmpl *template.Template) *Hub {
	h := &Hub{
		logger: logger,
		tmpl:   tmpl,

		rooms: make(map[int64]notifier.Notifier[*Message]),
	}

	return h
}

func (h *Hub) Join(o notifier.Observer[*Message], id int64) *list.Element {
	room, ok := h.rooms[id]
	if !ok {
		room := notifier.New[*Message]()
		go func() { _ = room.Listen() }()

		h.rooms[id] = room
	}

	return room.Register(o)
}

func (h *Hub) Leave(e *list.Element, id int64) error {
	room, ok := h.rooms[id]
	if !ok {
		return ErrRoomNotFound
	}

	room.Deregister(e)
	if room.Len() == 0 {
		delete(h.rooms, id)
		room.Shutdown()
	}

	return nil
}
