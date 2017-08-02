package client

import (
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mmogo/mmo/pkg/shared"
)

type inputCache struct {
	destination pixel.Vec
}

//input processor detects user inputs and generates requests
type inputProcessor struct {
	win            *pixelgl.Window
	cam            *pixel.Matrix
	requestsToSend chan *shared.Request
	cache          inputCache
	screen2Map     projectionFunc

	//related to speech, consider wrapping in a struct
	typing bool
	typed  string
}

func newInputProcessor(win *pixelgl.Window, requests chan *shared.Request, screen2Map projectionFunc, cam *pixel.Matrix) *inputProcessor {
	return &inputProcessor{
		win:            win,
		cam:            cam,
		requestsToSend: requests,
		screen2Map:     screen2Map,
	}
}

// query for player inputs and generate requests based on them
// call during window update loop
func (ip *inputProcessor) handleInputs(player *shared.Player, data *renderData) {
	ip.handleMovement(player)
	ip.handleSpeech()
	ip.handleDebug(data)
}

func (ip *inputProcessor) handleMovement(player *shared.Player) {
	if ip.win.Pressed(pixelgl.MouseButtonLeft) {
		mouseWorldCoordinates := shared.RoundVec(ip.cam.Unproject(ip.win.MousePosition()), 1)
		destination := ip.screen2Map(mouseWorldCoordinates)
		// dont send request if player already in this direction
		if destination != player.Destination && destination != ip.cache.destination {
			log.Printf("newdest %v", destination)
			ip.cache.destination = destination
			ip.pushRequest(&shared.Request{MoveRequest: &shared.MoveRequest{
				Destination: destination,
			}})
		}
	}
}

func (ip *inputProcessor) handleSpeech() {
	if !ip.typing {
		if ip.win.JustPressed(pixelgl.KeyEnter) {
			ip.typing = true
			return
		}
	}
	ip.typed += ip.win.Typed()
	if ip.win.JustPressed(pixelgl.KeyBackspace) {
		if len(ip.typed) < 1 {
			ip.typed = ""
		} else {
			ip.typed = ip.typed[:len(ip.typed)-1]
		}
	}
	if ip.win.JustPressed(pixelgl.KeyEscape) {
		ip.typed = ""
		ip.typing = false
	}
	if ip.win.JustPressed(pixelgl.KeyEnter) {
		if len(ip.typed) > 0 {
			ip.pushRequest(&shared.Request{SpeakRequest: &shared.SpeakRequest{
				Text: ip.typed,
			}})
		}
		ip.typed = ""
		ip.typing = false
	}
}

// handle special cases / debug here
func (ip *inputProcessor) handleDebug(data *renderData) {
	if ip.win.JustPressed(pixelgl.KeyF2) {
		data.debugMode = !data.debugMode
	}
}

// so we don't block on processing inputs
// this could theoretically cause a crazy amount of goroutines to be running in parallel...
// TODO: be careful of the overhead this may cause
func (ip *inputProcessor) pushRequest(req *shared.Request) {
	go func() { ip.requestsToSend <- req }()
}
