package exu

import (
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/packets/ethernet"
	"net"
	"strconv"
	"strings"
	"sync"
)

func NewRemoteVport(rxPort int, ip net.IP, disconnectFn func(port *VPort)) (*VPort, error) {
	log.WithField("ip", ip.String()).
		WithField("rxPort", rxPort).
		Info("creating new remote vport")

	// create a new TCP socket
	tcp, err := net.Listen("tcp", ":"+strconv.Itoa(rxPort))
	if err != nil {
		panic(err)
	}

	conn, err := tcp.Accept()
	log.WithField("remote_addr", conn.RemoteAddr()).Info("accepted new connection")

	if err != nil {
		log.WithField("error", err).
			WithField("remote_addr", conn.RemoteAddr()).
			Error("failed to accept connection")
		return nil, err
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
		return nil, err
	}

	if magic[0] != 0x42 {
		log.WithField("remote_addr", conn.RemoteAddr()).Error("invalid magic byte")
		return nil, errors.New("invalid magic byte")
	}

	portBytes := make([]byte, 2)
	_, err = conn.Read(portBytes)
	if err != nil {
		panic(err)
	}
	port := binary.LittleEndian.Uint16(portBytes)

	macBytes := make([]byte, 6)
	_, err = conn.Read(macBytes)
	if err != nil {
		log.WithField("error", err).
			WithField("remote_addr", conn.RemoteAddr()).
			Error("failed to read mac address")
		return nil, err
	}

	log.WithField("remote_addr", conn.RemoteAddr()).
		WithField("port", port).
		Debug("received initial packet")

	rx := conn
	tx, err := net.Dial("tcp", strings.Split(rx.RemoteAddr().String(), ":")[0]+":"+strconv.Itoa(int(port)))
	if err != nil {
		log.WithField("error", err).
			WithField("remote_addr", conn.RemoteAddr()).
			Error("failed to connect to remote port")
		return nil, err
	}

	log.WithField("remote_addr", conn.RemoteAddr()).
		WithField("port", port).
		Debug("connected to remote tx port")

	// send ip address to client
	ipBytes := ip.To4()
	_, err = tx.Write([]byte{ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3]})
	if err != nil {
		log.WithField("error", err).
			WithField("remote_addr", conn.RemoteAddr()).
			Error("failed to send ip address to client")

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

		return nil, err
	}

	log.WithField("remote_addr", conn.RemoteAddr()).
		WithField("ip", ip).
		Debug("sent ip address to client")

	vPort := NewVPort(macBytes)
	done := make(chan bool)

	go func(ip net.IP, rx, tx net.Conn, mac net.HardwareAddr, vPort *VPort) {
		txMu := sync.Mutex{}

		defer func() {
			log.WithField("remote_addr", rx.RemoteAddr()).
				Info("closing connection")

			txMu.Lock()
			defer txMu.Unlock()

			disconnectFn(vPort)

			tErr, rErr := tx.Close(), rx.Close()
			if tErr != nil {
				log.WithField("error", tErr).
					WithField("remote_addr", rx.RemoteAddr()).
					Error("failed to close tx connection")
			}

			if rErr != nil {
				log.WithField("error", rErr).
					WithField("remote_addr", rx.RemoteAddr()).
					Error("failed to close rx connection")
			}
		}()

		errChan := make(chan error)

		vPort.SetOnReceive(func(data ethernet.Frame) {
			txMu.Lock()
			defer txMu.Unlock()

			_, err := tx.Write(data)
			if err != nil {
				log.WithField("error", err).
					WithField("remote_addr", rx.RemoteAddr()).
					Error("failed to write data to tx connection")

				if err != nil {
					select {
					case errChan <- err:
						// someone already sent an error, so we can just ignore this,
						// this is just to make sending the error non-blocking
						return
					default:
						return
					}
				}
			}
		})
		done <- true

		for {
			select {
			case err := <-errChan:
				log.WithField("error", err).
					WithField("remote_addr", rx.RemoteAddr()).
					Error("failed to write data to tx connection")
				return

			default:
				break
			}

			buff := make([]byte, 4096)
			n, err := rx.Read(buff)
			if err != nil {
				log.WithField("error", err).
					WithField("remote_addr", rx.RemoteAddr()).
					Error("failed to read from rx connection")
				break
			}

			log.WithField("remote_addr", rx.RemoteAddr()).
				WithField("ip", ip.String()).
				Trace("received packet from client")

			go func() {
				err = vPort.Write(buff[:n])
				if err != nil {
					select {
					case errChan <- err:
						// someone already sent an error, so we can just ignore this,
						// this is just to make sending the error non-blocking
						return
					default:
						return
					}
				}
			}()
		}
	}(ip, rx, tx, macBytes, vPort)

	<-done

	return vPort, nil
}
