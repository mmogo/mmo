package types

import (
	"github.com/faiface/pixel"
	"github.com/gorilla/websocket"
	"image/color"
)

type ServerPlayer struct {
	*Player
	Conn *websocket.Conn
}

type ClientPlayer struct {
	*Player
	Color color.Color
}

type Player struct {
	ID       string
	Position pixel.Vec
}
