// Copyright 2017 The MMOGO Authors. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package shared

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
)

type Message struct {
	Sent    time.Time `,omitempty`
	Request *Request  `,omitempty`
	Update  *Update   `,omitempty`
	Error   *Error    `,omitempty`
}

type Update struct {
	PlayerMoved        *PlayerMoved        `,omitempty`
	PlayerSpoke        *PlayerSpoke        `,omitempty`
	WorldState         *WorldState         `,omitempty`
	PlayerDisconnected *PlayerDisconnected `,omitempty`
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
	Direction pixel.Vec
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
	if m.Error != nil {
		return fmt.Sprintf("Error: %s", m.Error.Message)
	}
	if m.Request != nil {
		return m.Request.String()
	}

	if m.Update != nil {
		return m.Update.String()
	}

	if !m.Sent.IsZero() {
		return fmt.Sprintf("Ping: %s", m.Sent)
	}

	return "empty packet"
}

func (u Update) String() string {
	if u.PlayerMoved != nil {
		return fmt.Sprintf("PlayerMoved: %s: %s", u.PlayerMoved.ID, u.PlayerMoved.NewPosition)
	}

	if u.PlayerSpoke != nil {
		return fmt.Sprintf("PlayerSpoke: %s: %s", u.PlayerSpoke.ID, u.PlayerSpoke.Text)
	}

	if u.WorldState != nil {

		return fmt.Sprintf("WorldState: %v players", len(u.WorldState.Players))
	}
	if u.PlayerDisconnected != nil {
		return fmt.Sprintf("PlayerDisconnected: %s", u.PlayerDisconnected)
	}

	return "empty update"

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
