package main

import (
	_ "image/png"
	"log"

	"bytes"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/ilackarms/_anything/client/assets"
	"github.com/ilackarms/pkg/errors"
	"golang.org/x/image/colornames"
	"image"
	"time"
)

func main() {
	Main()
}

func Main() {
	pixelgl.Run(Run)
}

func Run() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := pixelgl.WindowConfig{
		Title:  "_anything",
		Bounds: pixel.R(0, 0, 800, 600),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		return errors.New("creating wiondow", err)
	}
	guysleyImage, err := loadPicture("guysley.png")
	if err != nil {
		return err
	}
	guysleySprite := pixel.NewSprite(guysleyImage, guysleyImage.Bounds())

	mrmanSheet, err := loadPicture("mrman.png")
	if err != nil {
		return err
	}
	mrmanFrames := []pixel.Rect{
		pixel.R(0, 0, 64, 64),
		pixel.R(0, 64, 64, 128),
		pixel.R(64, 64, 128, 128),
	}

	mrManSprite := pixel.NewSprite(mrmanSheet, mrmanFrames[2])

	FPS := 10.0
	elapsed := 0.0

	angle := 0.0
	last := time.Now()
	x, y := 0.0, 0.0
	vel := 5.00

	for !win.Closed() {
		win.Clear(colornames.Darkblue)
		dt := time.Since(last).Seconds()
		last = time.Now()

		if win.Pressed(pixelgl.KeyA) {
			x -= vel
		}
		if win.Pressed(pixelgl.KeyD) {
			x += vel
		}
		if win.Pressed(pixelgl.KeyW) {
			y += vel
		}
		if win.Pressed(pixelgl.KeyS) {
			y -= vel
		}

		angle += 3 * dt

		elapsed += dt
		frameChange := 1.0 / FPS
		frame := int(elapsed/frameChange) % len(mrmanFrames)
		mrManSprite = pixel.NewSprite(mrmanSheet, mrmanFrames[frame])

		guysleySprite.Draw(win, pixel.IM.Rotated(pixel.ZV, angle).Moved(win.Bounds().Center()))
		mrManPos := pixel.IM.Moved(win.Bounds().Center().Add(pixel.V(x, y)))
		mrManSprite.Draw(win, mrManPos)
		cam := pixel.IM.Moved(win.Bounds().Min.Sub(pixel.V(x, y)))
		win.SetMatrix(cam)
		win.Update()
	}
	return nil
}

func loadPicture(path string) (pixel.Picture, error) {
	contents, err := assets.Asset(path)
	if err != nil {
		return nil, err
	}
	file := bytes.NewBuffer(contents)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}
