package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/faiface/pixel"
	"github.com/mmogo/mmo/shared"
	"github.com/ilackarms/pkg/errors"
	"github.com/xtaci/kcp-go"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	ticksPerSecond = 10
	tickTime       = 1.0 / ticksPerSecond

	messagePerTickLimit = 60
)

func main() {
	port := flag.Int("p", 8080, "port to serve on")
	flag.Parse()
	errc := make(chan error)
	go func() { errc <- serve(*port, errc) }()
	go func() { gameLoop(errc) }()
	select {
	case err := <-errc:
		log.Fatal("error somewhere", err)
	}
}

func serve(port int, errc chan error) error {
	laddr := fmt.Sprintf(":%v", port)
	//get client checksums
	clientChecksums := map[string]string{
		"client-windows-4.0-amd64.exe": "",
		"client-darwin-10.6-amd64":     "",
		"client-linux-amd64":           "",
	}
	//requires clients to be in same dir as server
	for client := range clientChecksums {
		f, err := os.Open(client)
		if err != nil {
			continue
		}
		h := md5.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		clientChecksums[client] = string(h.Sum(nil))
	}

	mux := http.NewServeMux()
	mux.Handle("/",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method != "GET" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			for client, checksum := range clientChecksums {
				if strings.Contains(req.URL.Path, client) && req.URL.Query().Get("checksum") == checksum {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			http.FileServer(http.Dir(".")).ServeHTTP(w, req)
		}),
	)
	go func(){
		log.Printf("File server crashed: %v", http.ListenAndServe(laddr, mux))
	}()
	l, err := kcp.Listen(laddr)
	if err != nil {
		return err
	}
	log.Printf("listening for connections on %v", port)
	for {
		conn, err := l.Accept()
		if err != nil {
			errc <- errors.New("failed to establish connection", err)
			continue
		}
		if err := handleConnection(conn); err != nil {
			errc <- errors.New("error handling connection", err)
			continue
		}
	}
}

func handleConnection(conn net.Conn) error {
	msg, err := shared.GetMessage(conn)
	if err != nil {
		return err
	}
	if msg.ConnectRequest == nil {
		return errors.New("expected first message to be ConnectRequest", nil)
	}
	id := msg.ConnectRequest.ID
	pos := pixel.ZV
	playersLock.Lock()
	defer playersLock.Unlock()
	players[id] = &shared.ServerPlayer{
		Player: &shared.Player{
			ID:       id,
			Position: pos,
		},
		Conn: conn,
	}
	queueUpdate(&update{
		notifyPlayerMoved: &notifyPlayerMoved{
			id:          id,
			newPosition: pos,
			requestTime: time.Now(),
		},
	})
	queueUpdate(&update{
		notifyWorldState: &notifyWorldState{
			targetID: id,
		},
	})
	go handlePlayer(id)
	log.Printf("new connected player %s from %s", id, conn.RemoteAddr().String())
	return nil
}

func handlePlayer(id string) {
	for players[id] != nil {
		player := players[id]
		for len(player.RequestQueue) >= messagePerTickLimit {
			time.Sleep(time.Millisecond)
		}
		conn := player.Conn
		msg, err := shared.GetMessage(conn)
		if err != nil {
			log.Print(errors.New(fmt.Sprintf("Client disconnected: (failed getting message for player %s)", id), err))
			playersLock.Lock()
			delete(players, id)
			playersLock.Unlock()
			queueUpdate(&update{
				notifyPlayerDisconnected: &notifyPlayerDisconnected{
					id: id,
				},
			})
			continue
		}
		player.QueueLock.Lock()
		player.RequestQueue = append(player.RequestQueue, msg)
		player.QueueLock.Unlock()
	}
}

func gameLoop(errc chan error) {
	last := time.Now()
	dt := 0.0
	for {
		dt += time.Since(last).Seconds()
		last = time.Now()
		if dt < tickTime {
			sleepTime := time.Duration(1000000*(tickTime-dt)) * time.Microsecond
			time.Sleep(sleepTime)
		}
		dt = 0.0
		if err := tick(); err != nil {
			log.Printf("ERROR IN TICK: %v", err)
			//TODO: handle tick errors
			//errc <- err
		}
	}
}

func tick() error {
	for id, player := range players {
		player.QueueLock.Lock()
		for _, msg := range player.RequestQueue {
			switch {
			case msg.MoveRequest != nil:
				handleMoveRequest(id, msg.MoveRequest)
			case msg.SpeakRequest != nil:
				handleSpeakRequest(id, msg.SpeakRequest)
			}
		}
		player.RequestQueue = []*shared.Message{}
		player.QueueLock.Unlock()
	}

	updatesLock.Lock()
	defer updatesLock.Unlock()
	processed := 0
	for _, update := range updates {
		switch {
		case update.notifyPlayerMoved != nil:
			id, newPos, requestTime := update.notifyPlayerMoved.id, update.notifyPlayerMoved.newPosition, update.notifyPlayerMoved.requestTime
			if err := broadcastPlayerMoved(id, newPos, requestTime); err != nil {
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

func broadcastPlayerMoved(id string, newPos pixel.Vec, requestTime time.Time) error {
	playerMoved := &shared.Message{
		PlayerMoved: &shared.PlayerMoved{
			ID:          id,
			NewPosition: newPos,
			RequestTime: requestTime,
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
	ps := make([]*shared.Player, len(players))
	i := 0
	for _, player := range players {
		ps[i] = &shared.Player{
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
	return shared.SendMessage(&shared.Message{
		WorldState: &shared.WorldState{Players: ps},
	}, player.Conn)
}

func broadcastPlayerDisconnected(id string) error {
	playerDisconnected := &shared.Message{
		PlayerDisconnected: &shared.PlayerDisconnected{ID: id},
	}
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
	queueUpdate(&update{
		notifyPlayerMoved: &notifyPlayerMoved{
			id:          id,
			newPosition: player.Position,
			requestTime: req.Created,
		},
	})
	return nil
}

func handleSpeakRequest(id string, req *shared.SpeakRequest) error {
	playersLock.RLock()
	defer playersLock.RUnlock()
	player := players[id]
	if player == nil {
		return errors.New("requesting player "+id+" is nil??", nil)
	}
	queueUpdate(&update{
		notifyPlayerSpoke: &notifyPlayerSpoke{
			id:   id,
			text: req.Text,
		},
	})
	return nil
}

func queueUpdate(u *update) {
	updatesLock.Lock()
	defer updatesLock.Unlock()
	updates = append(updates, u)
}
