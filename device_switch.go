package exu

import (
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

type EthernetSwitch struct {
	*EthernetDevice
	macAddressMap   map[string]*VPort
	macAddressMapMu sync.RWMutex
	portMode        map[*VPort]PortModeConfig
	portModeMu      sync.RWMutex
}

func NewEthernetSwitch(name string, numberOfPorts int) *EthernetSwitch {
	ethernetSwitch := &EthernetSwitch{
		macAddressMap:   make(map[string]*VPort),
		macAddressMapMu: sync.RWMutex{},
		portMode:        make(map[*VPort]PortModeConfig),
		portModeMu:      sync.RWMutex{},
	}

	ethernetSwitch.EthernetDevice = NewEthernetDevice(name, numberOfPorts, ethernetSwitch.onReceive, func(*VPort) {}, ethernetSwitch.onDisconnect)

	for i := 0; i < numberOfPorts; i++ {
		ethernetSwitch.portMode[ethernetSwitch.ports[i]] = PortModeConfig{
			Mode: Access,
			Vlan: 1,
		}
	}

	return ethernetSwitch
}

func (s *EthernetSwitch) SetPortMode(port *VPort, mode PortModeConfig) {
	s.portModeMu.Lock()
	defer s.portModeMu.Unlock()

	s.portMode[port] = mode
}

func (s *EthernetSwitch) flood(srcPort *VPort, data *EthernetFrame) {
	s.portsMu.RLock()
	defer s.portsMu.RUnlock()

	for _, port := range s.ports {
		if port != srcPort {
			_ = s.WriteFromPort(port, data)
		}
	}
}

func (s *EthernetSwitch) onReceive(srcPort *VPort, data *EthernetFrame) {
	// TODO: handle vlan
	// Lookup tag of the incoming frame, if it is not the same as the dst port, drop the frame
	// If the tag is the same, forward the frame to the dst port
	// If we don't have the mac address in the mac address table, flood the frame to all trunk
	// ports and all access ports in the same vlan

	srcString := data.Source().String()
	dstString := data.Destination().String()

	log.WithField("src", srcString).
		WithField("dst", dstString).
		WithField("name", s.name).
		Trace("received frame")

	func() {
		s.macAddressMapMu.Lock()
		defer s.macAddressMapMu.Unlock()

		if _, ok := s.macAddressMap[srcString]; !ok {
			s.macAddressMap[srcString] = srcPort

			idx := 0
			for _, p := range s.macAddressMap {
				if p == s.macAddressMap[srcString] {
					break
				}

				idx++
			}

			log.WithField("mac", srcString).
				WithField("port", idx).
				WithField("name", s.name).
				Info("learned new mac address")
		}
	}()

	s.macAddressMapMu.RLock()
	defer s.macAddressMapMu.RUnlock()

	if dstPort, ok := s.macAddressMap[dstString]; ok {
		_ = s.WriteFromPort(dstPort, data)
		return
	}

	s.flood(srcPort, data)
}

func (s *EthernetSwitch) onDisconnect(port *VPort) {
	s.portModeMu.Lock()
	defer s.portModeMu.Unlock()

	s.macAddressMapMu.Lock()
	defer s.macAddressMapMu.Unlock()

	s.portMode[port] = PortModeConfig{
		Mode: Access,
		Vlan: 1,
	}

	for mac, p := range s.macAddressMap {
		if p == port {
			delete(s.macAddressMap, mac)
		}
	}
}
