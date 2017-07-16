package shared

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/faiface/pixel"
)

const fatalErrSig = "**FATAL_ERR**"

type ClientPlayer struct {
	Player *Player
	Color  color.Color
}

type Player struct {
	// global unique UUID
	ID string
	// cartesian coordinates
	Position    pixel.Vec
	Destination pixel.Vec
	// speed is the magnitude of player's velocity
	// in any direction of movement
	Speed float64
	// size in 2 dimensions (x=w, y=h)
	Size pixel.Vec
	// player speech; max buffer size 4
	SpeechBuffer []SpeechMesage
	// if set to false, player is treaded as though it has been deleted
	// this allows us to activate/deactivate players without deleting from state
	Active bool
}

func (p *Player) DeepCopy() *Player {
	speechCopy := make([]SpeechMesage, len(p.SpeechBuffer))
	for i, txt := range p.SpeechBuffer {
		speechCopy[i] = txt
	}
	return &Player{
		ID:           p.ID,
		Position:     p.Position,
		Destination:  p.Destination,
		Speed:        p.Speed,
		Size:         p.Size,
		SpeechBuffer: speechCopy,
		Active:       p.Active,
	}
}

type SpeechMesage struct {
	Txt       string
	Timestamp time.Time
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
