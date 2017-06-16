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
	tickTime       = 1.0 / ticksPerSecond

	maximumMessageSize = 1024 * 1024 //1MB
)

var playersLock = sync.RWMutex{}
var players = make(map[string]*types.ServerPlayer)

var updatesLock = sync.Mutex{}
var updates = []*update{}

type update struct {
	notifyPlayerMoved        *notifyPlayerMoved
	notifyPlayerSpoke        *notifyPlayerSpoke
	notifyWorldState         *notifyWorldState
	notifyPlayerDisconnected *notifyPlayerDisconnected
}

type notifyPlayerMoved struct {
	id          string
	newPosition pixel.Vec
}

type notifyPlayerSpoke struct {
	id   string
	text string
}

type notifyWorldState struct {
	targetID string
}

type notifyPlayerDisconnected struct {
	id string
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
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/connect", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, req, nil)
		if err != nil {
			errc <- errors.New(fmt.Sprintf("failed to upgrade connection for %v", req), err)
			return
		}
		if err := handleConnection(conn); err != nil {
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

func handleConnection(conn *websocket.Conn) error {
	//prevent messages that are too damn big
	conn.SetReadLimit(maximumMessageSize)
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
		queueNotifyPlayerDisconnected(id)
		return nil
	})
	playersLock.Lock()
	defer playersLock.Unlock()
	players[id] = &types.ServerPlayer{
		Player: &types.Player{
			ID:       id,
			Position: pos,
		},
		Conn: conn,
	}
	queuePlayerMovedUpdate(id, pos)
	queueSendWorldStateUpdate(id)
	go handlePlayer(id)
	log.Printf("new connected player %s", id)
	return nil
}

func handlePlayer(id string) {
	last := time.Now()
	dt := 0.0
	//rate limit player requests per second
	for players[id] != nil {
		dt += time.Since(last).Seconds()
		last = time.Now()
		if dt < tickTime {
			time.Sleep(time.Millisecond)
			continue
		}
		dt = 0.0
		player := players[id]
		conn := player.Conn
		msg, err := shared.GetMessage(conn)
		if err != nil {
			log.Print(errors.New(fmt.Sprintf("Client disconnected: (failed getting message for player %s)", id), err))
			delete(players, id)
			queueNotifyPlayerDisconnected(id)
			continue
		}
		switch {
		case msg.MoveRequest != nil:
			handleMoveRequest(id, msg.MoveRequest)
		case msg.SpeakRequest != nil:
			handleSpeakRequest(id, msg.SpeakRequest)
		}
	}
}

func gameLoop(errc chan error) {
	last := time.Now()
	dt := 0.0
	for {
		dt += time.Since(last).Seconds()
		last = time.Now()
		if dt < tickTime {
			time.Sleep(time.Millisecond)
			continue
		}
		dt = 0.0
		if err := tick(); err != nil {
			log.Printf("ERROR IN TICK: %v", err)
			errc <- err
		}
	}
}

func tick() error {
	updatesLock.Lock()
	defer updatesLock.Unlock()
	processed := 0
	for _, update := range updates {
		switch {
		case update.notifyPlayerMoved != nil:
			id, newPos := update.notifyPlayerMoved.id, update.notifyPlayerMoved.newPosition
			if err := broadcastPlayerMoved(id, newPos); err != nil {
				return err
			}
		case update.notifyPlayerSpoke != nil:
			id, txt := update.notifyPlayerSpoke.id, update.notifyPlayerSpoke.text
			if err := broadcastPlayerSpoke(id, txt); err != nil {
				return err
			}
		case update.notifyWorldState != nil:
			if err := sendWorldState(update.notifyWorldState.targetID); err != nil {
				return err
			}
		case update.notifyPlayerDisconnected != nil:
			if err := broadcastPlayerDisconnected(update.notifyPlayerDisconnected.id); err != nil {
				return err
			}
		}
		processed++
	}
	updates = updates[processed:]
	return nil
}

func broadcastPlayerMoved(id string, newPos pixel.Vec) error {
	playerMoved := &shared.Message{
		PlayerMoved: &shared.PlayerMoved{
			ID:          id,
			NewPosition: newPos,
		},
	}
	return broadcast(playerMoved)
}

func broadcastPlayerSpoke(id string, txt string) error {
	playerSpoke := &shared.Message{
		PlayerSpoke: &shared.PlayerSpoke{
			ID:   id,
			Text: txt,
		},
	}
	return broadcast(playerSpoke)
}

func sendWorldState(id string) error {
	playersLock.RLock()
	ps := make([]*types.Player, len(players))
	i := 0
	for _, player := range players {
		ps[i] = &types.Player{
			ID:       player.ID,
			Position: player.Position,
		}
		i++
	}
	player, ok := players[id]
	playersLock.RUnlock()
	if !ok {
		return errors.New("player "+id+" not found", nil)
	}
	return shared.SendMessage(&shared.Message{WorldState: &shared.WorldState{Players: ps}}, player.Conn)
}

func broadcastPlayerDisconnected(id string) error {
	playerDisconnected := &shared.Message{PlayerDisconnected: &shared.PlayerDisconnected{ID: id}}
	return broadcast(playerDisconnected)
}

func broadcast(msg *shared.Message) error {
	data, err := shared.Encode(msg)
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

func handleSpeakRequest(id string, req *shared.SpeakRequest) error {
	playersLock.RLock()
	defer playersLock.RUnlock()
	player := players[id]
	if player == nil {
		return errors.New("requesting player "+id+" is nil??", nil)
	}
	queuePlayerSpokeUpdate(id, req.Text)
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

func queuePlayerSpokeUpdate(id string, txt string) {
	updatesLock.Lock()
	defer updatesLock.Unlock()
	updates = append(updates, &update{
		notifyPlayerSpoke: &notifyPlayerSpoke{
			id:   id,
			text: txt,
		},
	})
}

func queueSendWorldStateUpdate(id string) {
	updatesLock.Lock()
	defer updatesLock.Unlock()
	updates = append(updates, &update{
		notifyWorldState: &notifyWorldState{
			targetID: id,
		},
	})
}

func queueNotifyPlayerDisconnected(id string) {
	updatesLock.Lock()
	defer updatesLock.Unlock()
	updates = append(updates, &update{
		notifyPlayerDisconnected: &notifyPlayerDisconnected{
			id: id,
		},
	})
}
