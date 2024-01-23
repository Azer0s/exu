package exu

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"sync"
)

type PortMode int

const (
	Access PortMode = iota
	Trunk
)

type PortModeConfig struct {
	Mode PortMode
	Vlan uint16
}

var PortModeTrunk = PortModeConfig{
	Mode: Trunk,
}

type VSwitch struct {
	*EthernetDevice
	portMode   map[*VPort]PortModeConfig
	portModeMu sync.RWMutex
}

func NewVSwitch(name string, numberOfPorts int) *VSwitch {
	vSwitch := &VSwitch{
		portMode:   make(map[*VPort]PortModeConfig),
		portModeMu: sync.RWMutex{},
	}

	vSwitch.EthernetDevice = NewEthernetDevice(name, numberOfPorts, vSwitch.onReceive, func(*VPort) {}, vSwitch.onDisconnect)

	for i := 0; i < numberOfPorts; i++ {
		vSwitch.portMode[vSwitch.ports[i]] = PortModeConfig{
			Mode: Access,
			Vlan: 1,
		}
	}

	return vSwitch
}

func (s *VSwitch) SetPortMode(port *VPort, mode PortModeConfig) {
	s.portModeMu.Lock()
	defer s.portModeMu.Unlock()

	s.portMode[port] = mode
}

func (s *VSwitch) flood(srcPort *VPort, data *EthernetFrame) {
	s.portsMu.RLock()
	defer s.portsMu.RUnlock()

	for _, port := range s.ports {
		if port != srcPort {
			_ = s.WriteFromPort(port, data)
		}
	}
}

func (s *VSwitch) onReceive(srcPort *VPort, data *EthernetFrame) {
	// TODO: handle vlan
	// Lookup tag of the incoming frame, if it is not the same as the dst port, drop the frame
	// If the tag is the same, forward the frame to the dst port
	// If we don't have the mac address in the mac address table, flood the frame to all trunk
	// ports and all access ports in the same vlan

	// If the dst MAC is a broadcast MAC, flood the frame to all trunk ports and all access ports
	if bytes.Equal(data.Destination(), []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
		s.flood(srcPort, data)
		return
	}

	srcString := data.Source().String()
	dstString := data.Destination().String()

	log.WithField("src", srcString).
		WithField("dst", dstString).
		WithField("device", s.name).
		Trace("received frame")

	s.macAddressMapMu.RLock()
	defer s.macAddressMapMu.RUnlock()

	if dstPort, ok := s.macAddressMap[dstString]; ok {
		_ = s.WriteFromPort(dstPort, data)
		return
	}

	s.flood(srcPort, data)
}

func (s *VSwitch) onDisconnect(port *VPort) {
	s.portModeMu.Lock()
	defer s.portModeMu.Unlock()

	s.portMode[port] = PortModeConfig{
		Mode: Access,
		Vlan: 1,
	}
}
