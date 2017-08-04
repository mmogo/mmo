package debug

import (
	"github.com/faiface/pixel/pixelgl"
	"net"
	"github.com/mmogo/mmo/pkg/shared"
)

type client struct {
	conn            net.Conn
	win             *pixelgl.Window
	playerID        string
	world           *shared.World
	pongs           chan *shared.Pong
	updates         chan *shared.Update
	requests        chan *shared.Request
	inProcessor     *inputProcessor
	reqProcessor    *requestProcessor
	errc            chan error
	bufferedUpdates UpdateBuffer
}

func NewClient(id string, conn net.Conn, world *shared.World) *client {
	requests := make(chan *shared.Request, maxBufferedRequests)
	updates := make(chan *shared.Update, maxBufferedUpdates)

	return &client{
		conn:         conn,
		playerID:     id,
		win:          win,
		world:        world,
		requests:     requests,
		updates:      updates,
		inProcessor:  newInputProcessor(win, requests, screen2Map, &cam),
		reqProcessor: newRequestManager(id, requests, predictions, conn),
		errc:         make(chan error),
		pongs:        make(chan *shared.Pong),
	}
}
