package api

import (
	"github.com/faiface/pixel"
	"github.com/mmogo/mmo/pkg/shared"
)

// Renderer is responsible for rendering the state onto a target
type Renderer interface {
	// RenderFrame draws frame draws a single frame (within a refresh loop)
	RenderFrame(target pixel.Target, world *shared.World)
}
