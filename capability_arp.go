package exu

import log "github.com/sirupsen/logrus"

type CapabilityArp struct {
	*IpDevice
}

func (c CapabilityArp) HandleRequest(port *VPort, data *EthernetFrame) CapabilityStatus {
	// get the ARP payload
	arpPayload := &ArpPacket{}
	err := arpPayload.FromBytes(data.Payload())
	if err != nil {
		return CapabilityStatusFail
	}

	// if this is an ARP response, check if it's for one of our ports
	if arpPayload.Opcode == ArpOpcodeReply && arpPayload.TargetIP.Equal(c.portIPs[port].IP) {
		c.arpTableMu.Lock()
		defer c.arpTableMu.Unlock()

		c.arpTable[arpPayload.SenderIP.String()] = arpPayload.SenderMac

		log.WithFields(log.Fields{
			"device":      c.name,
			"port":        port.portCname,
			"ip":          arpPayload.SenderIP,
			"learned_mac": arpPayload.SenderMac,
			"capabilty":   "arp",
		}).Debug("learned ARP entry")
		return CapabilityStatusDone
	}

	// if the ARP packet is for one of our ports, reply with our MAC address
	if c.portIPs[port].IP.Equal(arpPayload.TargetIP) {
		// create the ARP payload
		arpResponsePayload := &ArpPacket{
			HardwareType: arpPayload.HardwareType,
			ProtocolType: arpPayload.ProtocolType,
			Opcode:       ArpOpcodeReply,
			SenderIP:     arpPayload.TargetIP,
			TargetIP:     arpPayload.SenderIP,
			SenderMac:    port.mac,
			TargetMac:    arpPayload.SenderMac,
		}

		// create the ethernet frame
		var ethernetFrame *EthernetFrame
		ethernetFrame, err = NewEthernetFrame(data.Source(), data.Destination(), WithTagging(TaggingUntagged), arpResponsePayload)
		if err != nil {
			return CapabilityStatusFail
		}

		// write the frame to the source port
		_ = port.Write(ethernetFrame)

		log.WithFields(log.Fields{
			"device":    c.name,
			"port":      port.portCname,
			"capabilty": "arp",
		}).Debug("sent ARP response")
		return CapabilityStatusDone
	}

	return CapabilityStatusPass
}

func (c CapabilityArp) Match(_ *VPort, data *EthernetFrame) bool {
	if data.EtherType() != EtherTypeARP {
		return false
	}
	return true
}
