package exu

import (
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
)

func NewRemoteVport(rxPort int, ip net.IP, onConnect, onDisconnect func(port *VPort)) (*VPort, error) {
	log.WithField("ip", ip.String()).
		WithField("rxPort", rxPort).
		Info("creating new remote vport")

	// create a new UDP socket
	srv, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: rxPort,
	})
	if err != nil {
		panic(err)
	}

	// read initial packet
	// first byte is the magic hello byte (0x42)
	// n + 1 byte is the remote port (2 bytes)

	initial := make([]byte, 9)
	_, remoteAddr, err := srv.ReadFrom(initial)
	if err != nil {
		log.WithField("error", err).
			WithField("remote_addr", remoteAddr.String()).
			Error("failed to read initial packet")
		return nil, err
	}

	if initial[0] != 0x42 {
		log.WithField("remote_addr", remoteAddr.String()).Error("invalid magic byte")
		return nil, errors.New("invalid magic byte")
	}

	port := binary.LittleEndian.Uint16(initial[1:3])
	macBytes := initial[3:]

	log.WithField("remote_addr", remoteAddr.String()).
		WithField("port", port).
		Debug("received initial packet")

	rx := srv
	tx, err := net.Dial("udp", strings.Split(remoteAddr.String(), ":")[0]+":"+strconv.Itoa(int(port)))
	if err != nil {
		log.WithField("error", err).
			WithField("remote_addr", remoteAddr.String()).
			Error("failed to connect to remote port")
		return nil, err
	}

	log.WithField("remote_addr", remoteAddr.String()).
		WithField("port", port).
		Debug("connected to remote tx port")

	// send ip address to client
	ipBytes := ip.To4()
	_, err = tx.Write([]byte{ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3]})
	if err != nil {
		log.WithField("error", err).
			WithField("remote_addr", remoteAddr.String()).
			Error("failed to send ip address to client")

		tErr, rErr := tx.Close(), rx.Close()
		if tErr != nil {
			log.WithField("error", tErr).
				WithField("remote_addr", remoteAddr.String()).
				Error("failed to close tx connection")
		}

		if rErr != nil {
			log.WithField("error", rErr).
				WithField("remote_addr", remoteAddr.String()).
				Error("failed to close rx connection")
		}

		return nil, err
	}

	log.WithField("remote_addr", remoteAddr.String()).
		WithField("ip", ip).
		Debug("sent ip address to client")

	vPort := NewVPort(macBytes, remoteAddr.String())
	done := make(chan bool)

	go func(ip net.IP, rx, tx *net.UDPConn, mac net.HardwareAddr, vPort *VPort, errChan chan error) {
		defer func() {
			log.WithField("remote_addr", remoteAddr.String()).
				Info("closing connection")

			onDisconnect(vPort)

			tErr, rErr := tx.Close(), rx.Close()
			if tErr != nil {
				log.WithField("error", tErr).
					WithField("remote_addr", remoteAddr.String()).
					Error("failed to close tx connection")
			}

			if rErr != nil {
				log.WithField("error", rErr).
					WithField("remote_addr", remoteAddr.String()).
					Error("failed to close rx connection")
			}
		}()

		vPort.SetOnReceive(func(data *EthernetFrame) {
			_, err := tx.Write(*data)
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
		})

		done <- true

		for {
			if vPort.connectedTo == nil {
				onConnect(vPort)
				continue
			}

			select {
			case err = <-errChan:
				log.WithField("error", err).
					WithField("remote_addr", remoteAddr.String()).
					Error("failed to write data to tx connection")
				return

			default:
				break
			}

			buff := make([]byte, 4096)

			_ = rx.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, _, err := rx.ReadFromUDP(buff)

			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) {
					// we timed out, lets try again next loop
					continue
				}

				log.WithField("error", err).
					WithField("remote_addr", remoteAddr.String()).
					Error("failed to read from rx connection")
				break
			}

			log.WithField("remote_addr", remoteAddr.String()).
				WithField("ip", ip.String()).
				Trace("received packet from client")

			go func(buff []byte) {
				frame := EthernetFrame(buff)
				err = vPort.Write(&frame)
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
			}(buff[:n])
		}
	}(ip, rx, tx.(*net.UDPConn), macBytes, vPort, make(chan error, 1))

	<-done

	return vPort, nil
}
