package constants

import "github.com/faiface/pixel"

var (
	Directions struct {
		Up, Down, Left, Right pixel.Vec
	} = struct {
		Up, Down, Left, Right pixel.Vec
	}{
		Up:    pixel.V(0, 1),
		Down:  pixel.V(0, -1),
		Left:  pixel.V(-1, 0),
		Right: pixel.V(1, 0),
	}
)
