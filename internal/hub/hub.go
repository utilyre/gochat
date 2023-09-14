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

func (h *Hub) Join(id int64, o notifier.Observer[*Message]) *list.Element {
	room, ok := h.rooms[id]
	if !ok {
		r := notifier.New[*Message]()
		go func() { _ = r.Listen() }()

		h.rooms[id] = r
		room = r

		h.logger.Info("created a new room", "id", id)
	}

	return room.Register(o)
}

func (h *Hub) Leave(id int64, e *list.Element) error {
	room, ok := h.rooms[id]
	if !ok {
		return ErrRoomNotFound
	}

	room.Deregister(e)
	if room.Len() == 0 {
		delete(h.rooms, id)
		room.Shutdown()

		h.logger.Info("deleted room", "id", id)
	}

	return nil
}

func (h *Hub) Send(id int64, msg *Message) error {
	room, ok := h.rooms[id]
	if !ok {
		return ErrRoomNotFound
	}

	room.Notify(msg)
	return nil
}
