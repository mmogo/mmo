package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ilackarms/pkg/errors"
	"github.com/mmogo/mmo/shared"
	"github.com/soheilhy/cmux"
	"github.com/xtaci/smux"
)

type mmoServer struct {
	mgr *updateManager
}

func newMMOServer() *mmoServer {
	return &mmoServer{
		mgr: newUpdateManager(),
	}
}

func (s *mmoServer) start(protocol string, port int, errc chan error) error {
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
		log.Printf("serving client: %s", client)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		for client, checksum := range clientChecksums {
			if strings.Contains(req.URL.Path, client) && req.URL.Query().Get("checksum") == checksum {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if strings.Contains(req.URL.Path, client) {
				log.Printf("serving client: %s", client)
				http.ServeFile(w, req, client)
				return
			}

		}
		log.Printf("bad http request: %s", req.URL.Path)
		http.NotFound(w, req)

	})

	l, err := shared.Listen(protocol, laddr)
	if err != nil {
		return fmt.Errorf("fatal: %v", err)
	}

	if protocol == shared.ProtocolTCP {
		// Create a cmux.
		m := cmux.New(l)
		httpL := m.Match(cmux.HTTP1Fast(), cmux.HTTP1())
		l = m.Match(cmux.Any())
		httpServer := &http.Server{
			Handler: mux,
		}
		go func() {
			go m.Serve()
			log.Printf("HTTP server crashed: %v", httpServer.Serve(httpL))
		}()
	} else {
		go func() {
			log.Printf("fileserver crashed: %v", http.ListenAndServe(laddr, mux))
		}()
	}

	// start game loop
	go s.gameLoop(errc)

	log.Printf("listening for connections on %v", port)
	for {
		conn, err := l.Accept()
		if err != nil {
			errc <- errors.New("failed to establish connection", err)
			continue
		}
		go func() {
			if err := s.handleConnection(conn); err != nil {
				errc <- errors.New("error handling connection", err)
			}
		}()
	}
}

func (s *mmoServer) handleConnection(conn net.Conn) error {
	defer conn.Close()

	session, err := smux.Server(conn, smux.DefaultConfig())
	if err != nil {
		return err
	}

	stream, err := session.AcceptStream()
	if err != nil {
		return err
	}

	conn = stream

	// read message
	msg, err := shared.GetMessage(conn)
	if err != nil {
		return err
	}

	// check first message is ConnectRequest
	if msg.Request == nil || msg.Request.ConnectRequest == nil {
		return errors.New("expected first message to be ConnectRequest", nil)
	}

	// get ID
	//TODO: instead connectrequest should contain a user/pass combo
	//we should look up the player's existing ID
	// (or generate a new db entry with ID if doesnt exist)
	id := msg.Request.ConnectRequest.ID

	// set up player connection
	if err := s.mgr.playerConnected(id, conn); err != nil {
		log.Printf("WARN: failed to accept connection from player %s at %s\n", id, conn.RemoteAddr())
		return s.mgr.sendError(conn, shared.FatalErr(err))
	}

	log.Printf("new connected player %s from %s", id, conn.RemoteAddr().String())

	// start serving requests from the player
	return s.mgr.startClientLoop(id)
}

// blocks as long as client is connected
func (mgr *updateManager) startClientLoop(id string) error {
	for cli := mgr.getClient(id); cli != nil; {
		msg, err := shared.GetMessage(cli.conn)
		if err != nil {
			log.Print(errors.New(fmt.Sprintf("Client disconnected: (failed getting message for player %s)", cli.player.ID), err))
			if err := mgr.playerDisconnected(id); err != nil {
				return errors.New("Failed to process player disconnected "+id, err)
			}
			return nil
		}
		log.Printf("recv: %s\n%s", msg, id)
		if msg.Request != nil {
			cli.requests <- msg.Request
		} else {
			log.Printf("invalid message from client: %s", msg)
		}
	}
	return nil
}

func (s *mmoServer) gameLoop(errc chan error) {
	tick := time.NewTicker(tickTime)
	last := time.Now()
	for {
		select {
		case now := <-tick.C:
			if err := s.update(now.Sub(last)); err != nil {
				log.Printf("ERROR IN TICK: %v", err)
				errc <- err
			}
		}
		last = time.Now()
	}
}

func (s *mmoServer) update(dt time.Duration) error {
	//copy clients to an array so we dont have to RLock the whole function
	clients := []*client{}
	s.mgr.connectedPlayersLock.RLock()
	for _, cli := range s.mgr.connectedPlayers {
		clients = append(clients, cli)
	}
	s.mgr.connectedPlayersLock.RUnlock()

	for _, cli := range clients {
	requestLoop:
		for {
			select {
			case req := <-cli.requests:
				if err := s.handleRequest(cli.player, req); err != nil {
					log.Printf("Error handling player request %#v: %v", req, err)
				}
			default:
				break requestLoop
			}
		}
	}
	// update world
	if err := s.mgr.world.Step(dt); err != nil {
		return fmt.Errorf("in step: %v", err)
	}

	// only keep the last 3 snapshots
	s.mgr.world.Keep(3)

	// broadcast all updates to clients
	for {
		select {
		default:
			return nil
		case update := <-s.mgr.world.ProcessedUpdates():
			log.Printf("gonna broadcast: %s", update)
			if err := s.mgr.broadcast(&shared.Message{Update: update}); err != nil {
				return errors.New("failed to broadcast update", err)
			}
		}
	}
}

func (s *mmoServer) handleRequest(player *shared.Player, req *shared.Request) error {
	switch {
	case req.MoveRequest != nil:
		return s.mgr.playerMoved(player, req.MoveRequest)
	case req.SpeakRequest != nil:
		return s.mgr.apply(&shared.PlayerSpoke{
			ID:   player.ID,
			Text: req.SpeakRequest.Text,
		})
	}
	return fmt.Errorf("unknown request type: %#v", req)
}
