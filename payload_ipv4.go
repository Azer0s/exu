package exu

import (
	"encoding/binary"
	"errors"
	"net"
)

// IPv4Packet represents the structure of an IPv4 packet
type IPv4Packet struct {
	Header  IPv4Header // IPv4 header
	Payload []byte     // Payload data
}

// MarshalBinary converts the IPv4Packet struct to its binary representation
func (p *IPv4Packet) MarshalBinary() ([]byte, error) {
	headerBytes, err := p.Header.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Combine header and payload
	packetBytes := append(headerBytes, p.Payload...)

	return packetBytes, nil
}

func (p *IPv4Packet) EtherType() EtherType {
	return EtherTypeIPv4
}

// UnmarshalBinary converts the binary representation to an IPv4Packet struct
func (p *IPv4Packet) UnmarshalBinary(data []byte) error {
	// Minimum size for a valid IPv4 packet is the size of the header (20 bytes)
	if len(data) < 20 {
		return errors.New("ipv4 packet must be at least 20 bytes")
	}

	// Unmarshal header
	if err := p.Header.UnmarshalBinary(data[:20]); err != nil {
		return err
	}

	// Set payload
	p.Payload = data[20:]

	return nil
}

type IPv4Protocol uint8

const (
	IPv4ProtocolICMP IPv4Protocol = 1
	IPv4ProtocolTCP  IPv4Protocol = 6
	IPv4ProtocolUDP  IPv4Protocol = 17
)

// IPv4Header represents the structure of an IPv4 header
type IPv4Header struct {
	Version        uint8        // 4-bit IP version
	IHL            uint8        // 4-bit Internet Header Length (IHL)
	TOS            uint8        // 8-bit Type of Service (TOS)
	TotalLength    uint16       // 16-bit Total Length
	ID             uint16       // 16-bit Identification
	FlagsFragment  uint16       // 3-bit Flags and 13-bit Fragment Offset
	TTL            uint8        // 8-bit Time to Live (TTL)
	Protocol       IPv4Protocol // 8-bit Protocol
	HeaderChecksum uint16       // 16-bit Header Checksum
	SourceIP       net.IP       // 32-bit Source IP Address
	DestinationIP  net.IP       // 32-bit Destination IP Address
}

// MarshalBinary converts the IPv4Header struct to its binary representation
func (header *IPv4Header) MarshalBinary() ([]byte, error) {
	b := make([]byte, 20) // IPv4 header length is 20 bytes

	// Version and IHL (Internet Header Length)
	b[0] = (header.Version << 4) | (header.IHL & 0x0F)

	// Type of Service (TOS)
	b[1] = header.TOS

	// Total Length
	binary.BigEndian.PutUint16(b[2:4], header.TotalLength)

	// Identification
	binary.BigEndian.PutUint16(b[4:6], header.ID)

	// Flags and Fragment Offset
	binary.BigEndian.PutUint16(b[6:8], header.FlagsFragment)

	// Time to Live (TTL)
	b[8] = header.TTL

	// Protocol
	b[9] = byte(header.Protocol)

	// Header Checksum
	binary.BigEndian.PutUint16(b[10:12], header.HeaderChecksum)

	// Source IP Address
	copy(b[12:16], header.SourceIP.To4())

	// Destination IP Address
	copy(b[16:20], header.DestinationIP.To4())

	return b, nil
}

// UnmarshalBinary converts the binary representation to an IPv4Header struct
func (header *IPv4Header) UnmarshalBinary(data []byte) error {
	if len(data) < 20 {
		return errors.New("ipv4 header must be at least 20 bytes")
	}

	// Version and IHL (Internet Header Length)
	header.Version = data[0] >> 4
	header.IHL = data[0] & 0x0F

	// Type of Service (TOS)
	header.TOS = data[1]

	// Total Length
	header.TotalLength = binary.BigEndian.Uint16(data[2:4])

	// Identification
	header.ID = binary.BigEndian.Uint16(data[4:6])

	// Flags and Fragment Offset
	header.FlagsFragment = binary.BigEndian.Uint16(data[6:8])

	// Time to Live (TTL)
	header.TTL = data[8]

	// Protocol
	header.Protocol = IPv4Protocol(data[9])

	// Header Checksum
	header.HeaderChecksum = binary.BigEndian.Uint16(data[10:12])

	// Source IP Address
	header.SourceIP = net.IP(data[12:16])

	// Destination IP Address
	header.DestinationIP = net.IP(data[16:20])

	return nil
}

// CalculateHeaderChecksum calculates the header checksum for an IPv4 header
func (header *IPv4Header) CalculateChecksum() uint16 {
	// Save the current checksum value
	oldChecksum := header.HeaderChecksum

	// Set the checksum field to 0
	header.HeaderChecksum = 0

	// Convert the header to a byte slice
	headerBytes, _ := header.MarshalBinary()

	// Initialize the checksum
	checksum := uint32(0)

	// Iterate over 16-bit words in the header
	for i := 0; i < len(headerBytes); i += 2 {
		checksum += uint32(binary.BigEndian.Uint16(headerBytes[i : i+2]))
	}

	// Fold the carry into the checksum
	checksum = (checksum >> 16) + (checksum & 0xffff)
	checksum += checksum >> 16

	// Take the one's complement
	checksum = ^checksum

	// Set the header checksum back to its original value
	header.HeaderChecksum = oldChecksum

	return uint16(checksum)
}
