package main

import (
	"github.com/ilackarms/_anything/shared"
	"sync"
)

var (
	playersLock = sync.RWMutex{}
	players     = make(map[string]*shared.ServerPlayer)

	updatesLock = sync.Mutex{}
	updates     = []*update{}
)
