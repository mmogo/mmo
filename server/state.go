package main

import (
	"github.com/ilackarms/_anything/shared/types"
	"sync"
)

var (
	playersLock = sync.RWMutex{}
	players     = make(map[string]*types.ServerPlayer)

	updatesLock = sync.Mutex{}
	updates     = []*update{}
)
