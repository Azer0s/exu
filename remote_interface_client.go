package exu

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type RemoteInterfaceClient struct {
	ifce *water.Interface
	host net.TCPAddr
}

func NewRemoteInterfaceClient(host net.TCPAddr) RemoteInterfaceClient {
	config := water.Config{
		DeviceType: water.TAP,
	}

	// get tap interfaces
	cmd := exec.Command("ip", "tuntap")

	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	highestTap := 0
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "tap") {
			config.Name = strings.Split(line, ":")[0]
			tapNum, err := strconv.Atoi(config.Name[3:])
			if err != nil {
				continue
			}

			if tapNum >= highestTap {
				highestTap = tapNum + 1
			}
		}
	}

	config.Name = "tap" + strconv.Itoa(highestTap)

	ifce, err := water.New(config)
	if err != nil {
		log.Fatal(err)
	}

	return RemoteInterfaceClient{
		ifce: ifce,
		host: host,
	}
}

func (c *RemoteInterfaceClient) Run() {
	log.Info("starting client")

	portChan := make(chan uint16)
	ipChan := make(chan uint32)

	go func() {
		// run tcp server on random port
		tcp, err := net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}

		portChan <- uint16(tcp.Addr().(*net.TCPAddr).Port)

		// accept connection
		conn, err := tcp.Accept()
		if err != nil {
			panic(err)
		}

		log.WithField("remote_addr", conn.RemoteAddr().String()).Info("remote connected to rx port")

		// read ip address assigned to us
		ipBytes := make([]byte, 4)
		_, err = conn.Read(ipBytes)
		if err != nil {
			panic(err)
		}
		ip := binary.LittleEndian.Uint32(ipBytes)
		ipChan <- ip

		// forward to tap interface
		buff := make([]byte, 4096)
		for {
			n, err := conn.Read(buff)
			if err != nil {
				panic(err)
			}

			log.Trace("packet received from remote")

			_, err = c.ifce.Write(buff[:n])
			if err != nil {
				panic(err)
			}

			log.Trace("packet sent to tap interface")
		}
	}()

	port := <-portChan
	portBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBytes, port)

	// connect to server
	conn, err := net.Dial("tcp", c.host.String())
	if err != nil {
		panic(err)
	}

	log.WithField("remote_addr", conn.RemoteAddr().String()).Info("connected to server")

	// get mac address of tap interface
	cmd := exec.Command("ip", "address", "show", c.ifce.Name())
	out, err := cmd.Output()
	if err != nil {
		log.WithField("error", err).Fatal("failed to get mac address")
	}

	// mac should be after "link/ether"
	mac, err := net.ParseMAC(strings.Split(string(out), "link/ether ")[1][:17])
	if err != nil {
		log.WithField("error", err).Fatal("failed to parse mac address")
	}

	log.WithField("mac", mac.String()).Debug("got mac address of tap interface")

	// send initial packet
	// first byte is the magic hello byte (0x42)
	// second and third byte is the remote port (2 bytes)
	// next 6 bytes is the mac address
	_, err = conn.Write([]byte{0x42, portBytes[0], portBytes[1], mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]})
	if err != nil {
		panic(err)
	}

	log.WithField("remote_addr", conn.RemoteAddr().String()).Debug("sent initial packet")

	ipStr := ""
	select {
	case ip := <-ipChan:
		ipBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(ipBytes, ip)
		ipStr = net.IP(ipBytes).String() + "/24"

	case <-time.After(5 * time.Second):
		log.Fatal("timed out waiting for ip address")
	}

	log.WithField("ip", ipStr).Debug("received ip address")

	err = exec.Command("ip", "addr", "add", ipStr, "dev", c.ifce.Name()).Run()
	if err != nil {
		panic(err)
	}

	log.WithField("ip", ipStr).Debug("added ip address to tap interface")

	err = exec.Command("ip", "link", "set", "dev", c.ifce.Name(), "up").Run()
	if err != nil {
		panic(err)
	}

	log.WithField("ifce", c.ifce.Name()).Debug("tap interface set to up")

	log.WithField("ip", ipStr).Info("client started")

	var frame ethernet.Frame
	for {
		frame.Resize(1500)
		n, err := c.ifce.Read(frame)
		if err != nil {
			log.Fatal(err)
		}
		frame = frame[:n]

		log.Trace("packet received from tap interface")

		_, err = conn.Write(frame)
		if err != nil {
			panic(err)
		}

		log.Trace("packet sent to server")
	}
}
