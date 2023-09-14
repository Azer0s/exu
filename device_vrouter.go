package exu

import (
	"net"
)

type Route struct {
	Network   net.IPNet
	Netmask   net.IPMask
	Gateway   net.IP
	Interface *VPort
}

type EthernetRouter struct {
	*EthernetDevice
	routingTable []Route
	portIPs      map[*VPort]net.IP
}

func NewEthernetRouter(name string, numberOfPorts int) *EthernetRouter {
	ethernetRouter := &EthernetRouter{}

	ethernetRouter.EthernetDevice = NewEthernetDevice(name, numberOfPorts, ethernetRouter.onReceive, func(*VPort) {}, ethernetRouter.onDisconnect)

	return ethernetRouter
}

func (r *EthernetRouter) SetPortIP(port *VPort, ip net.IP) {

}

func (r *EthernetRouter) onReceive(srcPort *VPort, data *EthernetFrame) {

}

func (r *EthernetRouter) onDisconnect(port *VPort) {

}
