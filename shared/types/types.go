package types

import (
	"github.com/faiface/pixel"
	"github.com/gorilla/websocket"
)

type ServerPlayer struct {
	*Player
	Conn *websocket.Conn
}

type Player struct {
	ID       string
	Position pixel.Vec
}
