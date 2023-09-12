package test

import (
	"bytes"
	"exu"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/packets/ethernet"
	"github.com/stretchr/testify/assert"
	"net"
	"regexp"
	"testing"
	"time"
)

func TestEthernetSwitchLearnMac(t *testing.T) {
	// create a new bytes buffer
	buff := bytes.NewBuffer(make([]byte, 4096))
	log.SetOutput(buff)

	sw1 := exu.NewEthernetSwitch("sw1", 10)
	sw2 := exu.NewEthernetSwitch("sw2", 10)

	err := sw1.ConnectToFirstAvailablePort(sw2.GetFirstFreePort())
	if err != nil {
		return
	}

	p1 := exu.NewVPort(net.HardwareAddr{0x42, 0x69, 0x00, 0x00, 0x00, 0x01})
	p1.SetOnReceive(func(data ethernet.Frame) {
		log.WithField("data", string(data.Payload())).
			Info("received data on p1")
	})

	p2 := exu.NewVPort(net.HardwareAddr{0x42, 0x69, 0x00, 0x00, 0x00, 0x02})
	p2.SetOnReceive(func(data ethernet.Frame) {
		log.WithField("data", string(data.Payload())).
			Info("received data on p2")
		returnFrame := ethernet.Frame{
			0x42, 0x69, 0x00, 0x00, 0x00, 0x01,
			0x42, 0x69, 0x00, 0x00, 0x00, 0x02,
			0x10, 0x01,
			0x48, 0x65, 0x6c, 0x6c, 0x6f,
		}

		_ = p2.Write(returnFrame)
	})

	_ = sw1.ConnectToFirstAvailablePort(p1)
	_ = sw2.ConnectToFirstAvailablePort(p2)

	// construct a new ethernet frame
	// first 6 bytes are the destination mac address
	// second 6 bytes are the source mac address
	// last 2 bytes are the ethernet type
	// the rest is the payload
	frame := ethernet.Frame{
		0x42, 0x69, 0x00, 0x00, 0x00, 0x02,
		0x42, 0x69, 0x00, 0x00, 0x00, 0x01,
		0x10, 0x01,
		0x48, 0x65, 0x6c, 0x6c, 0x6f,
	}

	_ = p1.Write(frame)

	<-time.After(5 * time.Second)

	assert.Contains(t, buff.String(), "received data on p2")
	assert.Contains(t, buff.String(), "received data on p1")

	// test how often the buffer contains "learned new mac address"
	rgx := regexp.MustCompile(`learned new mac address`)
	matches := rgx.FindAllStringIndex(buff.String(), -1)
	assert.Equal(t, 4, len(matches))
}
