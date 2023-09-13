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

	tcpAddr := net.TCPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}

	// create a new remote interface client
	client := exu.NewRemoteInterfaceClient(tcpAddr)
	client.Run()
}
