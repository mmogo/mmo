package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/mmogo/mmo/shared"
	"github.com/xtaci/smux"
)

type GameWorld struct {
	playerID            string
	lock                sync.RWMutex
	players             map[string]*shared.ClientPlayer
	speechLock          sync.RWMutex
	playerSpeech        map[string][]string
	errc                chan error
	speechMode          bool
	currentSpeechBuffer string
	simLock             sync.Mutex
	wincenter           pixel.Vec
	centerMatrix        pixel.Matrix
	facing              shared.Direction
	action              shared.Action
}

func NewGame() *GameWorld {
	g := new(GameWorld)
	g.players = make(map[string]*shared.ClientPlayer)
	g.playerSpeech = make(map[string][]string)
	g.errc = make(chan error)
	return g
}
func main() {
	addr := flag.String("addr", "au.isupon.us:8080", "address of server")
	id := flag.String("id", "bot", "playerid to use")
	proto := flag.String("protocol", "udp", "protocol to connect")
	flag.Parse()
	log.Fatal(connect(*proto, *addr, *id))

}

func connect(protocol, addr, id string) error {
	log.Printf("connecting to %s", addr)
	g := NewGame()
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

	go func(conn net.Conn) {
		for {
			if msg, err := shared.GetMessage(conn); err != nil {
				log.Fatal(err)
			} else {
				log.Println(msg)
				if msg.Update != nil && msg.Update.WorldState != nil {

					g.handleWorldState(msg.Update.WorldState)

				}
			}
		}
	}(conn)

	ping := time.Tick(2 * time.Second)
	rate := time.Tick(50 * time.Millisecond)

	for {
		select {
		default:
		case <-ping:

			go func() {
				if err := shared.SendMessage(&shared.Message{}, conn); err != nil {
					log.Fatal(err)

				}
			}()
		}

		select {
		default:
		case <-rate:
			switch rand.Intn(100) {
			case 0:
				go func() {
					if err := shared.SendMessage(&shared.Message{
						Request: &shared.Request{SpeakRequest: &shared.SpeakRequest{
							Text: g.Fortune(),
						}}}, conn); err != nil {

						log.Fatal(err)
					}
				}()
			case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15:
				go func() {
					if err := shared.SendMessage(&shared.Message{
						Request: &shared.Request{MoveRequest: &shared.MoveRequest{
							Direction: randomDirection(),
							Created:   time.Now(),
						}}}, conn); err != nil {

						log.Fatal(err)
					}

				}()
			}
		}
	}

	return nil
}

func (g *GameWorld) Fortune() string {
	switch rand.Intn(3) {
	case 0:
		return time.Now().String()
	case 1:
		return "hi " + g.randPlayer()
	case 2:
		return time.Now().String()
	default:
		return time.Now().String()
	}
}

func (g *GameWorld) randPlayer() string {
	if len(g.players) < 1 {
		return ""
	}
	var i int
	for _, player := range g.players {

		if rand.Intn(len(g.players)) == i {

			return player.ID
		}
		i++
	}
	return "bugs"
}

func (g *GameWorld) handleWorldState(worldState *shared.WorldState) {
	g.lock.Lock()
	defer g.lock.Unlock()
	for _, player := range worldState.Players {
		g.players[player.ID] = &shared.ClientPlayer{
			Player: player,
		}
	}
}

func randomDirection() pixel.Vec {
	dirs := []shared.Direction{shared.UP, shared.DOWN, shared.LEFT, shared.RIGHT}
	r := rand.Intn(len(dirs))
	return dirs[r].ToVec().Scaled(10)
}
