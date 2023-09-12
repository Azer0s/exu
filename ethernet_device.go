package exu

import (
	"errors"
	"github.com/songgao/packets/ethernet"
	"math/rand"
	"net"
	"sync"
)

type EthernetDevice struct {
	ports        []*VPort
	name         string
	portsMu      sync.RWMutex
	onReceive    func(srcPort *VPort, data ethernet.Frame)
	onConnect    func(port *VPort)
	onDisconnect func(port *VPort)
}

type EthernetReceiver interface {
	OnReceive(srcPort *VPort, src net.HardwareAddr, dst net.HardwareAddr, data ethernet.Frame)
}

func NewEthernetDevice(name string, numberOfPorts int, onReceive func(srcPort *VPort, data ethernet.Frame), onConnect func(port *VPort), onDisconnect func(port *VPort)) *EthernetDevice {
	dev := new(EthernetDevice)
	*dev = EthernetDevice{
		name:         name,
		ports:        make([]*VPort, numberOfPorts),
		portsMu:      sync.RWMutex{},
		onReceive:    onReceive,
		onConnect:    onConnect,
		onDisconnect: onDisconnect,
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

		dev.ports[i] = NewVPort(mac)
		func(i int) {
			dev.ports[i].SetOnReceive(func(data ethernet.Frame) {
				dev.onReceive(dev.ports[i], data)
			})
		}(i)
	}

	return dev
}

func (e *EthernetDevice) WriteFromPort(port *VPort, data ethernet.Frame) error {
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
		data[i] = port.mac[i-6]
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
	e.ports[portOnMachine].connectedTo = target
	target.connectedTo = e.ports[portOnMachine]
	e.onConnect(e.ports[portOnMachine])
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

func (e *EthernetDevice) DisconnectPort(target *VPort) {
	e.portsMu.Lock()
	defer e.portsMu.Unlock()

	for _, port := range e.ports {
		if port.connectedTo == target {
			port.connectedTo = nil
			port.onReceive = nil
			e.onDisconnect(port)
			break
		}
	}
}
