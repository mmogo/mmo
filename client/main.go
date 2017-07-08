package main

import (
	_ "image/png"

	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/mmogo/mmo/shared"
	"github.com/xtaci/smux"
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
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
}

type simulation struct {
	f       func()
	created time.Time
}

type GameWorld struct {
	playerID            string
	lock                sync.RWMutex
	players             map[string]*shared.ClientPlayer
	speechLock          sync.RWMutex
	playerSpeech        map[string][]string
	errc                chan error
	speechMode          bool
	currentSpeechBuffer string
	simulations         []*simulation
	runSimulations      []*simulation
	simLock             sync.Mutex
	wincenter           pixel.Vec
	centerMatrix        pixel.Matrix
	facing              shared.Direction
	action              shared.Action
}

func main() {
	addr := flag.String("addr", "localhost:8080", "address of server")
	id := flag.String("id", "", "playerid to use")
	protocol := flag.String("protocol", "udp", fmt.Sprintf("network protocol to use. available %s | %s", shared.ProtocolTCP, shared.ProtocolUDP))
	flag.Parse()
	if *id == "" {
		log.Fatal("id must be provided")
	}
	pixelgl.Run(Run(*protocol, *addr, *id))
}

func Run(protocol, addr, id string) func() {
	return func() {
		if err := run(protocol, addr, id); err != nil {
			log.Fatal(err)
		}
	}
}

func NewGame() *GameWorld {
	g := new(GameWorld)
	g.players = make(map[string]*shared.ClientPlayer)
	g.playerSpeech = make(map[string][]string)
	g.errc = make(chan error)
	return g
}

func run(protocol, addr, id string) error {
	log.Printf("connecting to %s", addr)
	conn, err := shared.Dial(protocol, addr)
	if err != nil {
		return err
	}
	session, err := smux.Client(conn, smux.DefaultConfig())
	if err != nil {
		return err
	}
	stream, err := session.OpenStream()
	if err != nil {
		return err
	}
	conn = stream

	connectionRequest := &shared.ConnectRequest{
		ID: id,
	}

	if err := shared.SendMessage(&shared.Message{
		Request: &shared.Request{
			ConnectRequest: connectionRequest,
		}}, conn); err != nil {
		return err
	}
	log.Printf("connection successful")

	g := NewGame()
	g.playerID = id
	go func() { g.handleConnection(conn) }()
	g.lock.Lock()
	g.players[id] = &shared.ClientPlayer{
		Player: &shared.Player{
			ID:       id,
			Position: pixel.ZV,
		},
	}
	g.lock.Unlock()

	cfg := pixelgl.WindowConfig{
		Title:  "loading",
		Bounds: pixel.R(0, 0, 800, 600),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		return fmt.Errorf("creating window: %v", err)
	}

	// load assets
	lootImage, err := loadPicture("sprites/loot.png")
	if err != nil {
		return err
	}
	lootSprite := pixel.NewSprite(lootImage, lootImage.Bounds())

	playerSprite, err := LoadSpriteSheet("sprites/char1.png", nil)
	if err != nil {
		return shared.FatalErr(err)
	}

	fps := 0 // calculated frames per second
	second := time.Tick(time.Second)
	ping := time.Tick(time.Second * 2)
	last := time.Now()
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	tilebatch := LoadWorld()
	g.wincenter = win.Bounds().Center()
	g.centerMatrix = pixel.IM.Moved(g.wincenter)
	if g.facing == shared.DIR_NONE {
		g.facing = DOWN
	}
	g.action = shared.A_WALK
	go func() {
		for {
			err := <-g.errc
			if shared.IsFatal(err) {
				log.Fatal(err)
			}
			log.Printf("Non-fatal Error: %v", err)
		}
	}()
	camPos := pixel.ZV
	playerText := text.New(pixel.ZV, atlas)
	for !win.Closed() {
		win.Clear(colornames.Yellow)
		dt := time.Since(last).Seconds()
		last = time.Now()

		if err := g.processPlayerInput(conn, win); err != nil {
			return err
		}

		g.applySimulations()

		playerSprite.Animate(dt, g.facing, g.action)

		tilebatch.Draw(win)

		lootSprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 2.0))
		g.lock.RLock()
		pos := g.players[id].Position
		camPos = pixel.Lerp(camPos, g.wincenter.Sub(pos), 1-math.Pow(1.0/128, dt))
		cam := pixel.IM.Moved(camPos)
		win.SetMatrix(cam)
		for _, player := range g.players {
			playerPos := pixel.IM.Moved(player.Position)
			playerSprite.Draw(win, playerPos, player.Color)
			g.speechLock.RLock()
			txt, ok := g.playerSpeech[player.ID]
			g.speechLock.RUnlock()
			if ok && len(txt) > 0 {
				for i, line := range txt {
					playerText.Clear()
					playerText.Dot = playerText.Orig
					playerText.Dot.X -= playerText.BoundsOf(line).W() / 2
					playerText.Dot.Y += playerText.BoundsOf(line).H() * float64(len(txt)-i)
					playerText.WriteString(line + "\n")
					playerText.DrawColorMask(win,
						pixel.IM.Scaled(pixel.ZV, 2).Moved(pixel.V(player.Position.X, player.Position.Y+20)),
						player.Color)
				}
			}

			if g.speechMode && id == player.ID {
				playerText.Clear()
				playerText.Dot = playerText.Orig
				playerText.Dot.X -= playerText.BoundsOf(g.currentSpeechBuffer+"_").W() / 2
				playerText.WriteString(g.currentSpeechBuffer + "_")
				playerText.DrawColorMask(win,
					pixel.IM.Scaled(pixel.ZV, 2).Moved(pixel.V(player.Position.X, player.Position.Y-64)),
					colornames.White)
			}
		}
		g.lock.RUnlock()

		// show mouse coordinates
		mousePos := cam.Unproject(win.MousePosition())
		playerText.Clear()
		playerText.Dot = playerText.Orig
		playerText.WriteString(fmt.Sprintf("%v", mousePos))
		playerText.DrawColorMask(win, pixel.IM.Moved(mousePos), colornames.White)

		win.Update()

		fps++
		select {
		default:
		case <-ping:
			shared.SendMessage(&shared.Message{}, conn)
		}
		select {
		default:
		case <-second:
			win.SetTitle(fmt.Sprintf("%v fps", fps))
			fps = 0
		}
	}
	return nil
}

