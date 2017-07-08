package main

import (
	"log"
	"math"
	"time"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

func LoadWorld() *pixel.Batch {
	t1 := time.Now()
	batch := pixel.NewBatch(&pixel.TrianglesData{}, nil)
	batch.SetMatrix(pixel.IM.Rotated(pixel.ZV, 45*(math.Pi/180)).ScaledXY(pixel.ZV, pixel.V(1, 0.5)))
	imd := imdraw.New(nil)
	imd.Color = pixel.ToRGBA(colornames.White).Scaled(0.6)
	var i int
	tilesize := 64.00
	mapsize := 1000.00 // height and width
	for y := -mapsize / 2; y <= mapsize/2; y = y + tilesize {
		for x := -mapsize / 2; x <= mapsize/2; x = x + tilesize {
			imd.Clear()
			i++
			imd.Push(pixel.V(x-(tilesize/2), y-(tilesize/2)))
			imd.Push(pixel.V(x+(tilesize/2), y+(tilesize/2)))
			imd.Rectangle(1)
			imd.Draw(batch)
		}
	}
	log.Printf("world render: %v iter took %s", i, time.Since(t1))
	return batch
}
