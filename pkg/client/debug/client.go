package debug

import (
	"fmt"
	"math"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/mmogo/mmo/pkg/shared"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

const (
	maxBufferedRequests = 30
	maxBufferedUpdates  = 30

	gameScale = 64.0
)

var (
	map2Screen = func(v pixel.Vec) pixel.Vec {
		return v.Scaled(gameScale)
	}
	screen2Map = func(v pixel.Vec) pixel.Vec {
		return shared.RoundVec(v.Scaled(1.0/gameScale), 0)
	}
)

type client struct {
	conn     net.Conn
	win      *pixelgl.Window
	playerID string
	world    *shared.World
	pongs    chan *shared.Pong
	updates  chan *shared.Update
	requests chan *shared.Request
	//inProcessor     *inputProcessor
	//reqProcessor    *requestProcessor
	errc chan error
	//bufferedUpdates UpdateBuffer
}

func NewClient(id string, conn net.Conn, world *shared.World) *client {
	requests := make(chan *shared.Request, maxBufferedRequests)
	updates := make(chan *shared.Update, maxBufferedUpdates)

	return &client{
		conn:     conn,
		playerID: id,
		//win:          win,
		world:    world,
		requests: requests,
		updates:  updates,
		//inProcessor:  newInputProcessor(win, requests, screen2Map, &cam),
		//reqProcessor: newRequestManager(id, requests, predictions, conn),
		errc:  make(chan error),
		pongs: make(chan *shared.Pong),
	}
}

func (c *client) Run() {
	pixelgl.Run(func() {
		// start window
		cfg := pixelgl.WindowConfig{
			Title:  "loading",
			Bounds: pixel.R(0, 0, 800, 600),
			VSync:  true,
		}
		win, err := pixelgl.NewWindow(cfg)
		if err != nil {
			panic(fmt.Errorf("creating window: %v", err))
		}
		//c.inProcessor = newInputProcessor(win, c.requests, screen2Map, &cam)

		c.win = win

		go c.handleErrors()
		go c.receiveUpdates()
		go c.handleUpdates()
		go c.sendRequests()

		c.run()
	})
}

func (c *client) run() {
	win := c.win

	log.Info("client started")

	txt := text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))

	transform := pixel.IM
	camPosition := pixel.ZV
	winCenter := win.Bounds().Center()
	last := time.Now()
	fps := 0 // calculated frames per second
	second := time.NewTicker(time.Second)

	for !win.Closed() {
		// show fps on title bar
		fps++
		select {
		default:
		case <-second.C:
			win.SetTitle(fmt.Sprintf("MMO (%v fps)", fps))
			fps = 0
		}

		dt := time.Since(last)
		last = time.Now()

		win.Clear(colornames.Darkgray)

		drawDebugCoords(win)

		c.world.ForEach(func(player *shared.Player) {
			screenPos := pixel.V(player.Position.X*gameScale, player.Position.Y*gameScale)
			transform = pixel.IM.Moved(screenPos)
			drawText(win, txt, player.ID, screenPos)

			if player.ID == c.playerID {
				camPosition = pixel.Lerp(camPosition, winCenter.Sub(screenPos), 1-math.Pow(1.0/128, dt.Seconds()))
				cam := pixel.IM.Moved(camPosition)

				mousePos := cam.Unproject(win.MousePosition())
				drawText(win, txt, fmt.Sprintf("%v", screen2Map(mousePos)), mousePos)

				win.SetMatrix(cam)
				c.processInputs(player, cam)
			}
		})
		win.Update()

	}
}

func drawText(target pixel.Target, txt *text.Text, content string, screenPosition pixel.Vec) {
	txt.Clear()
	txt.Dot = txt.Orig
	txt.Dot.X -= txt.BoundsOf(content).W() / 2
	txt.Dot.Y += txt.BoundsOf(content).H()
	txt.WriteString(content)
	txt.DrawColorMask(target, pixel.IM.Moved(screenPosition), colornames.White)
}

func (c *client) processInputs(player *shared.Player, cam pixel.Matrix) {
	if c.win.Pressed(pixelgl.MouseButtonLeft) {
		mouseWorldCoordinates := shared.RoundVec(cam.Unproject(c.win.MousePosition()), 1)
		destination := screen2Map(mouseWorldCoordinates)
		// dont send request if player already in this direction
		if destination != player.Destination {
			c.requests <- &shared.Request{MoveRequest: &shared.MoveRequest{
				Destination: destination,
			}}
		}
	}
}

func (c *client) handleErrors() {
	log.Fatal(<-c.errc)
}

func (c *client) sendRequests() {
	for {
		select {
		case req := <-c.requests:
			if err := shared.SendMessage(&shared.Message{Request: req}, c.conn); err != nil {
				c.errc <- err
			}
		}
	}
}

func (c *client) receiveUpdates() {
	readUpdate := func() error {
		msg, err := shared.GetMessage(c.conn)
		if err != nil {
			return shared.FatalErr(err)
		}
		log.Debugf("RECV", msg)
		if msg.Error != nil {
			return fmt.Errorf("server returned an error: %v", msg.Error.Message)
		}
		if msg.Update != nil {
			c.updates <- msg.Update
		}
		return nil
	}
	for {
		if err := readUpdate(); err != nil {
			c.errc <- err
			continue
		}
	}
}

func (c *client) handleUpdates() {
	for {
		select {
		//authoritative, server-sent
		case update := <-c.updates:
			c.world.ApplyUpdates(update)
		}
	}
}
