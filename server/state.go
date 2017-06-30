package main

import (
	"github.com/mmogo/mmo/shared"
	"sync"
)

var (
	playersLock = sync.RWMutex{}
	players     = make(map[string]*shared.ServerPlayer)

	updatesLock = sync.Mutex{}
	updates     = []*update{}
)
