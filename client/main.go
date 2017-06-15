package main

import (
	_ "image/png"

	"bytes"
	"flag"
	"image"
	"log"
	"net/url"
	"os"
	"time"

	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/gorilla/websocket"
	"github.com/ilackarms/_anything/client/assets"
	"github.com/ilackarms/_anything/shared"
	"github.com/ilackarms/_anything/shared/constants"
	"github.com/ilackarms/_anything/shared/types"
	"github.com/ilackarms/pkg/errors"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"strings"
	"sync"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "address for websocket connection")
	flag.Parse()
	Main(*addr)
}

func Main(addr string) {
	pixelgl.Run(Run(addr))
}

func Run(addr string) func() {
	return func() {
		if err := run(addr); err != nil {
			log.Fatal(err)
		}
	}
}

var id = os.Getenv("PLAYERID")
var lock sync.RWMutex
var players = make(map[string]*types.Player)
var speechLock sync.RWMutex
var playerSpeech = make(map[string][]string)
var errc = make(chan error)
var speechMode bool
var currentSpeechBuffer string

var buttonLetters = map[pixelgl.Button]rune{
	pixelgl.KeySpace:        ' ',
	pixelgl.KeyApostrophe:   '\'',
	pixelgl.KeyComma:        ',',
	pixelgl.KeyMinus:        '-',
	pixelgl.KeyPeriod:       '.',
	pixelgl.KeySlash:        '/',
	pixelgl.Key0:            '0',
	pixelgl.Key1:            '1',
	pixelgl.Key2:            '2',
	pixelgl.Key3:            '3',
	pixelgl.Key4:            '4',
	pixelgl.Key5:            '5',
	pixelgl.Key6:            '6',
	pixelgl.Key7:            '7',
	pixelgl.Key8:            '8',
	pixelgl.Key9:            '9',
	pixelgl.KeySemicolon:    ';',
	pixelgl.KeyEqual:        '=',
	pixelgl.KeyA:            'a',
	pixelgl.KeyB:            'b',
	pixelgl.KeyC:            'c',
	pixelgl.KeyD:            'd',
	pixelgl.KeyE:            'e',
	pixelgl.KeyF:            'f',
	pixelgl.KeyG:            'g',
	pixelgl.KeyH:            'h',
	pixelgl.KeyI:            'i',
	pixelgl.KeyJ:            'j',
	pixelgl.KeyK:            'k',
	pixelgl.KeyL:            'l',
	pixelgl.KeyM:            'm',
	pixelgl.KeyN:            'n',
	pixelgl.KeyO:            'o',
	pixelgl.KeyP:            'p',
	pixelgl.KeyQ:            'q',
	pixelgl.KeyR:            'r',
	pixelgl.KeyS:            's',
	pixelgl.KeyT:            't',
	pixelgl.KeyU:            'u',
	pixelgl.KeyV:            'v',
	pixelgl.KeyW:            'w',
	pixelgl.KeyX:            'x',
	pixelgl.KeyY:            'y',
	pixelgl.KeyZ:            'z',
	pixelgl.KeyLeftBracket:  '[',
	pixelgl.KeyBackslash:    '\\',
	pixelgl.KeyRightBracket: ']',
	pixelgl.KeyGraveAccent:  '`',
}

var shiftedLetters = map[rune]rune{
	'\'': '"',
	',':  '<',
	'-':  '_',
	'.':  '>',
	'/':  '?',
	'0':  ')',
	'1':  '!',
	'2':  '@',
	'3':  '#',
	'4':  '$',
	'5':  '%',
	'6':  '^',
	'7':  '&',
	'8':  '*',
	'9':  '(',
	';':  ':',
	'=':  '+',
	'a':  'A',
	'b':  'B',
	'c':  'C',
	'd':  'D',
	'e':  'E',
	'f':  'F',
	'g':  'G',
	'h':  'H',
	'i':  'I',
	'j':  'J',
	'k':  'K',
	'l':  'L',
	'm':  'M',
	'n':  'N',
	'o':  'O',
	'p':  'P',
	'q':  'Q',
	'r':  'R',
	's':  'S',
	't':  'T',
	'u':  'U',
	'v':  'V',
	'w':  'W',
	'x':  'X',
	'y':  'Y',
	'z':  'Z',
	'[':  '{',
	'\\': '|',
	']':  '}',
	'`':  '~',
}

