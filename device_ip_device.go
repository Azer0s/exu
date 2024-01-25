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
	ipDevice.EthernetDevice.capabilities = append(ipDevice.EthernetDevice.capabilities, CapabilityIcmp{
		IpDevice: ipDevice,
	})
	ipDevice.EthernetDevice.capabilities = append(ipDevice.EthernetDevice.capabilities, CapabilityArp{
		IpDevice: ipDevice,
	})

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

func (d *IpDevice) onReceive(srcPort *VPort, data *EthernetFrame) {
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
		arpRequestPayload := &ArpPacket{
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
