package shared

import (
	"fmt"
	"math"

	"github.com/faiface/pixel"
)

type Direction byte

const (
	DIR_NONE Direction = iota
	LEFT
	RIGHT
	UP
	DOWN
	UPLEFT
	UPRIGHT
	DOWNLEFT
	DOWNRIGHT
)

const (
	WEST  = LEFT
	EAST  = RIGHT
	NORTH = UP
	SOUTH = DOWN
)

func (d Direction) String() string {
	switch d {
	case LEFT:
		return "left"
	case RIGHT:
		return "right"
	case UP:
		return "up"
	case DOWN:
		return "down"
	case DOWNLEFT:
		return "down-left"
	case DOWNRIGHT:
		return "down-right"
	case UPLEFT:
		return "up-left"
	case UPRIGHT:
		return "up-right"
	default:
		return fmt.Sprintf("invalid direction: %v", int(d))
	}
}

func (d Direction) ToVec() pixel.Vec {
	switch d {
	case LEFT:
		return pixel.V(-1, 0)
	case RIGHT:
		return pixel.V(1, 0)
	case UP:
		return pixel.V(0, 1)
	case DOWN:
		return pixel.V(0, -1)
	case UPRIGHT:
		return pixel.V(1, 1)
	case UPLEFT:
		return pixel.V(-1, 1)
	case DOWNLEFT:
		return pixel.V(-1, -1)
	case DOWNRIGHT:
		return pixel.V(1, -1)
	default:
		return pixel.V(0, 0)
	}
}

func UnitToDirection(v pixel.Vec) Direction {
	// round up: 0 or 1
	v.X = math.Floor(v.X + 0.5)
	v.Y = math.Floor(v.Y + 0.5)

	switch v {
	default:
		return DIR_NONE
	case LEFT.ToVec():
		return LEFT
	case UPLEFT.ToVec():
		return UPLEFT
	case DOWNLEFT.ToVec():
		return DOWNLEFT
	case RIGHT.ToVec():
		return RIGHT
	case UPRIGHT.ToVec():
		return UPRIGHT
	case DOWNRIGHT.ToVec():
		return DOWNRIGHT
	case DOWN.ToVec():
		return DOWN
	case UP.ToVec():
		return UP
	}
}
