package exu

type EthernetPayload interface {
	EtherType() EtherType
	MarshalBinary() ([]byte, error)
}
