package shared

import (
	"log"
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
	// these are used for creating a series of buffered world snapshots
	//automatically created on each step
	previous *World
	Updated  time.Time
	// processed is for updates that have been processed
	processed chan *Update
}

func NewEmptyWorld() *World {
	return &World{
		Players:   make(map[string]*Player),
		processed: make(chan *Update),
		Updated:   time.Now(),
	}
}

func (w *World) ProcessedUpdates() <-chan *Update {
	return w.processed
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

func (w *World) finishUpdate(update *Update) {
	update.Processed = time.Now()
	go func() {
		w.processed <- update
	}()
}

func (w *World) ApplyUpdates(updates ...*Update) error {
	for _, update := range updates {
		log.Printf("applying %s", update.String())
		if err := w.applyUpdate(update); err != nil {
			return err
		}
		w.finishUpdate(update)
	}
	return nil
}

func (w *World) applyUpdate(update *Update) error {
	if update.AddPlayer != nil {
		return w.addPlayer(update.AddPlayer)
	}
	if update.PlayerDestination != nil {
		return w.updateDestination(update.PlayerDestination)
	}
	if update.PlayerPosition != nil {
		return w.updatePosition(update.PlayerPosition)
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

func (w *World) Len() int {
	if w.previous == nil {
		return 1
	}
	return w.previous.Len() + 1
}

// Before returns the most recent world that existed before t
func (w *World) Before(t time.Time) *World {
	log.Printf("%s\n%s (%v left)", t, w.Updated, w.Len())
	if w.Updated.Before(t) {
		return w
	}
	if w.previous != nil {
		return w.previous.Before(t)
	}
	// if no world existed before t, just return the earliest available world
	return w
}

func (w *World) Prev() *World {
	return w.previous
}

// Trim trims snapsohts at and before time t
func (w *World) Trim(t time.Time) {
	trimFrom := w.Before(t)
	prev := w.previous
	next := w
	for prev != nil {
		if prev == trimFrom {
			next.previous = nil
		}
		tmp := prev.previous
		next = prev
		prev = tmp
	}
}

// Keep drops all snapshots after the nth
func (w *World) Keep(n int) {
	if n == 1 {
		w.previous = nil
	}
	// n should always be >= 1
	if n < 1 {
		return
	}
	w.Keep(n - 1)
}

// process game-world self update
// step wraps the previous state for rolling back
func (w *World) Step(dt time.Duration) (err error) {
	w.previous = w.DeepCopy()
	w.Updated = time.Now()
	w.playersLock.Lock()
	defer w.playersLock.Unlock()
	for id, player := range w.Players {
		// update player positions based on speed and destination
		if !WithinRange(player.Destination, player.Position, 1) {
			// TODO change this to use astar pathing
			newPos := RoundVec(player.Position.Add(player.Destination.Sub(player.Position).Unit().Scaled(player.Speed*dt.Seconds())), 0)
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
			// on new player position, send internal update
			w.finishUpdate(&Update{PlayerPosition: &PlayerPosition{ID: player.ID, Position: newPos}})
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

func (w *World) updateDestination(dest *PlayerDestination) error {
	player, err := w.getActivePlayer(dest.ID)
	if err != nil {
		return err
	}
	player.Destination = dest.Destination
	return nil
}

func (w *World) updatePosition(moved *PlayerPosition) error {
	player, err := w.getActivePlayer(moved.ID)
	if err != nil {
		return err
	}
	player.Position = moved.Position
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
