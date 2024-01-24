package exu

import (
	"bytes"
	"errors"
	"net"
)

type Route struct {
	Network net.IPNet
	Via     net.IP
}

type VRouter struct {
	*IpDevice
	routingTable []Route
	defaultRoute *Route
}

func NewVRouter(name string, numberOfPorts int) *VRouter {
	ethernetRouter := &VRouter{
		routingTable: make([]Route, 0),
	}
	ethernetRouter.IpDevice = NewIpDevice(name, numberOfPorts, ethernetRouter.onReceive, func(*VPort) {}, ethernetRouter.onDisconnect)

	return ethernetRouter
}

func (r *VRouter) AddRoute(route Route) error {
	// if the network is 0.0.0.0/0, set it as the default route
	if route.Network.IP.Equal(net.IPv4zero) && bytes.Equal(route.Network.Mask, net.IPv4Mask(0, 0, 0, 0)) {
		if r.defaultRoute != nil {
			return errors.New("default route already set")
		}

		r.defaultRoute = &route
		return nil
	}

	r.routingTable = append(r.routingTable, route)
	return nil
}

func (r *VRouter) onReceive(_ *VPort, data *EthernetFrame) {
	// right now we can only route IPv4 and ARP packets
	if !data.EtherType().Equal(EtherTypeIPv4) {
		return
	}

	// get the destination IP
	ipv4Packet := &IPv4Packet{}
	err := ipv4Packet.UnmarshalBinary(data.Payload())
	if err != nil {
		return
	}

	bestRoute := r.defaultRoute
	for _, route := range r.routingTable {
		// find a more fitting route
		if route.Network.Contains(ipv4Packet.Header.DestinationIP) {
			bestRoute = &route
			break
		}
	}

	// find an interface that is in the network of the next hop of the best route
	var nextHopPort *VPort
	for port, ipNet := range r.portIPs {
		networkIP := net.IPNet{
			IP:   ipNet.IP.Mask(ipNet.Mask),
			Mask: ipNet.Mask,
		}

		if networkIP.Contains(bestRoute.Via) {
			nextHopPort = port
			break
		}
	}

	// if we don't have a next hop port, we can't route the packet
	if nextHopPort == nil {
		return
	}

	// decrement the TTL
	ipv4Packet.Header.TTL--

	// if the TTL is 0, drop the packet
	if ipv4Packet.Header.TTL == 0 {
		return
	}

	// recalculate the checksum
	ipv4Packet.Header.HeaderChecksum = 0
	ipv4Packet.Header.HeaderChecksum = ipv4Packet.Header.CalculateChecksum()

	// create the ethernet frame
	ethernetFrame, _ := NewEthernetFrame(nextHopPort.mac, data.Source(), TagData{}, ipv4Packet)
	err = nextHopPort.Write(ethernetFrame)
	if err != nil {
		return
	}
}

func (r *VRouter) onDisconnect(*VPort) {

}