func run(addr string) error {
	log.Printf("connecting to %s", addr)
	u := url.URL{Scheme: "ws", Host: addr, Path: "/connect"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	connectionRequest := &shared.ConnectRequest{
		ID: os.Getenv("PLAYERID"),
	}
	if err := shared.SendMessage(&shared.Message{ConnectRequest: connectionRequest}, conn); err != nil {
		return err
	}
	go func() { handleConnection(conn) }()
	lock.Lock()
	players[id] = &types.Player{
		ID:       id,
		Position: pixel.ZV,
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

		angle += 3 * dt

		elapsed += dt
		frameChange := 1.0 / FPS
		frame := int(elapsed/frameChange) % len(mrmanFrames)
		mrManSprite = pixel.NewSprite(mrmanSheet, mrmanFrames[frame])
		playerText := text.New(pixel.ZV, atlas)

		guysleySprite.Draw(win, pixel.IM.Rotated(pixel.ZV, angle).Moved(win.Bounds().Center()))
		lock.RLock()
		pos := players[id].Position
		for _, player := range players {
			mrManPos := pixel.IM.Moved(win.Bounds().Center().Add(pixel.V(player.Position.X, player.Position.Y)))
			//mrManSprite.Draw(win, mrManPos)
			mrManSprite.DrawColorMask(win, mrManPos, colornames.Beige)
			speechLock.RLock()
			txt, ok := playerSpeech[player.ID]
			speechLock.RUnlock()
			if ok {
				speechTxt := strings.Join(txt, "\n")
				playerText.Clear()
				playerText.Dot = playerText.Orig
				playerText.Dot.X -= playerText.BoundsOf(speechTxt).W() / 2
				fmt.Fprint(playerText, speechTxt)
				playerText.DrawColorMask(win, mrManPos.Moved(pixel.V(0, playerText.Bounds().H()+20)), colornames.White)
			}
			if speechMode {
				playerText.Clear()
				playerText.Dot = playerText.Orig
				playerText.Dot.X -= playerText.BoundsOf(currentSpeechBuffer+"_").W() / 2
				fmt.Fprint(playerText, currentSpeechBuffer+"_")
				playerText.DrawColorMask(win, mrManPos.
					Moved(pixel.V(0, playerText.Bounds().H()+20)), colornames.White)
			}
		}
		lock.RUnlock()

		//playerText.Clear()
		cam := pixel.IM.Moved(win.Bounds().Min.Sub(pixel.V(pos.X, pos.Y)))
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

func requestMove(direction pixel.Vec, conn *websocket.Conn) error {
	msg := &shared.Message{
		MoveRequest: &shared.MoveRequest{
			Direction: direction,
		},
	}
	return shared.SendMessage(msg, conn)
}

func requestSpeak(txt string, conn *websocket.Conn) error {
	msg := &shared.Message{
		SpeakRequest: &shared.SpeakRequest{
			Text: txt,
		},
	}
	return shared.SendMessage(msg, conn)
}

func handleConnection(conn *websocket.Conn) {
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
	id := moved.ID
	lock.RLock()
	defer lock.RUnlock()
	player, ok := players[id]
	if !ok {
		player = &types.Player{
			ID:       id,
			Position: moved.NewPosition,
		}
		players[id] = player
	}
	player.Position = moved.NewPosition
}

func handlePlayerSpoke(speech *shared.PlayerSpoke) {
	id := speech.ID
	speechLock.Lock()
	defer speechLock.Unlock()
	txt, ok := playerSpeech[id]
	if !ok {
		txt = []string{}
		playerSpeech[id] = txt
	}
	txt = append(txt, speech.Text)
	go func() {
		time.Sleep(time.Second * 5)
		txt = txt[1:]
	}()
}

func handleWorldState(worldState *shared.WorldState) {
	lock.Lock()
	defer lock.Unlock()
	for _, player := range worldState.Players {
		players[player.ID] = player
	}
}

func handlePlayerDisconnected(disconnected *shared.PlayerDisconnected) {
	lock.Lock()
	defer lock.Unlock()
	delete(players, disconnected.ID)
}

func processPlayerInput(conn *websocket.Conn, win *pixelgl.Window) error {
	if speechMode {
		return processPlayerSpeechInput(conn, win)
	}
	if win.JustPressed(pixelgl.KeyEnter) {
		speechMode = true
		return nil
	}

	//movement
	if win.Pressed(pixelgl.KeyA) {
		if err := requestMove(constants.Directions.Left, conn); err != nil {
			return err
		}
	}
	if win.Pressed(pixelgl.KeyD) {
		if err := requestMove(constants.Directions.Right, conn); err != nil {
			return err
		}
	}
	if win.Pressed(pixelgl.KeyW) {
		if err := requestMove(constants.Directions.Up, conn); err != nil {
			return err
		}
	}
	if win.Pressed(pixelgl.KeyS) {
		if err := requestMove(constants.Directions.Down, conn); err != nil {
			return err
		}
	}
	return nil
}

func processPlayerSpeechInput(conn *websocket.Conn, win *pixelgl.Window) error {
	for button, letter := range buttonLetters {
		if win.JustPressed(button) {
			currentSpeechBuffer += string(letter)
		}
	}
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
		err := requestSpeak(currentSpeechBuffer, conn)
		currentSpeechBuffer = ""
		speechMode = false
		return err
	}
	return nil
}
