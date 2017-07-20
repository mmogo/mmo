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
	AddPlayer         *AddPlayer         `,omitempty`
	PlayerDestination *PlayerDestination `,omitempty`
	PlayerPosition    *PlayerPosition    `,omitempty`
	PlayerSpoke       *PlayerSpoke       `,omitempty`
	WorldState        *WorldState        `,omitempty`
	RemovePlayer      *RemovePlayer      `,omitempty`
	Processed         time.Time
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

type PlayerDestination struct {
	ID          string
	Destination pixel.Vec
}

type PlayerPosition struct {
	ID       string
	Position pixel.Vec
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
	if u.AddPlayer != nil {
		return fmt.Sprintf("AddPlayer: %s", u.AddPlayer.ID)
	}

	if u.PlayerPosition != nil {
		return fmt.Sprintf("PlayerPosition: %s: %s", u.PlayerPosition.ID, u.PlayerPosition.Position)
	}

	if u.PlayerDestination != nil {
		return fmt.Sprintf("PlayerDestination: %s: %s", u.PlayerDestination.ID, u.PlayerDestination.Destination)
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
