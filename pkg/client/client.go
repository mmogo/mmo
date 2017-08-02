package client

import (
	"fmt"
	"image/color"
	"math"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/ilackarms/pkg/errors"
	"github.com/mmogo/mmo/pkg/shared"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

const (
	UP        = shared.UP
	DOWN      = shared.DOWN
	LEFT      = shared.LEFT
	RIGHT     = shared.RIGHT
	UPLEFT    = shared.UPLEFT
	UPRIGHT   = shared.UPRIGHT
	DOWNLEFT  = shared.DOWNLEFT
	DOWNRIGHT = shared.DOWNRIGHT

	gameScale = 64.0

	maxBufferedUpdates  = 30
	maxBufferedRequests = 30

	tickTime = time.Second / 10 //10 updates per sec

	speechDisplayDuration = time.Second * 5
)

var (
	map2Screen = func(v pixel.Vec) pixel.Vec {
		return v.Scaled(gameScale)
	}
	screen2Map = func(v pixel.Vec) pixel.Vec {
		return shared.RoundVec(v.Scaled(1.0/gameScale), 0)
	}

	cam pixel.Matrix
)

type client struct {
	conn            net.Conn
	win             *pixelgl.Window
	playerID        string
	world           *shared.World
	pongs           chan *shared.Pong
	updates         chan *shared.Update
	predictions     chan *shared.Update
	requests        chan *shared.Request
	inProcessor     *inputProcessor
	reqProcessor    *requestProcessor
	errc            chan error
	bufferedUpdates UpdateBuffer
}

func NewClient(id string, conn net.Conn, win *pixelgl.Window, world *shared.World) *client {
	requests := make(chan *shared.Request, maxBufferedRequests)
	updates := make(chan *shared.Update, maxBufferedUpdates)
	predictions := make(chan *shared.Update, maxBufferedUpdates)

	return &client{
		conn:         conn,
		playerID:     id,
		win:          win,
		world:        world,
		requests:     requests,
		updates:      updates,
		predictions:  predictions,
		inProcessor:  newInputProcessor(win, requests, screen2Map, &cam),
		reqProcessor: newRequestManager(id, requests, predictions, conn),
		errc:         make(chan error),
		pongs:        make(chan *shared.Pong),
	}
}

func (c *client) Run() {
	go c.readUpdates()
	go c.processUpdates()
	go c.reqProcessor.processPending(c.errc)
	go c.handleErrors()
	go c.stepWorld()

	log.Info("client started")

	c.render()
}

func (c *client) handleErrors() {
	for {
		err := <-c.errc
		if shared.IsFatal(err) {
			log.Fatal(err)
		}
		log.Errorf("Error: %v", err)
	}
}

func (c *client) readUpdates() {
	readUpdate := func() error {
		msg, err := shared.GetMessage(c.conn)
		if err != nil {
			return shared.FatalErr(err)
		}
		log.Debugf("RECV", msg)
		if msg.Error != nil {
			return fmt.Errorf("server returned an error: %v", msg.Error.Message)
		}
		if msg.Pong != nil {
			go func() { c.pongs <- msg.Pong }()
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

func (c *client) latency() time.Duration {
	start := time.Now()
	shared.SendMessage(&shared.Message{Ping: &shared.Ping{}}, c.conn)
	select {
	case <-time.After(time.Second):
		c.errc <- fmt.Errorf("timed out waiting for pong")
	case <-c.pongs:
	}
	return time.Since(start)
}

func (c *client) processUpdates() {
	for {
		select {
		//authoritative, server-sent
		case update := <-c.updates:
			processed := update.Processed.Add(c.latency() / 2)
			// TODO: evaluate whether we should do this rollback of state
			// right now it makes things jittery and adds nothing useful
			if false {
				bef := c.world.Before(processed)
				if bef != c.world {
					log.Printf("stepping back to world at %v", processed)
				}
				c.world = bef
			}
			c.bufferedUpdates = c.bufferedUpdates.From(processed)
			if err := c.world.ApplyUpdates(update); err != nil {
				c.errc <- err
			}
			if err := c.world.ApplyUpdates(c.bufferedUpdates...); err != nil {
				c.errc <- err
			}
		case prediction := <-c.predictions:
			if false {
				c.world.ApplyUpdates(prediction)
			}
		case processed := <-c.world.ProcessedUpdates():
			c.bufferedUpdates.Insert(processed)
		}
	}
}

func (c *client) stepWorld() {
	tick := time.NewTicker(tickTime)
	last := time.Now()
	for {
		select {
		case now := <-tick.C:
			c.world.Step(now.Sub(last))
		}
		last = time.Now()
	}
}

func (c *client) render() {
	data, err := c.loadDrawables()
	if err != nil {
		c.errc <- shared.FatalErr(err)
	}

	var (
		camPosition pixel.Vec
	)

	// for drawing
	win := c.win
	batches := data.batches
	drawables := data.drawables
	txt := data.txt

	// for calculating camera position
	windowCenter := win.Bounds().Center()

	// for calculating deltatime
	last := time.Now()

	fps := 0 // calculated frames per second
	second := time.NewTicker(time.Second)

	var lerpTime time.Duration

	var prev *shared.World

	for !win.Closed() {
		// a step happened
		if prev != c.world.Prev() {
			prev = c.world.Prev()
			lerpTime = 0
		}
		// wait for a step
		if prev == nil {
			continue
		}
		dt := time.Since(last)
		last = time.Now()
		win.Clear(colornames.Darkgray)
		if !data.debugMode {
			batches["debug_grid"].Draw(win)
			drawDebugCoords(win)
		}
		playerSprite := drawables["player"]

		var self *shared.Player
		var selfTransform pixel.Matrix
		var mappedPos pixel.Vec

		lerpTime += dt
		//t := time.Since(prev.Updated).Seconds() / c.world.Updated.Sub(prev.Updated).Seconds()
		t := shared.Clamp(lerpTime.Seconds()/c.world.Updated.Sub(prev.Updated).Seconds(), 0, 1)
		//log.Printf("lerpin thru time: %v", t)
		LerpWorld(prev, c.world, t).ForEach(func(player *shared.Player) {
			if !player.Active {
				return
			}
			mappedPos = map2Screen(player.Position)
			transform := pixel.IM.Moved(mappedPos)
			clr := stringToColor(player.ID)
			playerAnimation := playerSprite.(*Sprite)
			playerAnimation.Animate(dt.Seconds(), shared.UnitToDirection(mappedPos), shared.A_WALK)

			playerSprite.DrawColorMask(win, transform, clr)
			for i, speechMsg := range player.SpeechBuffer {
				line := speechMsg.Txt
				if line == "" {
					break
				}
				if time.Since(speechMsg.Timestamp) > speechDisplayDuration {
					break
				}
				txt.Clear()
				txt.Dot = txt.Orig
				txt.Dot.X -= txt.BoundsOf(line).W() / 2
				txt.Dot.Y += txt.BoundsOf(line).H() * float64(len(player.SpeechBuffer)-i)
				txt.WriteString(line + "\n")
				txt.DrawColorMask(win,
					pixel.IM.Scaled(pixel.ZV, 2).Moved(pixel.V(mappedPos.X, mappedPos.Y+20)),
					clr)
			}
			if c.playerID == player.ID {
				self = player
				selfTransform = transform
			}
		})

		if self == nil {
			panic("self not found in world??")
		}

		// handle inputs here
		c.inProcessor.handleInputs(self, data)

		if c.inProcessor.typing {
			txt.Clear()
			txt.Dot = txt.Orig
			txt.Dot.X -= txt.BoundsOf(c.inProcessor.typed+"_").W() / 2
			txt.WriteString(c.inProcessor.typed + "_")
			txt.DrawColorMask(win, selfTransform.Moved(pixel.V(0, -64)), colornames.White)
		}

		camPosition = pixel.Lerp(camPosition, windowCenter.Sub(mappedPos), 1-math.Pow(1.0/128, dt.Seconds()))
		cam = pixel.IM.Moved(camPosition)
		if !data.debugMode {
			mousePos := cam.Unproject(win.MousePosition())
			txt.Clear()
			txt.Dot = txt.Orig
			txt.WriteString(fmt.Sprintf("%v", screen2Map(mousePos)))
			txt.DrawColorMask(win, pixel.IM.Moved(mousePos), colornames.White)
		}

		win.SetMatrix(cam)

		drawables["loot"].Draw(win, selfTransform)
		playerSprite.Draw(win, selfTransform)

		win.Update()

		// show fps on title bar
		fps++
		select {
		default:
		case <-second.C:
			win.SetTitle(fmt.Sprintf("MMO (%v fps)", fps))
			fps = 0
		}
	}
}

type renderData struct {
	drawables map[string]drawable
	batches   map[string]*pixel.Batch
	txt       *text.Text
	debugMode bool
}

func (c *client) loadDrawables() (*renderData, error) {
	drawables := make(map[string]drawable)
	batches := make(map[string]*pixel.Batch)
	batches["debug_grid"] = debugTiles(gameScale)
	lootImage, err := loadImage("sprites/loot.png")
	if err != nil {
		return nil, errors.New("failed to load image", err)
	}
	drawables["loot"] = pixel.NewSprite(lootImage, lootImage.Bounds())
	drawables["player"], err = loadSpriteSheet("sprites/char1.png", nil)
	if err != nil {
		return nil, errors.New("failed to load player sprite", err)
	}
	textAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pixel.ZV, textAtlas)
	return &renderData{
		drawables: drawables,
		batches:   batches,
		txt:       txt,
	}, nil
}

func stringToColor(str string) color.Color {
	colornum := 0
	for _, s := range str {
		colornum += int(s)
	}
	all := len(colornames.Names)
	name := colornames.Names[colornum%all]
	return colornames.Map[name]
}
