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
	"github.com/ilackarms/_anything/client/assets"
	"github.com/ilackarms/_anything/shared"
	"github.com/ilackarms/_anything/shared/constants"
	"github.com/ilackarms/pkg/errors"
	"github.com/xtaci/kcp-go"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

type simulation struct {
	f       func()
	created time.Time
}

var simulations []*simulation
var runSimulations []*simulation
var simLock sync.Mutex

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

	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	for !win.Closed() {
		win.Clear(colornames.Darkblue)
		dt := time.Since(last).Seconds()
		last = time.Now()

		if err := processPlayerInput(conn, win); err != nil {
			return err
		}

		applySimulations()

		angle += 3 * dt

		elapsed += dt
		frameChange := 1.0 / FPS
		frame := int(elapsed/frameChange) % len(mrmanFrames)
		mrManSprite = pixel.NewSprite(mrmanSheet, mrmanFrames[frame])
		playerText := text.New(pixel.ZV, atlas)

		guysleySprite.Draw(win, pixel.IM.Rotated(pixel.ZV, angle))
		lock.RLock()
		pos := players[id].Position
		for _, player := range players {
			mrManPos := pixel.IM.Moved(pixel.V(player.Position.X, player.Position.Y))
			//mrManSprite.Draw(win, mrManPos)
			mrManSprite.DrawColorMask(win, mrManPos, player.Color)
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
						pixel.IM.Scaled(pixel.ZV, 2).Chained(mrManPos.Moved(pixel.V(0, playerText.Bounds().H()+20))),
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
					pixel.IM.Scaled(pixel.ZV, 2).Chained(mrManPos.Moved(pixel.V(0, playerText.Bounds().H()+20))),
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

func requestMove(direction pixel.Vec, conn net.Conn) error {
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

	//movement
	if win.Pressed(pixelgl.KeyA) {
		queueSimulation(func() {
			setPlayerPosition(playerID, players[playerID].Position.Add(constants.Directions.Left))
		})
		if err := requestMove(constants.Directions.Left, conn); err != nil {
			return err
		}
	}
	if win.Pressed(pixelgl.KeyD) {
		queueSimulation(func() {
			setPlayerPosition(playerID, players[playerID].Position.Add(constants.Directions.Right))
		})
		if err := requestMove(constants.Directions.Right, conn); err != nil {
			return err
		}
	}
	if win.Pressed(pixelgl.KeyW) {
		queueSimulation(func() {
			setPlayerPosition(playerID, players[playerID].Position.Add(constants.Directions.Up))
		})
		if err := requestMove(constants.Directions.Up, conn); err != nil {
			return err
		}
	}
	if win.Pressed(pixelgl.KeyS) {
		queueSimulation(func() {
			setPlayerPosition(playerID, players[playerID].Position.Add(constants.Directions.Down))
		})
		if err := requestMove(constants.Directions.Down, conn); err != nil {
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
