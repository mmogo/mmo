package shared

import (
	"github.com/faiface/pixel"
	"github.com/ilackarms/_anything/shared/types"
)

type Message struct {
	ConnectRequest *ConnectRequest
	MoveRequest    *MoveRequest

	PlayerMoved        *PlayerMoved
	WorldState         *WorldState
	PlayerDisconnected *PlayerDisconnected
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

type WorldState struct {
	Players []*types.Player
}

type PlayerDisconnected struct {
	ID string
}
