package main

import (
	"log"
	"time"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

func LoadWorld() *pixel.Batch {
	t1 := time.Now()
	batch := pixel.NewBatch(&pixel.TrianglesData{}, nil)
	imd := imdraw.New(nil)
	var i int
	for y := -10000.00; y <= 10000; y = y + 100 {

		for x := -10000.00; x <= 10000; x = x + 100 {
			i++
			imd.Color = colornames.Purple
			imd.Push(pixel.V(x, y))
			imd.Color = colornames.Green
			imd.Push(pixel.V(x+50, y+50))
			imd.Rectangle(0)
			imd.Color = colornames.Purple
			imd.Push(pixel.V(x+50, y+50))
			imd.Color = colornames.Green
			imd.Push(pixel.V(x+100, y+100))
		}
	}
	imd.Draw(batch)
	log.Printf("world render: %v iter took %s", i, time.Since(t1))
	return batch
}
