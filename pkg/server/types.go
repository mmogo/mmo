package server

import (
	"net"

	"github.com/mmogo/mmo/pkg/shared"
)

// the server's wrapper for a Player object
// contains player info specific to the server
type client struct {
	player   *shared.Player
	conn     net.Conn
	requests chan *shared.Request
}

func newServerPlayer(player *shared.Player, conn net.Conn) *client {
	return &client{
		player:   player,
		conn:     conn,
		requests: make(chan *shared.Request, bufferedMessageLimit),
	}
}
