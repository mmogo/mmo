// Copyright 2017 The MMOGO Authors. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package shared

// Action describes the activity of an entity
type Action int

const (
	A_IDLE Action = iota
	A_WALK
	A_CHAT_
	A_SLASH
	A_SHOOT
	A_SPELL
	A_THRUST
	A_HURT
	A_DEAD
)
