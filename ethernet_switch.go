package exu

import (
	log "github.com/sirupsen/logrus"
	"github.com/songgao/packets/ethernet"
	"sync"
)

type EthernetSwitch struct {
	*EthernetDevice
	macAddressMap   map[string]*VPort
	macAddressMapMu sync.RWMutex
}

func NewEthernetSwitch(name string, numberOfPorts int) *EthernetSwitch {
	ethernetSwitch := &EthernetSwitch{
		macAddressMap:   make(map[string]*VPort),
		macAddressMapMu: sync.RWMutex{},
	}

	ethernetSwitch.EthernetDevice = NewEthernetDevice(name, numberOfPorts, ethernetSwitch.onReceive, func(*VPort) {}, ethernetSwitch.onDisconnect)
	return ethernetSwitch
}

func (s *EthernetSwitch) flood(srcPort *VPort, data ethernet.Frame) {
	s.portsMu.RLock()
	defer s.portsMu.RUnlock()

	for _, port := range s.ports {
		if port != srcPort {
			_ = s.WriteFromPort(port, data)
		}
	}
}

func (s *EthernetSwitch) onReceive(srcPort *VPort, data ethernet.Frame) {
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
	s.macAddressMapMu.Lock()
	defer s.macAddressMapMu.Unlock()

	for mac, p := range s.macAddressMap {
		if p == port {
			delete(s.macAddressMap, mac)
		}
	}
}
