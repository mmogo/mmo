package shared

import (
	"github.com/faiface/pixel"
	"image/color"
	"net"
	"sync"
)

type ServerPlayer struct {
	*Player
	Conn         net.Conn
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
