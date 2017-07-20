package shared

import (
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/ilackarms/pkg/errors"
)

const (
	basePlayerSpeed = 2.0
)

var (
	defaultSize = pixel.V(1, 1)
)

type World struct {
	//needs to be exported to support serialization
	//treat this field as unexported
	Players     map[string]*Player
	playersLock sync.RWMutex
	previous    *World
	updated     time.Time
}

func NewEmptyWorld() *World {
	return &World{
		Players: make(map[string]*Player),
		updated: time.Now(),
	}
}

func (w *World) DeepCopy() *World {
	cpy := NewEmptyWorld()
	if w.previous != nil {
		cpy.previous = w.previous.DeepCopy()
	}
	w.playersLock.RLock()
	defer w.playersLock.RUnlock()
	for id, player := range w.Players {
		cpy.Players[id] = player.DeepCopy()
	}
	return cpy
}

func (w *World) ApplyUpdate(update *Update) (err error) {
	if update.AddPlayer != nil {
		return w.addPlayer(update.AddPlayer)
	}
	if update.PlayerMoved != nil {
		return w.applyPlayerMoved(update.PlayerMoved)
	}
	if update.PlayerSpoke != nil {
		return w.applyPlayerSpoke(update.PlayerSpoke)
	}
	if update.WorldState != nil {
		return w.setWorldState(update.WorldState)
	}
	if update.RemovePlayer != nil {
		return w.applyRemovePlayer(update.RemovePlayer)
	}
	return errors.New("empty update given? wtf", nil)
}

// At returns the most recent world that existed at time t
func (w *World) At(t time.Time) *World {
	if w.updated.Before(t) {
		return w
	}
	if w.previous != nil {
		return w.previous.At(t)
	}
	// be careful of nil pointers
	return nil
}

// Trim trims snapsohts at and before time t
func (w *World) Trim(t time.Time) {
	*(w.At(t)) = nil
}

// process game-world self update
// step wraps the previous state for rolling back
func (w *World) Step(dt time.Duration) (err error) {
	w.previous = w.DeepCopy()
	w.updated = time.Now()
	w.playersLock.Lock()
	defer w.playersLock.Unlock()
	for id, player := range w.Players {
		// update player positions based on speed and destination
		if !WithinRange(player.Destination, player.Position, 0.5) {
			// TODO change this to use astar pathing
			newPos := RoundVec(player.Position.Add(player.Destination.Sub(player.Position).Unit().Scaled(player.Speed*dt.Seconds())), 5)
			//check collisions
			var collisionFound bool
			hitbox := RectFromCenter(newPos, player.Size.X, player.Size.Y)
			for otherID, otherPlayer := range w.Players {
				// player cant collide with self
				if id == otherID {
					continue
				}
				otherHitbox := RectFromCenter(otherPlayer.Position, otherPlayer.Size.X, otherPlayer.Size.Y)
				if hitbox.Intersect(otherHitbox).Area() > 0 {
					collisionFound = true
					break
				}
			}
			if collisionFound {
				continue
			}
			player.Position = newPos
		}
	}
	return nil
}

// GetPlayer returns a referece to player
// PLEASE do not use this reference to modify player directly!
// Objects returned by GetPlayer should be read-only
// Looking forward to go supporting immutable references
func (w *World) GetPlayer(id string) (*Player, bool) {
	player, err := w.getPlayer(id)
	if err != nil {
		return nil, false
	}
	return player, true
}

// ForEach calls f on each player in the world
// PLEASE do not use this to modify player
// This is intended for reading only
func (w *World) ForEach(f func(player *Player)) {
	w.playersLock.RLock()
	defer w.playersLock.RUnlock()
	for _, player := range w.Players {
		f(player)
	}
}

// if player doesnt exist, add. if player is inactive, activate. if player is active, error
func (w *World) addPlayer(added *AddPlayer) error {
	if player, err := w.getPlayer(added.ID); err == nil {
		if player.Active {
			return errors.New("player "+added.ID+" already active!", nil)
		}
		player.Active = true
		return nil
	}
	w.setPlayer(added.ID, &Player{
		ID:           added.ID,
		Position:     added.Position,
		Destination:  added.Position,
		Speed:        basePlayerSpeed,
		Size:         defaultSize,
		SpeechBuffer: []SpeechMesage{},
		Active:       true,
	})
	return nil
}

func (w *World) applyPlayerMoved(moved *PlayerMoved) error {
	player, err := w.getActivePlayer(moved.ID)
	if err != nil {
		return err
	}
	player.Destination = moved.Destination
	return nil
}

func (w *World) applyPlayerSpoke(speech *PlayerSpoke) error {
	id := speech.ID
	player, err := w.getActivePlayer(id)
	if err != nil {
		return err
	}
	txt := player.SpeechBuffer
	// speech  buffer size 4
	if len(txt) > 4 {
		txt = txt[1:]
	}
	txt = append(txt, SpeechMesage{Txt: speech.Text, Timestamp: time.Now()})
	w.setPlayer(id, player)
	return nil
}

func (w *World) setWorldState(worldState *WorldState) error {
	w = worldState.World
	return nil
}

func (w *World) applyRemovePlayer(removed *RemovePlayer) error {
	player, err := w.getActivePlayer(removed.ID)
	if err != nil {
		return err
	}
	player.Active = false
	return nil
}

func (w *World) getActivePlayer(id string) (*Player, error) {
	player, err := w.getPlayer(id)
	if err != nil {
		return nil, err
	}
	if !player.Active {
		return nil, errors.New("player "+id+" requested but inactive", nil)
	}
	return player, nil
}

func (w *World) getPlayer(id string) (*Player, error) {
	w.playersLock.RLock()
	player, ok := w.Players[id]
	w.playersLock.RUnlock()
	if !ok {
		return nil, errors.New("player "+id+" requested but not found", nil)
	}
	return player, nil
}

func (w *World) setPlayer(id string, player *Player) {
	w.playersLock.Lock()
	w.Players[id] = player
	w.playersLock.Unlock()
}
