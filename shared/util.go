package shared

import (
	"math"

	"github.com/faiface/pixel"
)

const tilesize = 64
const tileheighthalf = 32
const tilewidthhalf = 32

// MapToScreen converts a map coordinate to screen coordinate
func MapToScreen(m pixel.Vec, screen pixel.Rect) pixel.Vec {
	x := (m.X - m.Y) * tilewidthhalf
	y := (m.X + m.Y) * tileheighthalf
	x -= tilewidthhalf
	return pixel.Vec{x, y}
}

// ScreenToMap converts map coordinates to tile number on screen
func ScreenToMap(screen pixel.Vec) pixel.Vec {
	x := (screen.X/tilewidthhalf + screen.Y/tileheighthalf) / 2
	y := (screen.Y/tileheighthalf - (screen.X / tilewidthhalf)) / 2
	x += tilewidthhalf
	return pixel.Vec{x, y}
}

// wrong??
func IsoToMap(iso pixel.Vec) pixel.Vec {
	x := (iso.X + (2 * iso.Y)) / 2
	y := ((2 * iso.Y) + iso.X) / 2
	return pixel.Vec{x, y}
}

// Distance between two vectors
func Distance(v1, v2 pixel.Vec) float64 {
	r := pixel.Rect{v1, v2}.Norm()
	v1 = r.Min
	v2 = r.Max
	h := (v1.X - v2.X) * (v1.X - v2.X)
	v := (v1.Y - v2.Y) * (v1.Y - v2.Y)
	return math.Sqrt(h + v)
}
