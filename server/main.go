package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/gorilla/websocket"
	"github.com/ilackarms/_anything/shared"
	"github.com/ilackarms/_anything/shared/types"
	"github.com/ilackarms/pkg/errors"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	ticksPerSecond = 60
)

var playersLock = sync.RWMutex{}
var players = make(map[string]*types.Player)

var updatesLock = sync.Mutex{}
var updates = []*update{}

type update struct {
	notifyPlayerMoved *notifyPlayerMoved
}

type notifyPlayerMoved struct {
	id          string
	newPosition pixel.Vec
}

func main() {
	errc := make(chan error)
	go func() { errc <- serveClient(errc) }()
	go func() { gameLoop(errc) }()
	select {
	case err := <-errc:
		log.Fatal("error somewhere", err)
	}
}

func serveClient(errc chan error) error {
	http.Handle("/client/", http.StripPrefix("/client/", http.FileServer(http.Dir("."))))
	http.Handle("/connect", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, req, nil)
		if err != nil {
			errc <- errors.New(fmt.Sprintf("failed to upgrade connection for %v", req), err)
			return
		}
		if err := handleConnection(conn, errc); err != nil {
			errc <- errors.New(fmt.Sprintf("error handling connection %v", req), err)
			return
		}
	}))
	log.Printf("serving client")
	if err := http.ListenAndServe(":8080", http.DefaultServeMux); err != nil {
		return errors.New("failed listening on socket", err)
	}
	return nil
}

func handleConnection(conn *websocket.Conn, errc chan error) error {
	msg, err := shared.GetMessage(conn)
	if err != nil {
		return err
	}
	if msg.ConnectRequest == nil {
		return errors.New("expected first message to be ConnectRequest", nil)
	}
	id := msg.ConnectRequest.ID
	pos := pixel.ZV
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("Client %s disconnected: (%v) %s", id, code, text)
		playersLock.Lock()
		defer playersLock.Unlock()
		delete(players, id)
		return nil
	})
	playersLock.Lock()
	defer playersLock.Unlock()
	players[id] = &types.Player{
		ID:       id,
		Position: pos,
		Conn:     conn,
	}
	queuePlayerMovedUpdate(id, pos)
	go handlePlayer(id, errc)
	log.Printf("new connected player %s", id)
	return nil
}

func handlePlayer(id string, errc chan error) {
	for players[id] != nil {
		player := players[id]
		conn := player.Conn
		msg, err := shared.GetMessage(conn)
		if err != nil {
			log.Print(errors.New(fmt.Sprintf("ERROR: getting message for player %s", id), err))
			delete(players, id)
			continue
		}
		switch {
		case msg.MoveRequest != nil:
			handleMoveRequest(id, msg.MoveRequest)
		}
	}
}

func gameLoop(errc chan error) {
	last := time.Now()
	dt := 0.0
	for {
		dt += time.Since(last).Seconds()
		last = time.Now()
		if dt > 1.0/ticksPerSecond {
			dt = 0.0
			if err := tick(); err != nil {
				log.Printf("ERROR IN TICK: %v", err)
				errc <- err
			}
		} else {
			//prevent locking goroutines
			time.Sleep(time.Millisecond)
		}
	}
}

func tick() error {
	updatesLock.Lock()
	defer updatesLock.Unlock()
	processed := 0
	for _, update := range updates {
		if update.notifyPlayerMoved != nil {
			id, newPos := update.notifyPlayerMoved.id, update.notifyPlayerMoved.newPosition
			if err := broadcastPlayerMoved(id, newPos); err != nil {
				return err
			}
		}
		processed++
	}
	updates = updates[processed:]
	return nil
}

func broadcastPlayerMoved(id string, newPos pixel.Vec) error {
	playerMoved := shared.Message{
		PlayerMoved: &shared.PlayerMoved{
			ID:          id,
			NewPosition: newPos,
		},
	}
	data, err := shared.Encode(playerMoved)
	if err != nil {
		return err
	}
	playersLock.RLock()
	defer playersLock.RUnlock()
	for _, player := range players {
		if err := shared.SendRaw(data, player.Conn); err != nil {
			return err
		}
	}
	return nil
}

func handleMoveRequest(id string, req *shared.MoveRequest) error {
	playersLock.RLock()
	defer playersLock.RUnlock()
	player := players[id]
	if player == nil {
		return errors.New("requesting player "+id+" is nil??", nil)
	}

	player.Position = player.Position.Add(req.Direction)
	queuePlayerMovedUpdate(id, player.Position)
	return nil
}

func queuePlayerMovedUpdate(id string, pos pixel.Vec) {
	updatesLock.Lock()
	defer updatesLock.Unlock()
	updates = append(updates, &update{
		notifyPlayerMoved: &notifyPlayerMoved{
			id:          id,
			newPosition: pos,
		},
	})
}
