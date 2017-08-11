package debug

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"
	"github.com/mmogo/mmo/pkg/shared"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var (
	//tile length of region
	tilesPerRegion = 10.0

	// grids that have been drawn
	grids      = make(map[pixel.Vec]*pixel.Batch)
	coordTexts = make(map[pixel.Vec]*pixel.Batch)
)

func drawDebugGrid(target pixel.Target, center pixel.Vec) {
	grid, coords := gridForCener(center)
	grid.Draw(target)
	coords.Draw(target)
}

func gridForCener(center pixel.Vec) (*pixel.Batch, *pixel.Batch) {
	closestCenter := shared.RoundVec(center.Scaled(2/tilesPerRegion), 0).Scaled(tilesPerRegion / 2)
	grid, ok := grids[closestCenter]
	if !ok {
		grid = tilesBatch(closestCenter)
		grids[closestCenter] = grid
	}
	coordsText, ok := coordTexts[closestCenter]
	if !ok {
		coordsText = coordsBatch(closestCenter)
		coordTexts[closestCenter] = coordsText
	}
	return grid, coordsText
}

func tilesBatch(center pixel.Vec) *pixel.Batch {
	batch := pixel.NewBatch(&pixel.TrianglesData{}, nil)
	// http://flarerpg.org/tutorials/isometric_intro/
	imd := imdraw.New(nil)
	imd.Color = colornames.Purple
	imd.Push(pixel.V(0, 0))
	imd.Color = colornames.Green
	imd.Push(pixel.V(gameScale, gameScale))
	imd.Rectangle(1)
	forEachCell(center, func(x, y float64) {
		batch.SetMatrix(pixel.IM.Moved(pixel.V(x, y).Scaled(gameScale)))
		imd.Draw(batch)
	})
	return batch
}

func coordsBatch(center pixel.Vec) *pixel.Batch {
	coords := text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))
	batch := pixel.NewBatch(&pixel.TrianglesData{}, coords.Atlas().Picture())
	forEachCell(center, func(x, y float64) {
		coords.Clear()
		coords.Dot = coords.Orig
		coords.WriteString(fmt.Sprintf("%v,%v", x, y))
		coords.Draw(batch, pixel.IM.Moved(pixel.V(x+0.5, y+0.5).Scaled(gameScale)))
	})
	return batch
}

func forEachCell(center pixel.Vec, f func(x, y float64)) {
	start := time.Now()
	i := 0
	for x := -1*tilesPerRegion + center.X; x < tilesPerRegion+center.X; x++ {
		for y := -1*tilesPerRegion + center.Y; y < tilesPerRegion+center.Y; y++ {
			i++
			f(x, y)
		}
	}
	log.Printf("%v: %v iter took %s", reflect.ValueOf(f).Pointer(), i, time.Since(start))
}
