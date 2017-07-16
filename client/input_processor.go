package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mmogo/mmo/shared"
)

type inputCache struct {
	direction pixel.Vec
}

//input processor detects user inputs and generates requests
type inputProcessor struct {
	win            *pixelgl.Window
	centerMatrix   pixel.Matrix
	requestsToSend chan *shared.Request
	cache          inputCache

	//related to speech, consider wrapping in a struct
	typing bool
	typed  string
}

func newInputProcessor(win *pixelgl.Window, requests chan *shared.Request) *inputProcessor {
	return &inputProcessor{
		win:            win,
		centerMatrix:   pixel.IM.Moved(win.Bounds().Center()),
		requestsToSend: requests,
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
	direction := pixel.ZV
	if ip.win.Pressed(pixelgl.MouseButtonLeft) {
		mouseWorldCoordinates := shared.RoundVec(ip.centerMatrix.Unproject(ip.win.MousePosition()), 1)
		direction = shared.RoundVec(mouseWorldCoordinates.Unit(), 1)
	}
	// dont send request if player already in this direction
	if direction != player.Direction && direction != ip.cache.direction {
		ip.cache.direction = direction
		ip.pushRequest(&shared.Request{MoveRequest: &shared.MoveRequest{
			Direction: direction,
		}})
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
