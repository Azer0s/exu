package test

import (
	"exu"
	log "github.com/sirupsen/logrus"
	"net"
	"testing"
)

func TestRemoteVport(t *testing.T) {
	done := make(chan bool)
	log.SetLevel(log.TraceLevel)

	sw1 := exu.NewEthernetSwitch("sw1", 10)

	remote, _ := exu.NewRemoteVport(6554, net.ParseIP("10.0.0.1"), func(p *exu.VPort) {
		log.Info("remote disconnected")
		sw1.DisconnectPort(p)
		done <- true
	})
	_ = sw1.ConnectToFirstAvailablePort(remote)

	<-done
}
