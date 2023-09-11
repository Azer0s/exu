package vswitch

import (
	"sync"
)

func New() *Handler {
	return &Handler{
		connections:   make(map[string]*VPort),
		connectionsMu: sync.RWMutex{},
	}
}
