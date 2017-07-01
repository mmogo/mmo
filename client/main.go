package main

import (
	_ "image/png"

	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/ilackarms/pkg/errors"
	"github.com/mmogo/mmo/client/assets"
	"github.com/mmogo/mmo/shared"
	"github.com/xtaci/kcp-go"
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

type simulation struct {
	f       func()
	created time.Time
}

var simulations []*simulation
var runSimulations []*simulation
var simLock sync.Mutex
var center pixel.Matrix

func main() {
	addr := flag.String("addr", "localhost:8080", "address for websocket connection")
	id := flag.String("id", "", "playerid to use")
	flag.Parse()
	if *id == "" {
		log.Fatal("id must be provided")
	}
	Main(*addr, *id)
}

func Main(addr, id string) {
	pixelgl.Run(Run(addr, id))
}

func Run(addr, id string) func() {
	return func() {
		if err := run(addr, id); err != nil {
			log.Fatal(err)
		}
	}
}

var (
	playerID            string
	lock                sync.RWMutex
	players             = make(map[string]*shared.ClientPlayer)
	speechLock          sync.RWMutex
	playerSpeech        = make(map[string][]string)
	errc                = make(chan error)
	speechMode          bool
	currentSpeechBuffer string
)

func run(addr, id string) error {
	log.Printf("connecting to %s", addr)
	conn, err := kcp.Dial(addr)
	if err != nil {
		return err
	}
	connectionRequest := &shared.ConnectRequest{
		ID: id,
	}
	playerID = id
	if err := shared.SendMessage(&shared.Message{
		ConnectRequest: connectionRequest,
	}, conn); err != nil {
		return err
	}
	go func() { handleConnection(conn) }()
	lock.Lock()
	players[id] = &shared.ClientPlayer{
		Player: &shared.Player{
			ID:       id,
			Position: pixel.ZV,
		},
	}
	lock.Unlock()

	cfg := pixelgl.WindowConfig{
		Title:  "_anything",
		Bounds: pixel.R(0, 0, 800, 600),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		return errors.New("creating wiondow", err)
	}
	center = pixel.IM.Moved(win.Bounds().Center())
	lootImage, err := loadPicture("sprites/loot.png")
	if err != nil {
		return err
	}
	lootSprite := pixel.NewSprite(lootImage, lootImage.Bounds())

	playerSheet, err := loadPicture("sprites/player.png")
	if err != nil {
		return err
	}
	playerFrames := []pixel.Rect{
		pixel.R(0, 0, 64, 64),
		pixel.R(0, 64, 64, 128),
		pixel.R(64, 64, 128, 128),
	}

	playerSprite := pixel.NewSprite(playerSheet, playerFrames[2])

	animationRate := 10.0 // framerate of player animation
	elapsed := 0.0
	fps := 0
	second := time.Tick(time.Second)
	last := time.Now()

	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	for !win.Closed() {
		win.Clear(colornames.Darkblue)
		dt := time.Since(last).Seconds()
		last = time.Now()

		if err := processPlayerInput(conn, win); err != nil {
			return err
		}

		applySimulations()

		elapsed += dt
		frameChange := 1.0 / animationRate
		frame := int(elapsed/frameChange) % len(playerFrames)
		playerSprite = pixel.NewSprite(playerSheet, playerFrames[frame])
		playerText := text.New(pixel.ZV, atlas)

		lootSprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 2.0))
		lock.RLock()
		pos := players[id].Position
		for _, player := range players {
			playerPos := pixel.IM.Moved(pixel.V(player.Position.X, player.Position.Y))
			playerSprite.DrawColorMask(win, playerPos, player.Color)
			speechLock.RLock()
			txt, ok := playerSpeech[player.ID]
			speechLock.RUnlock()
			if ok && len(txt) > 0 {
				for i, line := range txt {
					playerText.Clear()
					playerText.Dot = playerText.Orig
					playerText.Dot.X -= playerText.BoundsOf(line).W() / 2
					playerText.Dot.Y += playerText.BoundsOf(line).H() * float64(len(txt)-i)
					playerText.WriteString(line + "\n")
					playerText.DrawColorMask(win,
						pixel.IM.Scaled(pixel.ZV, 2).Chained(playerPos.Moved(pixel.V(0, playerText.Bounds().H()+20))),
						player.Color)
				}
			}

			if speechMode && id == player.ID {
				playerText.Clear()
				playerText := text.New(pixel.ZV, atlas)
				playerText.Dot = playerText.Orig
				playerText.Dot.X -= playerText.BoundsOf(currentSpeechBuffer+"_").W() / 2
				playerText.WriteString(currentSpeechBuffer + "_")
				playerText.DrawColorMask(win,
					pixel.IM.Scaled(pixel.ZV, 2).Chained(playerPos.Moved(pixel.V(0, playerText.Bounds().H()+20))),
					colornames.White)
			}
		}
		lock.RUnlock()

		cam := pixel.IM.Moved(win.Bounds().Center().Sub(pixel.V(pos.X, pos.Y)))

		playerText.Clear()
		mousePos := cam.Unproject(win.MousePosition())
		playerText.WriteString(fmt.Sprintf("%v", mousePos))
		playerText.DrawColorMask(win, pixel.IM.Moved(mousePos), colornames.Firebrick)

		win.SetMatrix(cam)
		win.Update()

		fps++
		select {
		default:
		case <-second:
			win.SetTitle(fmt.Sprintf("%v fps", fps))
			fps = 0
		}

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

func requestMove(direction shared.Direction, conn net.Conn) error {
	msg := &shared.Message{
		MoveRequest: &shared.MoveRequest{
			Direction: direction,
			Created:   time.Now(),
		},
	}
	return shared.SendMessage(msg, conn)
}

func requestSpeak(txt string, conn net.Conn) error {
	msg := &shared.Message{
		SpeakRequest: &shared.SpeakRequest{
			Text: txt,
		},
	}
	return shared.SendMessage(msg, conn)
}

func handleConnection(conn net.Conn) {
	loop := func() error {
		msg, err := shared.GetMessage(conn)
		if err != nil {
			return err
		}
		switch {
		case msg.PlayerMoved != nil:
			handlePlayerMoved(msg.PlayerMoved)
		case msg.PlayerSpoke != nil:
			handlePlayerSpoke(msg.PlayerSpoke)
		case msg.WorldState != nil:
			handleWorldState(msg.WorldState)
		case msg.PlayerDisconnected != nil:
			handlePlayerDisconnected(msg.PlayerDisconnected)
		}
		return nil
	}
	for {
		if err := loop(); err != nil {
			errc <- err
			continue
		}
	}
}

func handlePlayerMoved(moved *shared.PlayerMoved) {
	setPlayerPosition(moved.ID, moved.NewPosition)
	reapplySimulations(moved.RequestTime)
}

func handlePlayerSpoke(speech *shared.PlayerSpoke) {
	id := speech.ID
	speechLock.Lock()
	defer speechLock.Unlock()
	txt, ok := playerSpeech[id]
	if !ok {
		txt = []string{}
	}
	if len(txt) > 4 {
		txt = txt[1:]
	}
	txt = append(txt, speech.Text)
	playerSpeech[id] = txt
	go func() {
		time.Sleep(time.Second * 5)
		speechLock.Lock()
		defer speechLock.Unlock()
		txt, ok := playerSpeech[id]
		if !ok {
			txt = []string{}
		}
		if len(txt) > 0 {
			txt = txt[1:]
		}
		playerSpeech[id] = txt
	}()
}

func handleWorldState(worldState *shared.WorldState) {
	lock.Lock()
	defer lock.Unlock()
	for _, player := range worldState.Players {
		players[player.ID] = &shared.ClientPlayer{
			Player: player,
			Color:  stringToColor(player.ID),
		}
	}
}

func handlePlayerDisconnected(disconnected *shared.PlayerDisconnected) {
	lock.Lock()
	defer lock.Unlock()
	delete(players, disconnected.ID)
}

func processPlayerInput(conn net.Conn, win *pixelgl.Window) error {
	if speechMode {
		return processPlayerSpeechInput(conn, win)
	}
	if win.JustPressed(pixelgl.KeyEnter) {
		speechMode = true
		return nil
	}

	// mouse movement
	mousedir := shared.DIR_NONE
	if win.Pressed(pixelgl.MouseButtonLeft) {
		mouse := center.Unproject(win.MousePosition())
		mousedir = shared.UnitToDirection(mouse.Unit())
	}

	if mousedir != shared.DIR_NONE {
		queueSimulation(func() {
			setPlayerPosition(playerID, players[playerID].Position.Add(mousedir.ToVec()))
		})
		if err := requestMove(mousedir, conn); err != nil {
			return err
		}
	}

	// key movement
	if win.Pressed(pixelgl.KeyA) {
		queueSimulation(func() {
			setPlayerPosition(playerID, players[playerID].Position.Add(LEFT.ToVec()))
		})
		if err := requestMove(LEFT, conn); err != nil {
			return err
		}
	}
	if win.Pressed(pixelgl.KeyD) {
		queueSimulation(func() {
			setPlayerPosition(playerID, players[playerID].Position.Add(RIGHT.ToVec()))
		})
		if err := requestMove(RIGHT, conn); err != nil {
			return err
		}
	}
	if win.Pressed(pixelgl.KeyW) {
		queueSimulation(func() {
			setPlayerPosition(playerID, players[playerID].Position.Add(UP.ToVec()))
		})
		if err := requestMove(UP, conn); err != nil {
			return err
		}
	}
	if win.Pressed(pixelgl.KeyS) {
		queueSimulation(func() {
			setPlayerPosition(playerID, players[playerID].Position.Add(DOWN.ToVec()))
		})
		if err := requestMove(DOWN, conn); err != nil {
			return err
		}
	}
	return nil
}

func processPlayerSpeechInput(conn net.Conn, win *pixelgl.Window) error {
	currentSpeechBuffer += win.Typed()
	if win.JustPressed(pixelgl.KeyBackspace) {
		if len(currentSpeechBuffer) < 1 {
			currentSpeechBuffer = ""
		} else {
			currentSpeechBuffer = currentSpeechBuffer[:len(currentSpeechBuffer)-1]
		}
	}
	if win.JustPressed(pixelgl.KeyEscape) {
		currentSpeechBuffer = ""
		speechMode = false
	}
	if win.JustPressed(pixelgl.KeyEnter) {
		var err error
		if len(currentSpeechBuffer) > 0 {
			err = requestSpeak(currentSpeechBuffer, conn)
		}
		currentSpeechBuffer = ""
		speechMode = false
		return err
	}
	return nil
}

func stringToColor(str string) color.Color {
	var c color.RGBA
	for _, char := range str {
		c.R += uint8(char % math.MaxUint8)
		c.G += uint8((char - char%math.MaxUint8) % math.MaxUint8)
		c.G += uint8((char - (char-char%math.MaxUint8)%math.MaxUint8) % math.MaxUint8)
	}
	c.A = 255
	return c
}

func setPlayerPosition(id string, pos pixel.Vec) {
	lock.RLock()
	defer lock.RUnlock()
	player, ok := players[id]
	if !ok {
		player = &shared.ClientPlayer{
			Player: &shared.Player{
				ID: id,
			},
			Color: stringToColor(id),
		}
		players[id] = player
	}
	player.Position = pos
}

func queueSimulation(f func()) {
	simLock.Lock()
	simulations = append(simulations, &simulation{
		f:       f,
		created: time.Now(),
	})
	simLock.Unlock()
}

func applySimulations() {
	simLock.Lock()
	for _, sim := range simulations {
		sim.f()
		runSimulations = append(runSimulations, sim)
	}
	simulations = []*simulation{}
	simLock.Unlock()
}

func reapplySimulations(from time.Time) {
	i := 0
	if len(runSimulations) == 0 {
		return
	}
	simLock.Lock()
	for _, sim := range runSimulations {
		if sim.created.After(from) {
			break
		}
		i++
	}
	simulations = append(runSimulations[i:], simulations...)
	runSimulations = []*simulation{}
	simLock.Unlock()
}
