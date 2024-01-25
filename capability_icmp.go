package exu

import log "github.com/sirupsen/logrus"

type CapabilityIcmp struct {
	*IpDevice
}

func (c CapabilityIcmp) HandleRequest(port *VPort, data *EthernetFrame) CapabilityStatus {
	ipv4Packet := &IPv4Packet{}
	err := ipv4Packet.UnmarshalBinary(data.Payload())
	if err != nil {
		return CapabilityStatusFail
	}

	icmpPayload := &ICMPPayload{}
	err = icmpPayload.UnmarshalBinary(ipv4Packet.Payload)
	if err != nil {
		return CapabilityStatusFail
	}

	// if the packet is for one of our ports, reply with an ICMP echo reply
	if c.portIPs[port].IP.Equal(ipv4Packet.Header.DestinationIP) {
		log.WithFields(log.Fields{
			"device":    c.name,
			"port":      port.portCname,
			"source_ip": ipv4Packet.Header.SourceIP,
			"capabilty": "icmp",
		}).Debug("received ICMP packet")

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
				SourceIP:       c.portIPs[port].IP,
				DestinationIP:  ipv4Packet.Header.SourceIP,
			},
			Payload: icmpResponsePayloadBytes,
		}
		ipv4ResponsePacket.Header.HeaderChecksum = ipv4ResponsePacket.Header.CalculateChecksum()

		// create the ethernet frame
		ethernetFrame, err := NewEthernetFrame(data.Source(), data.Destination(), WithTagging(TaggingUntagged), ipv4ResponsePacket)
		if err != nil {
			return CapabilityStatusFail
		}

		// write the frame to the source port
		_ = port.Write(ethernetFrame)

		return CapabilityStatusDone
	}

	return CapabilityStatusPass
}

func (c CapabilityIcmp) Match(_ *VPort, data *EthernetFrame) bool {
	if data.EtherType().Equal(EtherTypeIPv4) {
		ipv4Packet := &IPv4Packet{}
		err := ipv4Packet.UnmarshalBinary(data.Payload())
		if err != nil {
			return false
		}

		if ipv4Packet.Header.Protocol == IPv4ProtocolICMP {
			icmpPayload := &ICMPPayload{}
			err = icmpPayload.UnmarshalBinary(ipv4Packet.Payload)
			if err != nil {
				return false
			}

			return true
		}
	}

	return false
}
