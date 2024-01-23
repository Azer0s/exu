package test

import (
	"exu"
	"net"
	"testing"
)

func TestSimpleRouting(t *testing.T) {
	// v1 <-> (0) r1 (1) <-> (0) r2 (1) <-> v2
	// v1: 10.0.0.0/24
	// v2: 172.0.0.0/24
	// r1 - r2: 192.168.0.0/24
	// r1 routes:
	//  - 0.0.0.0/0 -> port 1
	//  - 10.0.0.0/24 -> port 0
	// r2 routes:
	//  - 0.0.0.0/0 -> port 0
	//  - 172.0.0.0/24 -> port 1

	// create a new router
	r1 := exu.NewEthernetRouter("r1", 2)
	r2 := exu.NewEthernetRouter("r2", 2)

	// create two virtual devices
	v1Chan := make(chan *exu.EthernetFrame)
	v1 := exu.NewIpDevice("v1", 1, func(srcPort *exu.VPort, data *exu.EthernetFrame) {
		v1Chan <- data
	}, func(port *exu.VPort) {}, func(port *exu.VPort) {})
	v1Port := v1.GetFirstFreePort()
	v1.SetPortIPNet(v1Port, net.IPNet{
		IP:   net.IPv4(10, 0, 0, 2),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	})

	v2 := exu.NewIpDevice("v2", 1, func(srcPort *exu.VPort, data *exu.EthernetFrame) {}, func(port *exu.VPort) {}, func(port *exu.VPort) {})
	v2Port := v2.GetFirstFreePort()
	v2.SetPortIPNet(v2Port, net.IPNet{
		IP:   net.IPv4(172, 0, 0, 2),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	})

	// connect the ports to the routers
	v1R1Port := r1.GetFirstFreePort()
	err := r1.ConnectPorts(v1R1Port, v1Port)
	if err != nil {
		t.Fatal(err)
	}

	v2R2Port := r2.GetFirstFreePort()
	err = r2.ConnectPorts(v2R2Port, v2Port)
	if err != nil {
		t.Fatal(err)
	}

	r1r2Port := r2.GetFirstFreePort()
	r2r1Port := r1.GetFirstFreePort()
	err = r2.ConnectPorts(r1r2Port, r2r1Port)
	if err != nil {
		t.Fatal(err)
	}

	// set the port IPs
	r1.SetPortIPNet(v1R1Port, net.IPNet{
		IP:   net.IPv4(10, 0, 0, 1),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	})
	r2.SetPortIPNet(v2R2Port, net.IPNet{
		IP:   net.IPv4(172, 0, 0, 1),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	})
	r1.SetPortIPNet(r1r2Port, net.IPNet{
		IP:   net.IPv4(192, 168, 0, 1),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	})
	r2.SetPortIPNet(r2r1Port, net.IPNet{
		IP:   net.IPv4(192, 168, 0, 2),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	})

	// set the routes on r1
	err = r1.AddRoute(exu.Route{
		Network: net.IPNet{
			IP:   net.IPv4(0, 0, 0, 0),
			Mask: net.IPv4Mask(0, 0, 0, 0),
		},
		Via: net.IPv4(192, 168, 0, 2),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = r1.AddRoute(exu.Route{
		Network: net.IPNet{
			IP:   net.IPv4(10, 0, 0, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// set the routes on r2
	err = r2.AddRoute(exu.Route{
		Network: net.IPNet{
			IP:   net.IPv4(0, 0, 0, 0),
			Mask: net.IPv4Mask(0, 0, 0, 0),
		},
		Via: net.IPv4(192, 168, 0, 1),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = r2.AddRoute(exu.Route{
		Network: net.IPNet{
			IP:   net.IPv4(172, 0, 0, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	icmpPayload := &exu.ICMPPayload{
		Type: exu.ICMPTypeEcho,
		Code: 0,
		Data: []byte("hello"),
	}
	icmpPayload.Checksum = icmpPayload.CalculateChecksum()
	icmpPayloadBytes, _ := icmpPayload.MarshalBinary()

	ipv4Packet := &exu.IPv4Packet{
		Header: exu.IPv4Header{
			Version:       4,
			IHL:           5,
			TOS:           0,
			ID:            0,
			TTL:           64,
			Protocol:      exu.IPv4ProtocolICMP,
			SourceIP:      net.IPv4(10, 0, 0, 2),
			DestinationIP: net.IPv4(172, 0, 0, 2),
		},
		Payload: icmpPayloadBytes,
	}
	ipv4Packet.Header.TotalLength = uint16(20 + len(ipv4Packet.Payload))
	ipv4Packet.Header.HeaderChecksum = ipv4Packet.Header.CalculateChecksum()
}
