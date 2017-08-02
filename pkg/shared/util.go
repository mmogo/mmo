package shared

import "github.com/faiface/pixel"

// MapToIso converts cartesian coordinates to isometric
func MapToIso(cart pixel.Vec) pixel.Vec {
	return pixel.Vec{cart.X - cart.Y, (cart.X + cart.Y) / 2}
}

// IsoToMap converts isometric coordinates to cartesion
func IsoToMap(iso pixel.Vec) pixel.Vec {
	x := (iso.X + (2 * iso.Y))
	y := ((2 * iso.Y) + iso.X) / 2
	return pixel.Vec{x, y}
}

// Clamps v between min and max
func Clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
