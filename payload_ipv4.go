package exu

import "net"

type IPv4Protocol uint8

const (
	IPv4ProtocolICMP IPv4Protocol = 1
	IPv4ProtocolTCP  IPv4Protocol = 6
	IPv4ProtocolUDP  IPv4Protocol = 17
)

type IPv4Payload struct {
	// Version is the IP version. It is always 4.
	Version uint8

	// IHL is the IP header length in 32-bit words.
	IHL uint8

	// DSCP is the differentiated services code point
	// (see https://en.wikipedia.org/wiki/Differentiated_services)
	// in most cases, this is set to 0.
	DSCP uint8

	// ECN is the explicit congestion notification (this is used for QoS)
	ECN uint8

	// TotalLength is the total length of the IP packet in bytes.
	TotalLength uint16

	// Identification is a unique identifier for the packet.
	Identification uint16

	// Flags is the flags field of the IP header.
	Flags uint8

	// FragmentOffset is the fragment offset field of the IP header.
	FragmentOffset uint16

	// TTL is the time to live field of the IP header.
	TTL uint8

	// Protocol is the protocol field of the IP header.
	Protocol IPv4Protocol

	// Checksum is the checksum field of the IP header.
	Checksum uint16

	// SourceIP is the source IP address of the packet.
	SourceIP net.IP

	// DestinationIP is the destination IP address of the packet.
	DestinationIP net.IP

	// Options is the options field of the IP header.
	Options []byte

	// Data is the payload of the IP packet.
	Data []byte
}

func (p *IPv4Payload) MarshalBinary() ([]byte, error) {
	res := make([]byte, 20+len(p.Options)+len(p.Data))

	res[0] = (p.Version << 4) | p.IHL
	res[1] = (p.DSCP << 2) | p.ECN
	res[2] = byte(p.TotalLength >> 8)
	res[3] = byte(p.TotalLength)
	res[4] = byte(p.Identification >> 8)
	res[5] = byte(p.Identification)
	res[6] = (p.Flags << 5) | byte(p.FragmentOffset>>8)
	res[7] = byte(p.FragmentOffset)
	res[8] = p.TTL
	res[9] = byte(p.Protocol)
	res[10] = byte(p.Checksum >> 8)
	res[11] = byte(p.Checksum)
	copy(res[12:16], p.SourceIP.To4())
	copy(res[16:20], p.DestinationIP.To4())
	copy(res[20:20+len(p.Options)], p.Options)
	copy(res[20+len(p.Options):], p.Data)

	return res, nil
}

func (p *IPv4Payload) EtherType() EtherType {
	return [2]byte{0x08, 0x00}
}

func (p *IPv4Payload) FromBytes(data []byte) error {
	p.Version = data[0] >> 4
	p.IHL = data[0] & 0x0f
	p.DSCP = data[1] >> 2
	p.ECN = data[1] & 0x03
	p.TotalLength = uint16(data[2])<<8 | uint16(data[3])
	p.Identification = uint16(data[4])<<8 | uint16(data[5])
	p.Flags = data[6] >> 5
	p.FragmentOffset = uint16(data[6]&0x1f)<<8 | uint16(data[7])
	p.TTL = data[8]
	p.Protocol = IPv4Protocol(data[9])
	p.Checksum = uint16(data[10])<<8 | uint16(data[11])
	p.SourceIP = data[12:16]
	p.DestinationIP = data[16:20]
	p.Options = data[20 : 20+p.IHL*4-20]
	p.Data = data[20+p.IHL*4-20:]

	return nil
}
