package main

import "github.com/faiface/pixel"

type (
	update struct {
		notifyPlayerMoved        *notifyPlayerMoved
		notifyPlayerSpoke        *notifyPlayerSpoke
		notifyWorldState         *notifyWorldState
		notifyPlayerDisconnected *notifyPlayerDisconnected
	}

	notifyPlayerMoved struct {
		id          string
		newPosition pixel.Vec
	}

	notifyPlayerSpoke struct {
		id   string
		text string
	}

	notifyWorldState struct {
		targetID string
	}

	notifyPlayerDisconnected struct {
		id string
	}
)
