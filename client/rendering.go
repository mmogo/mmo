package main

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/mmogo/mmo/shared"
)

type drawable interface {
	Draw(target pixel.Target, matrix pixel.Matrix)
	DrawColorMask(t pixel.Target, matrix pixel.Matrix, mask color.Color)
}

// projection func is used to project a vector in one space to a vector in another
// eg project from screen pixels to world coordinates, or vice versa
type projectionFunc func(vec pixel.Vec) pixel.Vec

// LerpWorld lineraly interpolates between two instances of world
func LerpWorld(w1, w2 *shared.World, t float64) *shared.World {
	lerpedWorld := w1.DeepCopy()
	// it's ok to modify lerpedWorld here
	// trust me
	lerpedWorld.ForEach(func(p *shared.Player) {
		if !p.Active {
			return
		}
		//lookup player in w2
		p2, ok := w2.GetPlayer(p.ID)
		if !ok {
			// player doesnt exist anymore in future; don't bother lerping it
			return
		}
		*p = *LerpPlayer(p, p2, t)
	})
	return lerpedWorld
}

// LerpPlayer lineraly interpolates between two players' math values
// it ignores things that can't be interpolated (e.g. text)
func LerpPlayer(p1, p2 *shared.Player, t float64) *shared.Player {
	lerpedPlayer := p1 //.DeepCopy()
	lerpedPlayer.Position = pixel.Lerp(p1.Position, p2.Position, t)
	lerpedPlayer.Destination = pixel.Lerp(p1.Destination, p2.Destination, t)
	lerpedPlayer.Size = pixel.Lerp(p1.Size, p2.Size, t)
	lerpedPlayer.Speed = p1.Speed + (p2.Speed-p1.Speed)*t
	return lerpedPlayer
}
