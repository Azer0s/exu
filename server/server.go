package server

import (
	"encoding/binary"
	"exu/vswitch"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"sync"
)

var ipRanges = []string{
	"10.0.0.1/24",
	"192.168.0.1/24",
}
var usedIPs = make(map[string]bool)
var usedIPsMu = sync.Mutex{}

func init() {
	_, subnet, _ := net.ParseCIDR(ipRanges[0])
	x := subnet.IP.To4()
	for i := 0; i < 255; i++ {
		x[3] = x[3] + 1
		usedIPs[x.String()] = false
	}
}

func New(mode Mode) Server {
	log.WithField("mode", mode).Info("starting server")

	// create a new UDP socket
	tcp, err := net.Listen("tcp", ":6885")
	if err != nil {
		panic(err)
	}

	srv := Server{
		tcp:  tcp,
		mode: mode,
	}

	switch mode {
	case ModeVSwitch:
		srv.handler = vswitch.New()
	}

	return srv
}

func (s Server) Run() {
	for {
		conn, err := s.tcp.Accept()
		log.WithField("remote_addr", conn.RemoteAddr()).Info("accepted new connection")

		if err != nil {
			log.WithField("error", err).
				WithField("remote_addr", conn.RemoteAddr()).
				Error("failed to accept connection")
			continue
		}

		// read initial packet
		// first byte is the magic hello byte (0x42)
		// n + 1 byte is the remote port (2 bytes)

		magic := make([]byte, 1)
		_, err = conn.Read(magic)
		if err != nil {
			log.WithField("error", err).
				WithField("remote_addr", conn.RemoteAddr()).
				Error("failed to read magic byte")
			continue
		}

		if magic[0] != 0x42 {
			log.WithField("remote_addr", conn.RemoteAddr()).Error("invalid magic byte")
			continue
		}

		portBytes := make([]byte, 2)
		_, err = conn.Read(portBytes)
		if err != nil {
			panic(err)
		}
		port := binary.LittleEndian.Uint16(portBytes)

		log.WithField("remote_addr", conn.RemoteAddr()).
			WithField("port", port).
			Debug("received initial packet")

		rx := conn
		tx, err := net.Dial("tcp", strings.Split(rx.RemoteAddr().String(), ":")[0]+":"+strconv.Itoa(int(port)))
		if err != nil {
			log.WithField("error", err).
				WithField("remote_addr", conn.RemoteAddr()).
				Error("failed to connect to remote port")
			continue
		}

		log.WithField("remote_addr", conn.RemoteAddr()).
			WithField("port", port).
			Debug("connected to remote tx port")

		// get a new ip address for the client
		usedIPsMu.Lock()
		var first net.IP
		for ip, used := range usedIPs {
			if !used {
				first = net.ParseIP(ip)
				usedIPs[ip] = true
				break
			}
		}
		usedIPsMu.Unlock()

		// send ip address to client
		ipBytes := first.To4()
		_, err = tx.Write([]byte{ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3]})
		if err != nil {
			log.WithField("error", err).
				WithField("remote_addr", conn.RemoteAddr()).
				Error("failed to send ip address to client")

			usedIPsMu.Lock()
			usedIPs[first.String()] = false
			usedIPsMu.Unlock()

			tErr, rErr := tx.Close(), rx.Close()
			if tErr != nil {
				log.WithField("error", tErr).
					WithField("remote_addr", conn.RemoteAddr()).
					Error("failed to close tx connection")
			}

			if rErr != nil {
				log.WithField("error", rErr).
					WithField("remote_addr", conn.RemoteAddr()).
					Error("failed to close rx connection")
			}

			continue
		}

		log.WithField("remote_addr", conn.RemoteAddr()).
			WithField("ip", first.String()).
			Debug("sent ip address to client")

		go func(ip net.IP) {
			s.handler.Handle(rx, tx)
			usedIPs[ip.String()] = false
		}(first)
	}
}
