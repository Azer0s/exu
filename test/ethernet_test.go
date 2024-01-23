package test

import (
	"bytes"
	"exu"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net"
	"regexp"
	"testing"
)

func TestEthernetSwitchLearnMac(t *testing.T) {
	// create a new bytes buffer
	buff := bytes.NewBuffer(make([]byte, 4096))
	log.SetOutput(buff)

	sw1 := exu.NewVSwitch("sw1", 10)
	sw2 := exu.NewVSwitch("sw2", 10)

	err := sw1.ConnectToFirstAvailablePort(sw2.GetFirstFreePort())
	if err != nil {
		return
	}

	p1 := exu.NewVPort(net.HardwareAddr{0x42, 0x69, 0x00, 0x00, 0x00, 0x01}, "p1")
	p1.SetOnReceive(func(data *exu.EthernetFrame) {
		log.WithField("data", string(data.Payload())).
			Info("received data on p1")
	})

	p2 := exu.NewVPort(net.HardwareAddr{0x42, 0x69, 0x00, 0x00, 0x00, 0x02}, "p2")
	p2.SetOnReceive(func(data *exu.EthernetFrame) {
		log.WithField("data", string(data.Payload())).
			Info("received data on p2")
		returnFrame := exu.EthernetFrame{
			0x42, 0x69, 0x00, 0x00, 0x00, 0x01,
			0x42, 0x69, 0x00, 0x00, 0x00, 0x02,
			0x10, 0x01,
			0x48, 0x65, 0x6c, 0x6c, 0x6f,
		}

		_ = p2.Write(&returnFrame)
	})

	_ = sw1.ConnectToFirstAvailablePort(p1)
	_ = sw2.ConnectToFirstAvailablePort(p2)

	// construct a new ethernet frame
	// first 6 bytes are the destination mac address
	// second 6 bytes are the source mac address
	// last 2 bytes are the ethernet type
	// the rest is the payload
	frame := exu.EthernetFrame{
		0x42, 0x69, 0x00, 0x00, 0x00, 0x02,
		0x42, 0x69, 0x00, 0x00, 0x00, 0x01,
		0x10, 0x01,
		0x48, 0x65, 0x6c, 0x6c, 0x6f,
	}

	_ = p1.Write(&frame)

	exu.AllSettled()

	assert.Contains(t, buff.String(), "received data on p2")
	assert.Contains(t, buff.String(), "received data on p1")

	// test how often the buffer contains "learned new mac address"
	rgx := regexp.MustCompile(`learned new mac address`)
	matches := rgx.FindAllStringIndex(buff.String(), -1)
	assert.Equal(t, 4, len(matches))
}

func TestEthernetRouter(t *testing.T) {
	// create a new bytes buffer
	buff := bytes.NewBuffer(make([]byte, 4096))
	log.SetOutput(buff)

	r1 := exu.NewEthernetRouter("r1", 10)
	sw1 := exu.NewVSwitch("sw1", 10)

	sw1Port := sw1.GetFirstFreePort()
	err := r1.ConnectToFirstAvailablePort(sw1Port)
	if err != nil {
		t.Fatal(err)
	}

	sw1.SetPortMode(sw1Port, exu.PortModeTrunk)

	// TODO: create a static route

	p1 := exu.NewVPort(net.HardwareAddr{0x42, 0x69, 0x00, 0x00, 0x00, 0x01}, "p1")
	p1.SetOnReceive(func(data *exu.EthernetFrame) {
		log.WithField("data", string(data.Payload())).
			Info("received data on p1")
	})

	p2 := exu.NewVPort(net.HardwareAddr{0x42, 0x69, 0x00, 0x00, 0x00, 0x02}, "p2")
	p2.SetOnReceive(func(data *exu.EthernetFrame) {
		log.WithField("data", string(data.Payload())).
			Info("received data on p2")
		returnFrame := exu.EthernetFrame{
			0x42, 0x69, 0x00, 0x00, 0x00, 0x01,
			0x42, 0x69, 0x00, 0x00, 0x00, 0x02,
			0x10, 0x01,
			0x48, 0x65, 0x6c, 0x6c, 0x6f,
		}

		_ = p2.Write(&returnFrame)
	})

	_ = r1.ConnectToFirstAvailablePort(p1)
	_ = sw1.ConnectToFirstAvailablePort(p2)
}
