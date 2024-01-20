package exu

import (
	"errors"
	"net"
)

type Route struct {
	Network   net.IPNet
	Via       net.IP
	Interface *VPort
}

type EthernetRouter struct {
	*EthernetDevice
	routingTable []Route
	portIPs      map[*VPort]net.IPNet
}

func NewEthernetRouter(name string, numberOfPorts int) *EthernetRouter {
	ethernetRouter := &EthernetRouter{}

	ethernetRouter.EthernetDevice = NewEthernetDevice(name, numberOfPorts, ethernetRouter.onReceive, func(*VPort) {}, ethernetRouter.onDisconnect)

	return ethernetRouter
}

func (r *EthernetRouter) SetPortIPNet(port *VPort, ipNet net.IPNet) {
	r.portIPs[port] = ipNet

}

func (r *EthernetRouter) AddRoute(route Route) error {
	// either the via or the interface must be set
	if route.Via == nil && route.Interface == nil {
		return errors.New("either the via or the interface must be set")
	}

	return nil
}

func (r *EthernetRouter) onReceive(srcPort *VPort, data *EthernetFrame) {

}

func (r *EthernetRouter) onDisconnect(port *VPort) {

}
