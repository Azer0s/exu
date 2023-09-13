package exu

import (
	"errors"
	"github.com/songgao/packets/ethernet"
	"net"
)

var VPortNotConnectedError = errors.New("vPort is not connected")

type VPort struct {
	mac         net.HardwareAddr
	connectedTo *VPort
	onReceive   func(data ethernet.Frame)
}

func (v *VPort) SetOnReceive(onReceive func(data ethernet.Frame)) {
	v.onReceive = onReceive
}

func (v *VPort) Write(data ethernet.Frame) error {
	if v.connectedTo == nil {
		return VPortNotConnectedError
	}

	WithWaitGroup(func() {
		v.connectedTo.onReceive(data)
	})
	return nil
}

func NewVPort(mac net.HardwareAddr) *VPort {
	return &VPort{
		mac: mac,
	}
}
