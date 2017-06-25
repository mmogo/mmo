package shared

import (
	"github.com/faiface/pixel"
	"github.com/gorilla/websocket"
	"image/color"
	"sync"
)

type ServerPlayer struct {
	*Player
	Conn         *websocket.Conn
	RequestQueue []*Message
	QueueLock    sync.RWMutex
}

type ClientPlayer struct {
	*Player
	Color color.Color
}

type Player struct {
	ID       string
	Position pixel.Vec
}
