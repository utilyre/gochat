package websocket

import "github.com/gorilla/websocket"

func NewUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{}
}
