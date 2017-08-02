package client

import (
	"log"
	"time"

	"fmt"

	"reflect"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func debugTiles(tileSize float64) *pixel.Batch {
	batch := pixel.NewBatch(&pixel.TrianglesData{}, nil)
	// http://flarerpg.org/tutorials/isometric_intro/
	imd := imdraw.New(nil)
	imd.Color = colornames.Purple
	imd.Push(pixel.V(0, 0))
	imd.Color = colornames.Green
	imd.Push(pixel.V(tileSize, tileSize))
	imd.Rectangle(1)
	grid(func(x, y float64) {
		batch.SetMatrix(pixel.IM.Moved(pixel.V(x, y).Scaled(tileSize)))
		imd.Draw(batch)
	})
	return batch
}

var coordbatch = debugCoordsBatch(gameScale)

func drawDebugCoords(target pixel.Target) {
	coordbatch.Draw(target)
}

func debugCoordsBatch(tileSize float64) *pixel.Batch {
	coords := text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))
	batch := pixel.NewBatch(&pixel.TrianglesData{}, coords.Atlas().Picture())
	grid(func(x, y float64) {
		coords.Clear()
		coords.Dot = coords.Orig
		coords.WriteString(fmt.Sprintf("%v,%v", x, y))
		coords.Draw(batch, pixel.IM.Moved(pixel.V(x+0.5, y+0.5).Scaled(tileSize)))
	})
	return batch
}

func grid(f func(x, y float64)) {
	start := time.Now()
	i := 0
	magnitude := 125.0
	for y := -1 * magnitude; y < magnitude; y++ {
		for x := -1 * magnitude; x < magnitude; x++ {
			i++
			f(x, y)
		}
	}
	log.Printf("%v: %v iter took %s", reflect.ValueOf(f).Pointer(), i, time.Since(start))
}
