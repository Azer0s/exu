package test

import (
	"exu"
	log "github.com/sirupsen/logrus"
	"net"
	"testing"
)

func TestRemoteV_Port(t *testing.T) {
	done := make(chan bool)
	log.SetLevel(log.TraceLevel)

	sw1 := exu.NewVSwitch("sw1", 10)

	_, _ = exu.NewRemoteVport(6554, net.ParseIP("10.0.0.1"), func(port *exu.VPort) {
		_ = sw1.ConnectToFirstAvailablePort(port)
	}, func(p *exu.VPort) {
		log.Info("remote disconnected")
		sw1.DisconnectPort(p)
		done <- true
	})

	<-done
}

func TestRemoteV_Port2(t *testing.T) {
	log.SetLevel(log.InfoLevel)

	sw1 := exu.NewVSwitch("sw1", 10)
	disconnectFn := func(p *exu.VPort) {
		log.Info("remote disconnected")
		sw1.DisconnectPort(p)
	}
	connectFn := func(port *exu.VPort) {
		_ = sw1.ConnectToFirstAvailablePort(port)
	}

	go func() {
		_, _ = exu.NewRemoteVport(6554, net.ParseIP("10.0.0.1"), connectFn, disconnectFn)
	}()

	go func() {
		_, _ = exu.NewRemoteVport(6555, net.ParseIP("10.0.0.2"), connectFn, disconnectFn)
	}()

	select {}
}

func TestRemoteV_Port3(t *testing.T) {
	// this is an icmp test of the VRouter
	log.SetLevel(log.InfoLevel)

	r1 := exu.NewEthernetRouter("r1", 10)
	p1 := r1.GetFirstFreePort()
	r1.SetPortIPNet(p1, net.IPNet{
		IP:   net.IPv4(10, 0, 0, 1),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	})

	_, _ = exu.NewRemoteVport(6554, net.ParseIP("10.0.0.2"), func(port *exu.VPort) {
		_ = r1.ConnectPorts(p1, port)
	}, func(p *exu.VPort) {
		log.Info("remote disconnected")
		r1.DisconnectPort(p1)
	})

	select {}
}
