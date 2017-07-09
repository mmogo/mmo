package main

import (
	"log"
	"math"
	"time"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

var isomatrix = pixel.IM.Rotated(pixel.ZV, 45*(math.Pi/180)).ScaledXY(pixel.ZV, pixel.V(1, 0.5))

func LoadWorld() *pixel.Batch {
	t1 := time.Now()
	grass, err := loadPicture("sprites/grass.png")
	if err != nil {
		log.Fatal(err)
	}
	ground := pixel.NewSprite(grass, grass.Bounds())
	batch := pixel.NewBatch(&pixel.TrianglesData{}, grass)
	batch.SetMatrix(pixel.IM.Rotated(pixel.ZV, 45*(math.Pi/180)).ScaledXY(pixel.ZV, pixel.V(1, 0.5)))
	tilesize := grass.Bounds().Max.X // 64
	mapsize := tilesize * 100        // 6400 wide/high
	var i int
	for y := -mapsize / 2; y <= mapsize/2; y = y + tilesize {
		for x := -mapsize / 2.00; x <= mapsize/2; x = x + tilesize {
			i++
			ground.Draw(batch, pixel.IM.Moved(pixel.V(x, y)))
		}
	}
	log.Printf("world render: %v iter took %s", i, time.Since(t1))
	return batch
}

func LoadGrid() *pixel.Batch {
	t1 := time.Now()
	batch := pixel.NewBatch(&pixel.TrianglesData{}, nil)
	batch.SetMatrix(pixel.IM.Rotated(pixel.ZV, 45*(math.Pi/180)).ScaledXY(pixel.ZV, pixel.V(1, 0.5)))
	imd := imdraw.New(nil)
	// draw spawn area
	imd.Color = colornames.Green
	imd.Push(pixel.ZV)
	imd.Circle(124, 0)
	imd.Color = colornames.Lightgreen
	imd.Push(pixel.ZV)
	imd.Circle(64, 0)
	imd.Color = colornames.White
	imd.Push(pixel.ZV)
	imd.Circle(8, 0)
	imd.Draw(batch)

	imd.Color = pixel.ToRGBA(colornames.Red)
	var i int
	tilesize := 64.00
	mapsize := 6400.00 // height and width
	for y := -mapsize / 2; y <= mapsize/2; y = y + tilesize {
		for x := -mapsize / 2; x <= mapsize/2; x = x + tilesize {
			i++
			imd.Clear()
			imd.Push(pixel.V(x-(tilesize/2), y-(tilesize/2)))
			imd.Push(pixel.V(x+(tilesize/2), y+(tilesize/2)))
			imd.Rectangle(1)
			imd.Draw(batch)
		}
	}

	log.Printf("grid render: %v iter took %s", i, time.Since(t1))
	return batch
}

func getcube(matrix pixel.Matrix, target pixel.Target) {
	imd := imdraw.New(nil)
	imd.SetMatrix(matrix.Chained(pixel.IM.Rotated(pixel.ZV, 45*(math.Pi/180)).ScaledXY(pixel.ZV, pixel.V(1, 0.5))))
	imd.Color = colornames.Yellow
	imd.Push(pixel.V(0, 0), pixel.V(-64, -64))
	imd.Rectangle(3)
	imd.Color = colornames.Blue
	imd.Push(pixel.V(0, 0), pixel.V(64, 64))
	imd.Rectangle(3)
	imd.Push(pixel.V(0, 0), pixel.V(-64, -64))
	imd.Line(3)
	imd.Push(pixel.V(64, 0), pixel.V(0, -64))
	imd.Line(3)
	imd.Push(pixel.V(-64, 0), pixel.V(0, 64))
	imd.Line(3)
	imd.Draw(target)
}
