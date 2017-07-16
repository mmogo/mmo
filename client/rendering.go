package main

import (
	"image/color"

	"github.com/faiface/pixel"
)

type drawable interface {
	Draw(target pixel.Target, matrix pixel.Matrix)
	DrawColorMask(t pixel.Target, matrix pixel.Matrix, mask color.Color)
}

// projection func is used to project a vector in one space to a vector in another
// eg project from screen pixels to world coordinates, or vice versa
type projectionFunc func(vec pixel.Vec) pixel.Vec