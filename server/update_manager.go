package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/ilackarms/pkg/errors"
	"github.com/mmogo/mmo/shared"
)

// update manager handles all updates
// update manager makes sure that updates are duplicated properly
// for the server's internal state, and broadcast to clients
// who are expected to apply updates to their internal state
type updateManager struct {
	world                *shared.World
	connectedPlayers     map[string]*client
	connectedPlayersLock sync.RWMutex
}

func newUpdateManager() *updateManager {
	return &updateManager{
		world:            shared.NewEmptyWorld(),
		connectedPlayers: make(map[string]*client),
	}
}

/*
	Utility functions
*/
func (mgr *updateManager) getClient(id string) *client {
	mgr.connectedPlayersLock.RLock()
	defer mgr.connectedPlayersLock.RUnlock()
	return mgr.connectedPlayers[id]
}

/*
	Messaging Stuff
*/
func (mgr *updateManager) sendError(conn net.Conn, err error) error {
	if err == nil {
		return errors.New("cannot send nil error!", nil)
	}
	return shared.SendMessage(&shared.Message{
		Error: &shared.Error{Message: err.Error()}}, conn)
}

func (mgr *updateManager) send(id string, msg *shared.Message) error {
	log.Printf("sending to %s: %s", id, msg)
	cli := mgr.getClient(id)
	err := shared.SendMessage(msg, cli.conn)
	if err != nil {
		//disconnect player
		log.Printf("failed to send update to connected player %s; disconnecting client: %v", id, err)
		mgr.playerDisconnected(id)
	}
	return err
}

func (mgr *updateManager) broadcast(msg *shared.Message) error {
	log.Printf("broadcasting: %s", msg)
	data, err := shared.Encode(msg)
	if err != nil {
		return err
	}
	mgr.connectedPlayersLock.RLock()
	defer mgr.connectedPlayersLock.RUnlock()
	for id, player := range mgr.connectedPlayers {
		if err := shared.SendRaw(data, player.conn); err != nil {
			defer func(id string) {
				//disconnect player
				log.Printf("failed to send update to connected player %s; disconnecting client: %v", id, err)
				mgr.playerDisconnected(id)
			}(id)
		}
	}
	return nil
}

func (mgr *updateManager) applyAndBroadcast(updateContents interface{}) error {
	update := &shared.Update{}
	switch contents := updateContents.(type) {
	case *shared.AddPlayer:
		update.AddPlayer = contents
	case *shared.RemovePlayer:
		update.RemovePlayer = contents
	case *shared.PlayerMoved:
		update.PlayerMoved = contents
	case *shared.PlayerSpoke:
		update.PlayerSpoke = contents
	default:
		return fmt.Errorf("unknown update type: %#v", updateContents)
	}
	if err := mgr.world.ApplyUpdate(update); err != nil {
		return fmt.Errorf("failed to apply update %v: %v", update, err)
	}

	if err := mgr.broadcast(&shared.Message{Update: update}); err != nil {
		return errors.New("failed to broadcast update", err)
	}
	return nil
}

func (mgr *updateManager) syncPlayerState(id string) error {
	// sync client state
	if err := mgr.send(id, &shared.Message{Update: &shared.Update{WorldState: &shared.WorldState{World: mgr.world}}}); err != nil {
		return errors.New("syncing state with client", err)
	}
	return nil
}

/*
	Event handlers
*/
func (mgr *updateManager) playerConnected(id string, conn net.Conn) error {
	if cli := mgr.getClient(id); cli != nil {
		return fmt.Errorf("Player %s already connected", id)
	}

	if err := mgr.applyAndBroadcast(&shared.AddPlayer{
		ID: id,
		// todo: dont pick random starting positions. rework how collisions work
		Position: shared.RandVec(-20, 20),
	}); err != nil {
		return errors.New("failed to apply and broadcast adding of player", err)
	}

	player, ok := mgr.world.GetPlayer(id)
	if !ok {
		return fmt.Errorf("player %s should have been added to state but was not", id)
	}

	sPlayer := newServerPlayer(player, conn)
	mgr.connectedPlayersLock.Lock()
	mgr.connectedPlayers[id] = sPlayer
	mgr.connectedPlayersLock.Unlock()

	// sync client state
	if err := mgr.syncPlayerState(id); err != nil {
		return errors.New("failed to initialize client state", err)
	}

	return nil
}

func (mgr *updateManager) playerDisconnected(id string) error {
	mgr.connectedPlayersLock.Lock()
	delete(mgr.connectedPlayers, id)
	mgr.connectedPlayersLock.Unlock()

	return mgr.applyAndBroadcast(&shared.RemovePlayer{
		ID: id,
	})
}

func (mgr *updateManager) playerMoved(player *shared.Player, move *shared.MoveRequest) error {
	if shared.UnitVec(player.Direction) == shared.UnitVec(move.Direction) {
		//no-op, ignore this request
		return nil
	}
	moveUpdate := shared.ToUpdate(player.ID, move).PlayerMoved
	return mgr.applyAndBroadcast(moveUpdate)
}
