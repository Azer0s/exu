package exu

import (
	"errors"
	"net"
)

var VPortNotConnectedError = errors.New("vPort is not connected")

type VPort struct {
	mac         net.HardwareAddr
	connectedTo *VPort
	onReceive   func(data *EthernetFrame)
}

func (v *VPort) SetOnReceive(onReceive func(data *EthernetFrame)) {
	v.onReceive = onReceive
}

func (v *VPort) Write(data *EthernetFrame) error {
	if v.connectedTo == nil {
		return VPortNotConnectedError
	}

	WithWaitGroup(func() {
		v.connectedTo.onReceive(data)
	})
	return nil
}

func (v *VPort) Mac() net.HardwareAddr {
	return v.mac
}

func NewVPort(mac net.HardwareAddr) *VPort {
	return &VPort{
		mac: mac,
	}
}
