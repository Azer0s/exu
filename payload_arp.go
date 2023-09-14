package exu

import "net"

var ArpMacBroadcast = net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

type ArpHardwareType uint16

const (
	ArpHardwareTypeEthernet ArpHardwareType = 1
	AprHardwareTypeIeee802  ArpHardwareType = 6
	ArpHardwareTypeArcnet   ArpHardwareType = 7
	ArpHardwareTypeFrameRel ArpHardwareType = 15
	ArpHardwareTypeAtm      ArpHardwareType = 16
	ArpHardwareTypeHDLC     ArpHardwareType = 17
	ArpHardwareTypeFibre    ArpHardwareType = 18
	ArpHardwareTypeAtm2     ArpHardwareType = 19
	ArpHardwareTypeSerial   ArpHardwareType = 20
)

type ArpProtocolType uint16

const (
	ArpProtocolTypeIPv4 ArpProtocolType = 0x0800
	ArpProtocolTypeIPv6 ArpProtocolType = 0x86DD
)

type ArpOpcode uint16

const (
	ArpOpcodeRequest ArpOpcode = 1
	ArpOpcodeReply   ArpOpcode = 2
)

type ArpPayload struct {
	HardwareType ArpHardwareType
	ProtocolType ArpProtocolType
	Opcode       ArpOpcode
	SenderMac    net.HardwareAddr
	SenderIP     net.IP
	TargetMac    net.HardwareAddr
	TargetIP     net.IP
}

func (a *ArpPayload) MarshalBinary() ([]byte, error) {
	senderIpBytes := a.SenderIP.To4()[0:4]
	targetIpBytes := a.TargetIP.To4()[0:4]

	return []byte{
		byte(a.HardwareType >> 8),
		byte(a.HardwareType),
		byte(a.ProtocolType >> 8),
		byte(a.ProtocolType),
		0x06, 0x04, // for now, we only support ethernet and ipv4 anyway
		byte(a.Opcode >> 8),
		byte(a.Opcode),
		a.SenderMac[0], a.SenderMac[1], a.SenderMac[2], a.SenderMac[3], a.SenderMac[4], a.SenderMac[5],
		senderIpBytes[0], senderIpBytes[1], senderIpBytes[2], senderIpBytes[3],
		a.TargetMac[0], a.TargetMac[1], a.TargetMac[2], a.TargetMac[3], a.TargetMac[4], a.TargetMac[5],
		targetIpBytes[0], targetIpBytes[1], targetIpBytes[2], targetIpBytes[3],
	}, nil
}

func (a *ArpPayload) EtherType() EtherType {
	return EtherTypeARP
}

func NewArpPayload(hardwareType ArpHardwareType, protocolType ArpProtocolType, opcode ArpOpcode, senderMac net.HardwareAddr, senderIP net.IP, targetMac net.HardwareAddr, targetIP net.IP) *ArpPayload {
	return &ArpPayload{
		HardwareType: hardwareType,
		ProtocolType: protocolType,
		Opcode:       opcode,
		SenderMac:    senderMac,
		SenderIP:     senderIP,
		TargetMac:    targetMac,
		TargetIP:     targetIP,
	}
}
