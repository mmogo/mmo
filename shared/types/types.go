package types

import (
	"github.com/faiface/pixel"
	"github.com/gorilla/websocket"
)

type Player struct {
	ID       string
	Position pixel.Vec
	Conn     *websocket.Conn
}
