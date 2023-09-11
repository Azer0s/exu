package vswitch

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
)

type MACAddress [6]byte

type VPort struct {
	id     string
	mac    string
	rx, tx *net.Conn
	txMu   sync.Mutex
}

func init() {
	MacBytesToStr([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func (v *VPort) Write(data []byte, src string) (int, error) {
	v.txMu.Lock()
	defer v.txMu.Unlock()

	log.WithField("dst", v.mac).
		WithField("src", src).
		Debug("writing packet to vPort")

	return (*v.tx).Write(data)
}

type Handler struct {
	connections   map[string]*VPort
	connectionsMu sync.RWMutex
}

func (v *Handler) registerMacIfNotExists(mac string, rx, tx *net.Conn) *VPort {
	v.connectionsMu.Lock()
	defer v.connectionsMu.Unlock()

	if _, ok := v.connections[mac]; !ok {
		port := new(VPort)
		port.id = uuid.New().String()
		port.mac = mac
		port.rx = rx
		port.tx = tx
		port.txMu = sync.Mutex{}

		log.WithField("mac", mac).
			WithField("id", port.id).
			Debug("registering mac address on vPort")

		v.connections[mac] = port
	}

	return v.connections[mac]
}

func (v *Handler) getPortByMac(mac string) *VPort {
	v.connectionsMu.RLock()
	defer v.connectionsMu.RUnlock()

	if port, ok := v.connections[mac]; ok {
		return port
	}

	return nil
}

func (v *Handler) flood(src *VPort, data []byte) {
	v.connectionsMu.RLock()
	defer v.connectionsMu.RUnlock()

	for _, port := range v.connections {
		if port != src {
			_, err := (*port).Write(data, src.mac)
			if err != nil {
				panic(err)
			}
		}
	}
}

func MacBytesToStr(mac []byte) string {
	return net.HardwareAddr(mac).String()
}

func (v *Handler) Handle(rx, tx net.Conn) {
	buff := make([]byte, 4096)

	for {
		n, err := rx.Read(buff)
		if err != nil {
			err := v.disconnect(&rx, &tx)
			if err != nil {
				log.WithField("error", err).Error("failed to disconnect")
			}
			return
		}

		// register mac address to vPort
		dst := buff[:6]
		dstStr := net.HardwareAddr(dst).String()
		dstPort := v.getPortByMac(dstStr)

		src := buff[6:12]
		srcStr := net.HardwareAddr(src).String()
		srcPort := v.registerMacIfNotExists(srcStr, &rx, &tx)

		log.WithField("src", srcStr).
			WithField("dst", dstStr).
			Trace("received packet")

		data := buff[12:n]
		if dstPort == nil {
			// flood to all connections
			log.WithField("src", srcStr).
				WithField("dst", dstStr).
				Debug("no destination port found, flooding packet")

			go func(srcPort *VPort, data []byte) {
				v.flood(srcPort, data)
			}(srcPort, data)
		} else {
			// send to dstConn
			go func(srcPort, dstPort *VPort, data []byte) {
				_, err := (*dstPort).Write(data, srcPort.mac)
				if err != nil {
					panic(err)
				}
			}(srcPort, dstPort, data)
		}
	}
}

func (v *Handler) disconnect(rx, tx *net.Conn) error {
	log.WithField("remote_addr", (*rx).RemoteAddr()).Info("closing connection")

	v.connectionsMu.Lock()
	defer v.connectionsMu.Unlock()

	for mac, port := range v.connections {
		if port.rx == rx && port.tx == tx {
			delete(v.connections, mac)
			break
		}
	}

	err := (*tx).Close()
	if err != nil {
		return err
	}

	err = (*rx).Close()
	if err != nil {
		return err
	}

	return nil
}
