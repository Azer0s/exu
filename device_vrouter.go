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
	// right now we can only route IPv4 and ARP packets
	if !data.EtherType().Equal(EtherTypeIPv4) && !data.EtherType().Equal(EtherTypeARP) {
		return
	}

	if data.EtherType() == EtherTypeARP {
		// get the ARP payload
		arpPayload := &ArpPayload{}
		err := arpPayload.FromBytes(data.Payload())
		if err != nil {
			return
		}

		// if the ARP packet is for one of our ports, reply with our MAC address
		if r.portIPs[srcPort].IP.Equal(arpPayload.TargetIP) {
			// create the ARP payload
			arpPayload := &ArpPayload{
				HardwareType: arpPayload.HardwareType,
				ProtocolType: arpPayload.ProtocolType,
				Opcode:       ArpOpcodeReply,
				SenderIP:     arpPayload.TargetIP,
				TargetIP:     arpPayload.SenderIP,
				SenderMac:    srcPort.mac,
				TargetMac:    arpPayload.SenderMac,
			}

			// create the ethernet frame
			ethernetFrame, err := NewEthernetFrame(data.Source(), data.Destination(), WithTagging(TaggingUntagged), arpPayload)
			if err != nil {
				return
			}

			// write the frame to the source port
			_ = srcPort.Write(ethernetFrame)
			return
		}
	}

	// get the destination IP
	ipv4Packet := &IPv4Packet{}
	err := ipv4Packet.UnmarshalBinary(data.Payload())
	if err != nil {
		return
	}

	// check if the packet is ICMP
	// if so, it could be for us
	if ipv4Packet.Header.Protocol == IPv4ProtocolICMP {
		icmpPayload := &ICMPPayload{}
		err = icmpPayload.FromBytes(ipv4Packet.Payload)
		if err != nil {
			return
		}

		// if the packet is for one of our ports, reply with an ICMP echo reply
		if r.portIPs[srcPort].IP.Equal(ipv4Packet.Header.DestinationIP) {
			// create the ICMP payload
			icmpResponsePayload := &ICMPPayload{
				Type: ICMPTypeEchoReply,
				Code: 0,
				Data: icmpPayload.Data,
			}
			icmpResponsePayload.Checksum = icmpResponsePayload.CalculateChecksum()

			icmpResponsePayloadBytes, _ := icmpResponsePayload.MarshalBinary()

			ipv4ResponsePacket := &IPv4Packet{
				Header: IPv4Header{
					Version:        4,
					IHL:            5,
					TOS:            0,
					TotalLength:    uint16(20 + len(icmpResponsePayloadBytes)),
					ID:             0,
					FlagsFragment:  0,
					TTL:            64,
					Protocol:       IPv4ProtocolICMP,
					HeaderChecksum: 0,
					SourceIP:       r.portIPs[srcPort].IP,
					DestinationIP:  ipv4Packet.Header.SourceIP,
				},
				Payload: icmpResponsePayloadBytes,
			}
			ipv4ResponsePacket.Header.HeaderChecksum = ipv4ResponsePacket.Header.CalculateChecksum()

			// create the ethernet frame
			ethernetFrame, err := NewEthernetFrame(data.Source(), data.Destination(), WithTagging(TaggingUntagged), ipv4ResponsePacket)
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
