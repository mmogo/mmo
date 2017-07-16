package main

import (
	"fmt"
	"log"

	"net"

	"github.com/ilackarms/pkg/errors"
	"github.com/mmogo/mmo/shared"
)

type requestProcessor struct {
	playerID          string
	pendingRequests   <-chan *shared.Request
	updatePredictions chan *shared.Update
	conn              net.Conn
}

func newRequestManager(playerID string, pendingRequests <-chan *shared.Request, updatePredictions chan *shared.Update, conn net.Conn) *requestProcessor {
	return &requestProcessor{
		playerID:          playerID,
		pendingRequests:   pendingRequests,
		updatePredictions: updatePredictions,
		conn:              conn,
	}
}

func (reqProcessor *requestProcessor) processPending() error {
requestLoop:
	for {
		select {
		default:
			break requestLoop
		case req := <-reqProcessor.pendingRequests:
			if err := reqProcessor.handleRequest(req); err != nil {
				log.Printf("Error handling player request %#v: %v", req, err)
			}
		}
	}
	return nil
}

func (reqProcessor *requestProcessor) handleRequest(req *shared.Request) error {
	if err := shared.SendMessage(&shared.Message{Request: req}, reqProcessor.conn); err != nil {
		return errors.New("failed to send request", err)
	}
	switch {
	case req.MoveRequest != nil:
		reqProcessor.updatePredictions <- shared.ToUpdate(reqProcessor.playerID, req.MoveRequest)
	case req.SpeakRequest != nil:
		reqProcessor.updatePredictions <- shared.ToUpdate(reqProcessor.playerID, req.SpeakRequest)
	default:
		return fmt.Errorf("unknown request type: %#v", req)
	}
	return nil
}
