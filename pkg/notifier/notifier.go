package notifier

import (
	"container/list"
	"context"
)

type (
	Observer[T any] interface {
		OnNotify(T)
	}

	Notifier[T any] interface {
		Notify(T)
		Register(Observer[T]) *list.Element
		Deregister(*list.Element) Observer[T]
		Len() int
		Listen() error
		Shutdown()
	}
)

type notifier[T any] struct {
	cancel context.CancelFunc

	observers *list.List
	events    chan T
}

func New[T any]() Notifier[T] {
	return &notifier[T]{
		observers: list.New(),
		events:    make(chan T),
	}
}

func (n *notifier[T]) Notify(e T) {
	n.events <- e
}

func (n *notifier[T]) Register(s Observer[T]) *list.Element {
	return n.observers.PushBack(s)
}

func (n *notifier[T]) Deregister(e *list.Element) Observer[T] {
	return n.observers.Remove(e).(Observer[T])
}

func (n *notifier[T]) Len() int {
	return n.observers.Len()
}

func (n *notifier[T]) broadcast(e T) {
	for cur := n.observers.Front(); cur != nil; cur = cur.Next() {
		o := cur.Value.(Observer[T])
		o.OnNotify(e)
	}
}

func (n *notifier[T]) Listen() error {
	ctx, cancel := context.WithCancel(context.Background())
	n.cancel = cancel

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-n.events:
			go n.broadcast(e)
		}
	}
}

func (n *notifier[T]) Shutdown() {
	n.cancel()
}
