package exu

// EtherType is a type used represent the EtherType of an ethernet frame.
// Defined as a 2-byte array, variables of this type are intended to be used as
// immutable values.
type EtherType [2]byte

func (e EtherType) Equal(other EtherType) bool {
	return e[0] == other[0] && e[1] == other[1]
}

// Common EtherType values
var (
	EtherTypeIPv4                = EtherType{0x08, 0x00}
	EtherTypeARP                 = EtherType{0x08, 0x06}
	EtherTypeWakeOnLAN           = EtherType{0x08, 0x42}
	EtherTypeRARP                = EtherType{0x80, 0x35}
	EtherTypeAppleTalk           = EtherType{0x80, 0x9B}
	EtherTypeAARP                = EtherType{0x80, 0xF3}
	EtherTypeQNXQnet             = EtherType{0x82, 0x04}
	EtherTypeIPv6                = EtherType{0x86, 0xDD}
	EtherTypeEthernetFlowControl = EtherType{0x88, 0x08}
	EtherTypeCobraNet            = EtherType{0x88, 0x19}
	EtherTypePPPoEDiscovery      = EtherType{0x88, 0x63}
	EtherTypePPPoESession        = EtherType{0x88, 0x64}
	EtherTypeJumboFrames         = EtherType{0x88, 0x70}
	EtherTypePROFINET            = EtherType{0x88, 0x92}
	EtherTypeHyperSCSI           = EtherType{0x88, 0x9A}
	EtherTypeAoE                 = EtherType{0x88, 0xA2}
	EtherTypeEtherCAT            = EtherType{0x88, 0xA4}
	EtherTypeEthernetPowerlink   = EtherType{0x88, 0xAB}
	EtherTypeLLDP                = EtherType{0x88, 0xCC}
	EtherTypeSERCOS3             = EtherType{0x88, 0xCD}
	EtherTypeWSMP                = EtherType{0x88, 0xDC}
	EtherTypeHomePlugAVMME       = EtherType{0x88, 0xE1}
	EtherTypeMRP                 = EtherType{0x88, 0xE3}
	EtherTypeFCoE                = EtherType{0x89, 0x06}
	EtherTypeFCoEInit            = EtherType{0x89, 0x14}
	EtherTypeRoCE                = EtherType{0x89, 0x15}
	EtherTypeCTP                 = EtherType{0x90, 0x00}
	EtherTypeVeritasLLT          = EtherType{0xCA, 0xFE}
)
