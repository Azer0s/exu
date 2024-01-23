package exu

import "net"

type IpDevice struct {
	*EthernetDevice
	portIPs        map[*VPort]net.IPNet
	onReceiveIp    func(srcPort *VPort, data *EthernetFrame)
	onConnectIp    func(port *VPort)
	onDisconnectIp func(port *VPort)
}

func NewIpDevice(name string, numberOfPorts int, onReceive func(srcPort *VPort, data *EthernetFrame), onConnect func(port *VPort), onDisconnect func(port *VPort)) *IpDevice {
	ipDevice := &IpDevice{
		portIPs:        make(map[*VPort]net.IPNet),
		onReceiveIp:    onReceive,
		onConnectIp:    onConnect,
		onDisconnectIp: onDisconnect,
	}

	ipDevice.EthernetDevice = NewEthernetDevice(name, numberOfPorts, ipDevice.onReceive, func(*VPort) {}, ipDevice.onDisconnect)

	return ipDevice
}

func (d *IpDevice) SetPortIPNet(port *VPort, ipNet net.IPNet) {
	d.portIPs[port] = ipNet
}

func (d *IpDevice) onReceive(srcPort *VPort, data *EthernetFrame) {
	if data.EtherType() == EtherTypeARP {
		// get the ARP payload
		arpPayload := &ArpPayload{}
		err := arpPayload.FromBytes(data.Payload())
		if err != nil {
			return
		}

		// if the ARP packet is for one of our ports, reply with our MAC address
		if d.portIPs[srcPort].IP.Equal(arpPayload.TargetIP) {
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

	// check if the packet is ICMP
	// if so, it could be for us
	if data.EtherType().Equal(EtherTypeIPv4) {
		ipv4Packet := &IPv4Packet{}
		err := ipv4Packet.UnmarshalBinary(data.Payload())
		if err != nil {
			return
		}

		if ipv4Packet.Header.Protocol == IPv4ProtocolICMP {
			icmpPayload := &ICMPPayload{}
			err = icmpPayload.FromBytes(ipv4Packet.Payload)
			if err != nil {
				return
			}

			// if the packet is for one of our ports, reply with an ICMP echo reply
			if d.portIPs[srcPort].IP.Equal(ipv4Packet.Header.DestinationIP) {
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
						SourceIP:       d.portIPs[srcPort].IP,
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

	d.onReceiveIp(srcPort, data)
}

func (d *IpDevice) onDisconnect(port *VPort) {
	d.onDisconnectIp(port)
}
