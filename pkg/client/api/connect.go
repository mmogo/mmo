package api

import (
	"log"
	"net"

	"github.com/mmogo/mmo/pkg/shared"
	"github.com/xtaci/smux"
	"github.com/ilackarms/pkg/errors"
)

func Dial(protocol, addr, id string) (net.Conn, *shared.World, error) {
	log.Printf("dialing %s", addr)
	conn, err := shared.Dial(protocol, addr)
	if err != nil {
		return nil, nil, err
	}
	session, err := smux.Client(conn, smux.DefaultConfig())
	if err != nil {
		return nil, nil, err
	}
	stream, err := session.OpenStream()
	if err != nil {
		return nil, nil, err
	}
	conn = stream

	if err := shared.SendMessage(&shared.Message{
		Request: &shared.Request{
			ConnectRequest: &shared.ConnectRequest{
				ID: id,
			},
		}}, conn); err != nil {
		return nil, nil, err
	}

	
	// sync with server
	msg, err := shared.GetMessage(conn)
	if err != nil {
		return nil, nil, errors.New("failed reading message", err)
	}

	if msg.Update == nil || msg.Update.WorldState == nil || msg.Update.WorldState.World == nil {
		return nil, nil, errors.New("expected Sync message on server handshake, got "+msg.String(), nil)
	}
	
	return conn, msg.Update.WorldState.World, nil
}
