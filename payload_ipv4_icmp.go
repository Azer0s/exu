package exu

type ICMPType uint8

const (
	ICMPTypeEchoReply ICMPType = 0
	ICMPTypeEcho      ICMPType = 8
)

type ICMPPayload struct {
	Type     ICMPType
	Code     uint8
	Checksum uint16
	Data     []byte
}

func (i *ICMPPayload) MarshalBinary() ([]byte, error) {
	res := make([]byte, 4+len(i.Data))
	res[0] = byte(i.Type)
	res[1] = i.Code
	res[2] = byte(i.Checksum >> 8)
	res[3] = byte(i.Checksum)
	copy(res[4:], i.Data)
	return res, nil
}

func (i *ICMPPayload) UnmarshalBinary(data []byte) error {
	i.Type = ICMPType(data[0])
	i.Code = data[1]
	i.Checksum = uint16(data[2])<<8 | uint16(data[3])
	i.Data = data[4:]
	return nil
}

func (i *ICMPPayload) CalculateChecksum() uint16 {
	data := i.Data
	if len(data)%2 != 0 {
		data = append(data, 0)
	}

	var sum uint32
	for i := 0; i < len(data); i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}

	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}

	return uint16(^sum)
}