func requestMove(direction pixel.Vec, conn net.Conn) error {
	msg := &shared.Message{
		Request: &shared.Request{MoveRequest: &shared.MoveRequest{
			Direction: direction,
			Created:   time.Now(),
		},
		}}
	return shared.SendMessage(msg, conn)
}

func requestSpeak(txt string, conn net.Conn) error {
	msg := &shared.Message{
		Request: &shared.Request{SpeakRequest: &shared.SpeakRequest{
			Text: txt,
		}},
	}
	return shared.SendMessage(msg, conn)
}

func (g *GameWorld) handleConnection(conn net.Conn) {
	loop := func() error {
		msg, err := shared.GetMessage(conn)
		if err != nil {
			return shared.FatalErr(err)
		}
		log.Println("RECV", msg)
		if msg.Error != nil {
			return fmt.Errorf("server returned an error: %v", msg.Error.Message)
		}
		if msg.Update != nil {
			g.ApplyUpdate(msg.Update)
		}
		return nil
	}
	for {
		if err := loop(); err != nil {
			g.errc <- err
			continue
		}

	}
}

func (g *GameWorld) ApplyUpdate(update *shared.Update) {
	if update == nil {
		log.Println("nil update")
		return
	}
	if update.PlayerMoved != nil {
		g.handlePlayerMoved(update.PlayerMoved)
	}
	if update.PlayerSpoke != nil {
		g.handlePlayerSpoke(update.PlayerSpoke)
	}
	if update.WorldState != nil {
		g.handleWorldState(update.WorldState)
	}
	if update.PlayerDisconnected != nil {
		g.handlePlayerDisconnected(update.PlayerDisconnected)
	}

}

func (g *GameWorld) handlePlayerMoved(moved *shared.PlayerMoved) {
	g.setPlayerPosition(moved.ID, moved.NewPosition)
	g.reapplySimulations(moved.RequestTime)
}

