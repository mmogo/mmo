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

var playersize = pixel.R(-16, -16, 16, 16)

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

func (p Player) Bounds() pixel.Rect {
	return playersize.Moved(p.Position)
}
