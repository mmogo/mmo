package main

import (
	"image/color"

	"github.com/faiface/pixel"
)

type drawable interface {
	Draw(target pixel.Target, matrix pixel.Matrix)
	DrawColorMask(t pixel.Target, matrix pixel.Matrix, mask color.Color)
}
