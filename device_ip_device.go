package exu

import (
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type IpDevice struct {
	*EthernetDevice
	portIPs        map[*VPort]net.IPNet
	onReceiveIp    func(srcPort *VPort, data *EthernetFrame)
	onConnectIp    func(port *VPort)
	onDisconnectIp func(port *VPort)
	arpTableMu     sync.RWMutex
	arpTable       map[string]net.HardwareAddr
}

func NewIpDevice(name string, numberOfPorts int, onReceive func(srcPort *VPort, data *EthernetFrame), onConnect func(port *VPort), onDisconnect func(port *VPort)) *IpDevice {
	ipDevice := &IpDevice{
		portIPs:        make(map[*VPort]net.IPNet),
		arpTable:       make(map[string]net.HardwareAddr),
		arpTableMu:     sync.RWMutex{},
		onReceiveIp:    onReceive,
		onConnectIp:    onConnect,
		onDisconnectIp: onDisconnect,
	}

	ipDevice.EthernetDevice = NewEthernetDevice(name, numberOfPorts, ipDevice.onReceive, func(*VPort) {}, ipDevice.onDisconnect)

	return ipDevice
}

func (d *IpDevice) SetPortIPNet(port *VPort, ipNet net.IPNet) {
	d.portIPs[port] = ipNet
	log.WithFields(log.Fields{
		"device": d.name,
		"port":   port.portCname,
		"ip":     ipNet.IP,
	}).Debug("set port IP")
}

func (d *IpDevice) handleArpResponseForLocalPort(srcPort *VPort, data *EthernetFrame, arpPayload *ArpPayload) bool {
	// if this is an ARP response, check if it's for one of our ports
	if arpPayload.Opcode == ArpOpcodeReply && arpPayload.TargetIP.Equal(d.portIPs[srcPort].IP) {
		d.arpTableMu.Lock()
		defer d.arpTableMu.Unlock()

		d.arpTable[arpPayload.SenderIP.String()] = arpPayload.SenderMac

		log.WithFields(log.Fields{
			"device":      d.name,
			"port":        srcPort.portCname,
			"ip":          arpPayload.SenderIP,
			"learned_mac": arpPayload.SenderMac,
		}).Debug("learned ARP entry")
		return true
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
			return true
		}

		// write the frame to the source port
		_ = srcPort.Write(ethernetFrame)

		log.WithFields(log.Fields{
			"device": d.name,
			"port":   srcPort.portCname,
		}).Debug("sent ARP response")
		return true
	}
	return false
}

func (d *IpDevice) handleIcmpResponseForLocalPort(srcPort *VPort, data *EthernetFrame, icmpPayload *ICMPPayload, ipv4Packet *IPv4Packet) {
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

func (d *IpDevice) onReceive(srcPort *VPort, data *EthernetFrame) {
	if data.EtherType() == EtherTypeARP {
		// get the ARP payload
		arpPayload := &ArpPayload{}
		err := arpPayload.FromBytes(data.Payload())
		if err != nil {
			return
		}

		if d.handleArpResponseForLocalPort(srcPort, data, arpPayload) {
			return
		}

		// if the ARP packet is not for one of our ports, forward it
		// look in the MAC address table for the destination port
		d.macAddressMapMu.RLock()
		dstPort, ok := d.macAddressMap[arpPayload.TargetMac.String()]
		d.macAddressMapMu.RUnlock()

		log.WithFields(log.Fields{
			"device": d.name,
			"port":   srcPort.portCname,
			"ip":     arpPayload.TargetIP,
		}).Trace("forwarding ARP packet")

		// if we don't have the MAC address in our map, flood the frame to all ports
		if !ok {
			for _, port := range d.ports {
				if port != srcPort {
					_ = d.WriteFromPort(port, data)
				}
			}
			return
		}

		// if we have the MAC address in our map, forward the frame to the destination port
		_ = d.WriteFromPort(dstPort, data)
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
				log.WithFields(log.Fields{
					"device":    d.name,
					"port":      srcPort.portCname,
					"source_ip": ipv4Packet.Header.SourceIP,
				}).Debug("received ICMP packet")
				d.handleIcmpResponseForLocalPort(srcPort, data, icmpPayload, ipv4Packet)
				return
			}
		}
	}

	d.onReceiveIp(srcPort, data)
}

func (d *IpDevice) onDisconnect(port *VPort) {
	d.onDisconnectIp(port)
}

func (d *IpDevice) ArpResolve(requested net.IP) (net.HardwareAddr, error) {
	log.WithFields(log.Fields{
		"device": d.name,
		"ip":     requested,
	}).Debug("resolving ARP")

	// check if we already have the MAC address in our ARP table
	if mac, ok := d.arpTable[requested.String()]; ok {
		return mac, nil
	}

	// check which port is in the same network as the requested IP
	ports := make([]*VPort, 0)
	for p, ipNet := range d.portIPs {
		networkAddress := net.IPNet{
			IP:   ipNet.IP.Mask(ipNet.Mask),
			Mask: ipNet.Mask,
		}

		if networkAddress.Contains(requested) {
			ports = append(ports, p)
		}
	}

	// if we don't have a port in the same network, we just ask every port
	if len(ports) == 0 {
		ports = d.ports
	}

	for _, port := range ports {
		arpRequestPayload := &ArpPayload{
			HardwareType: ArpHardwareTypeEthernet,
			ProtocolType: ArpProtocolTypeIPv4,
			Opcode:       ArpOpcodeRequest,
			SenderIP:     d.portIPs[d.ports[0]].IP,
			TargetIP:     requested,
			SenderMac:    d.ports[0].mac,
			TargetMac:    net.HardwareAddr{0, 0, 0, 0, 0, 0},
		}

		// create the ethernet frame
		ethernetFrame, err := NewEthernetFrame(net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, port.mac, WithTagging(TaggingUntagged), arpRequestPayload)
		if err != nil {
			return nil, err
		}

		// write the frame to the port
		_ = port.Write(ethernetFrame)
	}

	// wait for the ARP table to be updated with a timeout of 5 seconds
	resultChan := make(chan net.HardwareAddr)
	go func() {
		for i := 0; i < 50; i++ {
			d.arpTableMu.RLock()
			mac, ok := d.arpTable[requested.String()]
			d.arpTableMu.RUnlock()
			if ok {
				resultChan <- mac
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		resultChan <- nil
	}()

	select {
	case mac := <-resultChan:
		return mac, nil
		//case <-time.After(5 * time.Second):
		//	return nil, errors.New("timed out waiting for ARP response")
	}
}
