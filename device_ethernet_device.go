package exu

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"strconv"
	"sync"
)

type EthernetDevice struct {
	name string

	ports   []*VPort
	portsMu sync.RWMutex

	macAddressMap   map[string]*VPort
	macAddressMapMu sync.RWMutex

	onReceiveFn    func(srcPort *VPort, data *EthernetFrame)
	onConnectFn    func(port *VPort)
	onDisconnectFn func(port *VPort)
}

type EthernetReceiver interface {
	OnReceive(srcPort *VPort, src net.HardwareAddr, dst net.HardwareAddr, data EthernetFrame)
}

func NewEthernetDevice(name string, numberOfPorts int, onReceive func(srcPort *VPort, data *EthernetFrame), onConnect func(port *VPort), onDisconnect func(port *VPort)) *EthernetDevice {
	dev := new(EthernetDevice)
	*dev = EthernetDevice{
		name:            name,
		ports:           make([]*VPort, numberOfPorts),
		portsMu:         sync.RWMutex{},
		macAddressMap:   make(map[string]*VPort),
		macAddressMapMu: sync.RWMutex{},
		onReceiveFn:     onReceive,
		onConnectFn:     onConnect,
		onDisconnectFn:  onDisconnect,
	}

	for i := 0; i < numberOfPorts; i++ {
		mac := make(net.HardwareAddr, 6)
		mac[0] = 0x42
		mac[1] = 0x69
		r := rand.Uint32()
		mac[2] = byte(r >> 24)
		mac[3] = byte(r >> 16)
		mac[4] = byte(r >> 8)
		mac[5] = byte(r)

		dev.ports[i] = NewVPort(mac, "eth0/"+strconv.Itoa(i))
		func(i int) {
			dev.ports[i].SetOnReceive(func(data *EthernetFrame) {
				// check if we already have the source MAC address in our map
				dev.macAddressMapMu.RLock()
				_, ok := dev.macAddressMap[data.Source().String()]
				dev.macAddressMapMu.RUnlock()

				// if we don't have the source MAC address in our map, add it
				if !ok && data.Source().String() != "ff:ff:ff:ff:ff:ff" {
					dev.macAddressMapMu.Lock()
					dev.macAddressMap[data.Source().String()] = dev.ports[i]
					dev.macAddressMapMu.Unlock()

					log.WithField("mac", data.Source().String()).
						WithField("port", dev.ports[i].portCname).
						WithField("device", dev.name).
						Trace("learned new mac address")
				}

				dev.onReceiveFn(dev.ports[i], data)
			})
		}(i)
	}

	return dev
}

func (e *EthernetDevice) WriteFromPort(port *VPort, data *EthernetFrame) error {
	e.portsMu.RLock()
	defer e.portsMu.RUnlock()

	found := false
	for _, p := range e.ports {
		if p == port {
			found = true
			break
		}
	}

	if !found {
		return errors.New("port not found")
	}

	for i := 6; i < 12; i++ {
		(*data)[i] = port.mac[i-6]
	}

	return port.Write(data)
}

func (e *EthernetDevice) GetFirstFreePort() *VPort {
	e.portsMu.Lock()
	defer e.portsMu.Unlock()

	for _, port := range e.ports {
		if port.connectedTo == nil {
			return port
		}
	}

	return nil
}

func (e *EthernetDevice) connectPorts(portOnMachine int, target *VPort) {
	log.WithField("device", e.name).
		WithField("port", target.mac.String()).
		Info("connected port")

	e.ports[portOnMachine].connectedTo = target
	target.connectedTo = e.ports[portOnMachine]
	e.onConnectFn(e.ports[portOnMachine])
}

func (e *EthernetDevice) ConnectToFirstAvailablePort(target *VPort) error {
	e.portsMu.Lock()
	defer e.portsMu.Unlock()

	for i, port := range e.ports {
		if port.connectedTo == nil {
			e.connectPorts(i, target)
			return nil
		}
	}

	return errors.New("no available ports")
}

func (e *EthernetDevice) ConnectPorts(portOnMachine, target *VPort) error {
	e.portsMu.Lock()
	defer e.portsMu.Unlock()

	index := -1
	for i, port := range e.ports {
		if port == portOnMachine {
			index = i
			break
		}
	}

	if index == -1 {
		return errors.New("port not found on machine")
	}

	if e.ports[index].connectedTo != nil {
		return errors.New("port already connected")
	}

	e.connectPorts(index, target)
	return nil
}

func (e *EthernetDevice) DisconnectPort(target *VPort) {
	log.WithField("device", e.name).
		WithField("port", target.mac.String()).
		Info("disconnecting port")

	e.portsMu.Lock()
	defer e.portsMu.Unlock()

	for _, port := range e.ports {
		if port.connectedTo == target {
			port.connectedTo = nil
			port.onReceive = nil
			e.onDisconnectFn(port)
			break
		}
	}

	e.macAddressMapMu.Lock()
	defer e.macAddressMapMu.Unlock()

	for mac, port := range e.macAddressMap {
		if port == target {
			delete(e.macAddressMap, mac)
			break
		}
	}
}
