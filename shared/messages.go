package shared

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
)

type Message struct {
	Request *Request `,omitempty`
	Update  *Update  `,omitempty`
	Error   *Error   `,omitempty`
}

type Update struct {
	AddPlayer    *AddPlayer    `,omitempty`
	PlayerMoved  *PlayerMoved  `,omitempty`
	PlayerSpoke  *PlayerSpoke  `,omitempty`
	WorldState   *WorldState   `,omitempty`
	RemovePlayer *RemovePlayer `,omitempty`
}

type Request struct {
	ConnectRequest *ConnectRequest `,omitempty`
	MoveRequest    *MoveRequest    `,omitempty`
	SpeakRequest   *SpeakRequest   `,omitempty`
}

type Error struct {
	Message string
}

type ConnectRequest struct {
	ID string
}

type MoveRequest struct {
	Destination pixel.Vec
}

type SpeakRequest struct {
	Text string
}

type AddPlayer struct {
	ID       string
	Position pixel.Vec
}

type PlayerMoved struct {
	ID          string
	Destination pixel.Vec
	RequestTime time.Time
}

type PlayerSpoke struct {
	ID   string
	Text string
}

type WorldState struct {
	World *World
}

type RemovePlayer struct {
	ID string
}

func (m Message) String() string {
	if m.Error != nil {
		return fmt.Sprintf("Error: %s", m.Error.Message)
	}
	if m.Request != nil {
		return m.Request.String()
	}

	if m.Update != nil {
		return m.Update.String()
	}

	return "empty packet"
}

func (u Update) String() string {
	if u.PlayerMoved != nil {
		return fmt.Sprintf("PlayerMoved: %s: %s", u.PlayerMoved.ID, u.PlayerMoved.Destination)
	}

	if u.PlayerSpoke != nil {
		return fmt.Sprintf("PlayerSpoke: %s: %s", u.PlayerSpoke.ID, u.PlayerSpoke.Text)
	}

	if u.WorldState != nil {
		return fmt.Sprintf("WorldState: %#v", u.WorldState.World)
	}
	if u.RemovePlayer != nil {
		return fmt.Sprintf("PlayerDisconnected: %s", u.RemovePlayer)
	}

	return "empty update"
}

func (r Request) String() string {
	if r.ConnectRequest != nil {
		return fmt.Sprintf("ConnectRequest: %v", r.ConnectRequest.ID)
	}
	if r.MoveRequest != nil {
		return fmt.Sprintf("MoveRequest: %s", r.MoveRequest.Destination)
	}
	if r.SpeakRequest != nil {
		return fmt.Sprintf("SpeakRequest: %s", r.SpeakRequest.Text)
	}

	return "empty request"
}
