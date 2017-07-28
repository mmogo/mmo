// Copyright 2017 The MMOGO Authors. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package shared

import (
	"fmt"
	"image/color"
	"net"
	"strings"
	"sync"

	"github.com/faiface/pixel"
)

const fatalErrSig = "**FATAL_ERR**"

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

type fatalError struct {
	err error
}

func FatalErr(err error) error {
	return &fatalError{err: err}
}

func (e *fatalError) Error() string {
	return fmt.Sprintf("%s: %v", fatalErrSig, e.err)
}

func IsFatal(err error) bool {
	return err != nil && strings.Contains(err.Error(), fatalErrSig)
}
