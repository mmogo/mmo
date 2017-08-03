// Copyright 2017 The MMOGO Authors. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

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

func (reqProcessor *requestProcessor) processPending(errc chan error) {
	for {
		select {
		case req := <-reqProcessor.pendingRequests:
			if err := reqProcessor.handleRequest(req); err != nil {
				errc <- fmt.Errorf("Error handling player request %#v: %v", req, err)
			}
		}
	}
}

func (reqProcessor *requestProcessor) handleRequest(req *shared.Request) error {
	if err := shared.SendMessage(&shared.Message{Request: req}, reqProcessor.conn); err != nil {
		return errors.New("failed to send request", err)
	}
	switch {
	case req.MoveRequest != nil:
		reqProcessor.updatePredictions <- shared.ToUpdate(reqProcessor.playerID, req.MoveRequest)
		log.Printf("update prediction: %v", req.MoveRequest.Destination)
	case req.SpeakRequest != nil:
		reqProcessor.updatePredictions <- shared.ToUpdate(reqProcessor.playerID, req.SpeakRequest)
	default:
		return fmt.Errorf("unknown request type: %#v", req)
	}
	return nil
}
