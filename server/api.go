package server

import (
	"net"
)

type Mode int

const (
	ModeVSwitch Mode = iota
)

type Server struct {
	tcp  net.Listener
	mode Mode
}

type Handler interface {
	Handle(rx, tx net.Conn)
}
