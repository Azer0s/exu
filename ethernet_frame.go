package exu

import (
	"net"
)

var BroadcastMAC = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

// EthernetFrame represents an ethernet frame. The length of the underlying slice of a
// EthernetFrame should always reflect the ethernet frame length.
type EthernetFrame []byte

// VlanTagging is a type used to indicate whether/how a frame is tagged. The value
// is number of bytes taken by tagging.
type VlanTagging byte

// Const values for different VlanTagging
const (
	TaggingUntagged     VlanTagging = 0
	TaggingTagged       VlanTagging = 4
	TaggingDoubleTagged VlanTagging = 8
)

// Destination returns the destination address field of the frame. The address
// references a slice on the frame.
//
// It is not safe to use this method if f is nil or an invalid ethernet frame.
func (f *EthernetFrame) Destination() net.HardwareAddr {
	return net.HardwareAddr((*f)[:6:6])
}

// Source returns the source address field of the frame. The address references
// a slice on the frame.
//
// It is not safe to use this method if f is nil or an invalid ethernet frame.
func (f *EthernetFrame) Source() net.HardwareAddr {
	return net.HardwareAddr((*f)[6:12:12])
}

// Tagging returns whether/how the frame has 802.1Q tag(s).
func (f *EthernetFrame) Tagging() VlanTagging {
	if (*f)[12] == 0x81 && (*f)[13] == 0x00 {
		return TaggingTagged
	} else if (*f)[12] == 0x88 && (*f)[13] == 0xa8 {
		return TaggingDoubleTagged
	}
	return TaggingUntagged
}

// Tags returns a slice holding the tag part of the frame, if any. Note that
// this includes the Tag Protocol Identifier (TPID), e.g. 0x8100 or 0x88a8.
// Upper layer should use the returned slice for both reading and writing.
func (f *EthernetFrame) Tags() []byte {
	tagging := f.Tagging()
	return (*f)[12 : 12+tagging : 12+tagging]
}

// EtherType returns the ethertype field of the frame.
//
// It is not safe to use this method if f is nil or an invalid ethernet frame.
func (f *EthernetFrame) EtherType() EtherType {
	ethertypePos := 12 + f.Tagging()
	return EtherType{(*f)[ethertypePos], (*f)[ethertypePos+1]}
}

// Payload returns a slice holding the payload part of the frame. Upper layer
// should use the returned slice for both reading and writing purposes.
func (f *EthernetFrame) Payload() []byte {
	return (*f)[12+f.Tagging()+2:]
}

// Resize re-slices (*f) so that len(*f) holds exactly payloadSize bytes of
// payload. If cap(*f) is not large enough, a new slice is made and content
// from old slice is copied to the new one.
//
// If len(*f) is less than 14 bytes, it is assumed to be not tagged.
func (f *EthernetFrame) Resize(payloadSize int) {
	tagging := TaggingUntagged
	if len(*f) > 6+6+2 {
		tagging = f.Tagging()
	}
	f.resize(6 + 6 + int(tagging) + 2 + payloadSize)
}

func (f *EthernetFrame) resize(length int) {
	if cap(*f) < length {
		old := *f
		*f = make(EthernetFrame, length)
		copy(*f, old)
	} else {
		*f = (*f)[:length]
	}
}

// NewEthernetFrame prepares *f to be used, by filling in dst/src address, setting up
// proper tagging and ethertype, and resizing it to proper length.
func NewEthernetFrame(dst net.HardwareAddr, src net.HardwareAddr, tagData TagData, payload EthernetPayload) (f *EthernetFrame, err error) {
	tagging := tagData.GetTagging()
	ethertype := payload.EtherType()
	payloadData, err := payload.MarshalBinary()
	if err != nil {
		return
	}

	payloadSize := len(payloadData)

	f = &EthernetFrame{}

	f.resize(6 + 6 + int(tagging) + 2 + payloadSize)
	copy((*f)[0:6:6], dst)
	copy((*f)[6:12:12], src)
	if tagging == TaggingTagged {
		(*f)[12] = 0x81
		(*f)[13] = 0x00
	} else if tagging == TaggingDoubleTagged {
		(*f)[12] = 0x88
		(*f)[13] = 0xa8
	}
	(*f)[12+tagging] = ethertype[0]
	(*f)[12+tagging+1] = ethertype[1]

	copy((*f)[12+tagging+2:], payloadData)
	err = nil

	return
}
