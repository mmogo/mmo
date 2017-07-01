package shared

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
)

type Message struct {
	Sent               time.Time
	Request            *Request
	PlayerMoved        *PlayerMoved
	PlayerSpoke        *PlayerSpoke
	WorldState         *WorldState
	PlayerDisconnected *PlayerDisconnected
}

type Request struct {
	ConnectRequest *ConnectRequest
	MoveRequest    *MoveRequest
	SpeakRequest   *SpeakRequest
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

func (m Message) String() string {
	if m.Request != nil {
		return m.Request.String()
	}

	if m.PlayerMoved != nil {
		return fmt.Sprintf("PlayerMoved: %s: %s", m.PlayerMoved.ID, m.PlayerMoved.NewPosition)
	}

	if m.PlayerSpoke != nil {
		return fmt.Sprintf("PlayerSpoke: %s: %s", m.PlayerSpoke.ID, m.PlayerSpoke.Text)
	}

	if m.WorldState != nil {

		return fmt.Sprintf("WorldState: %v players", len(m.WorldState.Players))
	}
	if m.PlayerDisconnected != nil {
		return fmt.Sprintf("PlayerDisconnected: %s", m.PlayerDisconnected)
	}

	if !m.Sent.IsZero() {

		return fmt.Sprintf("Ping: %s", m.Sent)
	}

	return "empty packet"
}

func (r Request) String() string {
	if r.ConnectRequest != nil {
		return fmt.Sprintf("ConnectRequest: %v", r.ConnectRequest.ID)
	}
	if r.MoveRequest != nil {
		return fmt.Sprintf("MoveRequest: %s", r.MoveRequest.Direction)
	}
	if r.SpeakRequest != nil {
		return fmt.Sprintf("SpeakRequest: %s", r.SpeakRequest.Text)
	}

	return "empty request"
}
