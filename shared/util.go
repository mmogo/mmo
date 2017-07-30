// Copyright 2017 The MMOGO Authors. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

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
