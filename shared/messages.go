package shared

import (
	"time"

	"github.com/faiface/pixel"
)

type Message struct {
	ConnectRequest *ConnectRequest
	MoveRequest    *MoveRequest
	SpeakRequest   *SpeakRequest

	PlayerMoved        *PlayerMoved
	PlayerSpoke        *PlayerSpoke
	WorldState         *WorldState
	PlayerDisconnected *PlayerDisconnected
}

type ConnectRequest struct {
	ID string
}

type MoveRequest struct {
	Direction Direction
	Created   time.Time
}

type SpeakRequest struct {
	Text string
}

type PlayerMoved struct {
	ID          string
	NewPosition pixel.Vec
	RequestTime time.Time
}

type PlayerSpoke struct {
	ID   string
	Text string
}

type WorldState struct {
	Players []*Player
}

type PlayerDisconnected struct {
	ID string
}
