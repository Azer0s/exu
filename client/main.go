package main

import (
	"exu"
	"net"
	"os"
	"strconv"
)

func main() {
	host := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	udpAddr := net.UDPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}

	// create a new remote interface client
	client := exu.NewRemoteInterfaceClient(udpAddr)
	client.Run()
}
