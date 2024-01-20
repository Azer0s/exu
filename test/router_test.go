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

	// create two ports
	v1 := exu.NewVPort(net.HardwareAddr{0x42, 0x69, 0x00, 0x00, 0x00, 0x01})
	v2 := exu.NewVPort(net.HardwareAddr{0x42, 0x69, 0x00, 0x00, 0x00, 0x02})

	// connect the ports to the routers
	v1R1Port := r1.GetFirstFreePort()
	err := r1.ConnectPorts(v1R1Port, v1)
	if err != nil {
		t.Fatal(err)
	}

	v2R2Port := r2.GetFirstFreePort()
	err = r2.ConnectPorts(v2R2Port, v2)
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
		Interface: v1R1Port,
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
		Interface: v2R2Port,
	})
	if err != nil {
		t.Fatal(err)
	}
}
