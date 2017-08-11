package shared

import (
	"math"
	"math/rand"

	"github.com/faiface/pixel"
)

// snaps from to v1 or v2 whichever is closest
// v1 in a tie
func SnapTo(v1, v2, from float64) float64 {
	d1 := math.Abs(v1 - from)
	d2 := math.Abs(v2 - from)
	if d2 > d1 {
		return v2
	}
	return v1
}

// retuns whether or not the scalar distance between
// v1 and v2 is <= r
func WithinRange(v1, v2 pixel.Vec, r float64) bool {
	return v1.Sub(v2).Len() <= r
}

func RandVec(min, max float64) pixel.Vec {
	return pixel.V((max-min)*(rand.Float64()-1/2), (max-min)*(rand.Float64()-1/2))
}

func RectFromCenter(center pixel.Vec, w, h float64) pixel.Rect {
	return pixel.R(center.X-w/2, center.Y-h/2, center.X+w/2, center.Y+h/2)
}

// UnitVec differs from pixel.Vec.Unit() in that, in the case of
// zero vector, return zero vector instead
func UnitVec(v pixel.Vec) pixel.Vec {
	if v == pixel.ZV {
		return pixel.ZV
	}
	return v.Unit()
}

// RoundVec rounds the X and Y components of v
// within precision decimal places (e.g. 0 for integer rounding)
func RoundVec(v pixel.Vec, precision int) pixel.Vec {
	return pixel.V(Round(v.X, 0.5, precision), Round(v.Y, 0.5, precision))
}

// https://gist.github.com/DavidVaini/10308388
func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
