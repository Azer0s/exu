package exu

import (
	"bytes"
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
	defaultRoute *Route
}

func NewEthernetRouter(name string, numberOfPorts int) *EthernetRouter {
	ethernetRouter := &EthernetRouter{
		portIPs: make(map[*VPort]net.IPNet),
	}

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

func (r *EthernetRouter) onReceive(srcPort *VPort, data *EthernetFrame) {
	// right now we can only route IPv4 packets
	if data.EtherType() != EtherTypeIPv4 {
		return
	}

	// get the destination IP
	ipv4Payload := &IPv4Payload{}
	err := ipv4Payload.FromBytes(data.Payload())
	if err != nil {
		return
	}

	// check if the packet is ICMP
	// if so, it could be for us
	if ipv4Payload.Protocol == IPv4ProtocolICMP {
		icmpPayload := &ICMPPayload{}
		err = icmpPayload.FromBytes(ipv4Payload.Data)
		if err != nil {
			return
		}

		// if the packet is for one of our ports, reply with an ICMP echo reply
		if r.portIPs[srcPort].IP.Equal(ipv4Payload.DestinationIP) {
			// create the ICMP payload
			icmpPayload := &ICMPPayload{
				Type: ICMPTypeEchoReply,
				Code: 0,
				Data: icmpPayload.Data,
			}

			// create the IPv4 payload
			ipv4Payload := &IPv4Payload{
				Version:        4,
				IHL:            5,
				DSCP:           0,
				ECN:            0,
				TotalLength:    uint16(20 + 8 + len(icmpPayload.Data)),
				Identification: 0,
				Flags:          0,
				FragmentOffset: 0,
				TTL:            64,
				Protocol:       IPv4ProtocolICMP,
				SourceIP:       r.portIPs[srcPort].IP,
				DestinationIP:  ipv4Payload.SourceIP,
			}

			// create the ethernet frame
			ethernetFrame, err := NewEthernetFrame(data.Source(), data.Destination(), WithTagging(TaggingUntagged), ipv4Payload)
			if err != nil {
				return
			}

			// write the frame to the source port
			_ = srcPort.Write(ethernetFrame)
			return
		}
	}

}

func (r *EthernetRouter) onDisconnect(port *VPort) {

}