func (g *GameWorld) handlePlayerSpoke(speech *shared.PlayerSpoke) {
	id := speech.ID
	g.speechLock.Lock()
	defer g.speechLock.Unlock()
	txt, ok := g.playerSpeech[id]
	if !ok {
		txt = []string{}
	}
	if len(txt) > 4 {
		txt = txt[1:]
	}
	txt = append(txt, speech.Text)
	g.playerSpeech[id] = txt
	go func() {
		time.Sleep(time.Second * 5)
		g.speechLock.Lock()
		defer g.speechLock.Unlock()
		txt, ok := g.playerSpeech[id]
		if !ok {
			txt = []string{}
		}
		if len(txt) > 0 {
			txt = txt[1:]
		}
		g.playerSpeech[id] = txt
	}()
}

func (g *GameWorld) handleWorldState(worldState *shared.WorldState) {
	g.lock.Lock()
	defer g.lock.Unlock()
	for _, player := range worldState.Players {
		g.players[player.ID] = &shared.ClientPlayer{
			Player: player,
			Color:  stringToColor(player.ID),
		}
	}
}

func (g *GameWorld) handlePlayerDisconnected(disconnected *shared.PlayerDisconnected) {
	g.lock.Lock()
	defer g.lock.Unlock()
	delete(g.players, disconnected.ID)
}

func (g *GameWorld) processPlayerInput(conn net.Conn, win *pixelgl.Window) error {
	// set sprite facing if none
	if g.facing == shared.DIR_NONE {
		g.facing = shared.DOWN
	}
	g.action = shared.A_IDLE
	// mouse movement
	mousedir := shared.DIR_NONE
	if win.Pressed(pixelgl.MouseButtonLeft) {
		mouse := g.centerMatrix.Unproject(win.MousePosition())
		mousedir = shared.UnitToDirection(mouse.Unit())
		loc := g.players[g.playerID].Position
		g.queueSimulation(func() {
			g.setPlayerPosition(g.playerID, loc.Add(mouse.Unit().Scaled(2)))
		})

		// set sprite facing
		g.facing = mousedir
		g.action = shared.A_WALK

		// send to server
		if err := requestMove(mouse.Unit().Scaled(2), conn); err != nil {
			return err
		}
	}

	if g.speechMode {
		return g.processPlayerSpeechInput(conn, win)
	}
	if win.JustPressed(pixelgl.KeyEnter) {
		g.speechMode = true
		return nil
	}

	if mousedir != shared.DIR_NONE {
		return nil
	}

	return nil
}

func (g *GameWorld) processPlayerSpeechInput(conn net.Conn, win *pixelgl.Window) error {
	g.currentSpeechBuffer += win.Typed()
	if win.JustPressed(pixelgl.KeyBackspace) {
		if len(g.currentSpeechBuffer) < 1 {
			g.currentSpeechBuffer = ""
		} else {
			g.currentSpeechBuffer = g.currentSpeechBuffer[:len(g.currentSpeechBuffer)-1]
		}
	}
	if win.JustPressed(pixelgl.KeyEscape) {
		g.currentSpeechBuffer = ""
		g.speechMode = false
	}
	if win.JustPressed(pixelgl.KeyEnter) {
		var err error
		if len(g.currentSpeechBuffer) > 0 {
			err = requestSpeak(g.currentSpeechBuffer, conn)
		}
		g.currentSpeechBuffer = ""
		g.speechMode = false
		return err
	}
	return nil
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

func (g *GameWorld) setPlayerPosition(id string, pos pixel.Vec) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	player, ok := g.players[id]
	if !ok {
		player = &shared.ClientPlayer{
			Player: &shared.Player{
				ID: id,
			},
			Color: stringToColor(id),
		}
		g.players[id] = player
	}
	player.Position = pos
}

func (g *GameWorld) queueSimulation(f func()) {
	g.simLock.Lock()
	g.simulations = append(g.simulations, &simulation{
		f:       f,
		created: time.Now(),
	})
	g.simLock.Unlock()
}

func (g *GameWorld) applySimulations() {
	g.simLock.Lock()
	for _, sim := range g.simulations {
		sim.f()
		g.runSimulations = append(g.runSimulations, sim)
	}
	g.simulations = []*simulation{}
	g.simLock.Unlock()
}

func (g *GameWorld) reapplySimulations(from time.Time) {
	i := 0
	if len(g.runSimulations) == 0 {
		return
	}
	g.simLock.Lock()
	for _, sim := range g.runSimulations {
		if sim.created.After(from) {
			break
		}
		i++
	}
	g.simulations = append(g.runSimulations[i:], g.simulations...)
	g.runSimulations = []*simulation{}
	g.simLock.Unlock()
}
