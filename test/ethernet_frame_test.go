package test

import (
	"encoding/hex"
	"exu"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func mustParseMAC(mac string) net.HardwareAddr {
	parsed, err := net.ParseMAC(mac)
	if err != nil {
		panic(err)
	}
	return parsed
}

func TestArpRequest(t *testing.T) {
	sender := mustParseMAC("42:69:00:00:00:02")

	frame, err := exu.NewEthernetFrame(
		exu.BroadcastMAC,
		sender,
		exu.WithTagging(exu.TaggingUntagged),
		exu.NewArpPayload(
			exu.ArpHardwareTypeEthernet,
			exu.ArpProtocolTypeIPv4,
			exu.ArpOpcodeRequest,
			sender,
			net.IPv4(10, 0, 0, 2),
			exu.ArpMacBroadcast,
			net.IPv4(10, 0, 0, 1),
		),
	)

	assert.NoError(t, err)
	fmt.Printf("%s", hex.Dump(*frame))
}
