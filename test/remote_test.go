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

	sw1 := exu.NewEthernetSwitch("sw1", 10)

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
	log.SetLevel(log.TraceLevel)

	sw1 := exu.NewEthernetSwitch("sw1", 10)
	disconnectFn := func(p *exu.VPort) {
		log.Info("remote disconnected")
		sw1.DisconnectPort(p)
	}
	connectFn := func(port *exu.VPort) {
		_ = sw1.ConnectToFirstAvailablePort(port)
	}

	_, _ = exu.NewRemoteVport(6554, net.ParseIP("10.0.0.1"), connectFn, disconnectFn)
	_, _ = exu.NewRemoteVport(6555, net.ParseIP("10.0.0.2"), connectFn, disconnectFn)

	select {}
}
