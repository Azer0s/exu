package exu

import log "github.com/sirupsen/logrus"

type CapabilityForwardArp struct {
	*IpDevice
}

func (c CapabilityForwardArp) HandleRequest(port *VPort, data *EthernetFrame) CapabilityStatus {
	// get the ARP payload
	arpPayload := &ArpPacket{}
	err := arpPayload.FromBytes(data.Payload())
	if err != nil {
		return CapabilityStatusFail
	}

	// if the ARP packet is not for one of our ports, forward it
	// look in the MAC address table for the destination port
	c.macAddressMapMu.RLock()
	dstPort, ok := c.macAddressMap[arpPayload.TargetMac.String()]
	c.macAddressMapMu.RUnlock()

	log.WithFields(log.Fields{
		"device":    c.name,
		"port":      port.portCname,
		"ip":        arpPayload.TargetIP,
		"capabilty": "forward_arp",
	}).Trace("forwarding ARP packet")

	// if we don't have the MAC address in our map, flood the frame to all ports
	if !ok {
		for _, p := range c.ports {
			if p != port {
				_ = c.WriteFromPort(p, data)
			}
		}
		return CapabilityStatusDone
	}

	// if we have the MAC address in our map, forward the frame to the destination port
	_ = c.WriteFromPort(dstPort, data)
	return CapabilityStatusDone
}

func (c CapabilityForwardArp) Match(port *VPort, data *EthernetFrame) bool {
	if data.EtherType() == EtherTypeARP {
		// get the ARP payload
		arpPayload := &ArpPacket{}
		err := arpPayload.FromBytes(data.Payload())
		if err != nil {
			return false
		}

		// if the ARP packet is not for one of our ports, forward it
		if !c.portIPs[port].IP.Equal(arpPayload.TargetIP) {
			return true
		}
	}

	return false
}
