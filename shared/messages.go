package shared

import (
	"github.com/faiface/pixel"
)

type Message struct {
	ConnectRequest *ConnectRequest
	MoveRequest    *MoveRequest

	PlayerMoved *PlayerMoved
}

type ConnectRequest struct {
	ID string
}

type MoveRequest struct {
	Direction pixel.Vec
}

type PlayerMoved struct {
	ID          string
	NewPosition pixel.Vec
}
