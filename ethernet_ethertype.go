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
	EtherTypeTRILL               = EtherType{0x22, 0xF3}
	EtherTypeDECnetPhase4        = EtherType{0x60, 0x03}
	EtherTypeRARP                = EtherType{0x80, 0x35}
	EtherTypeAppleTalk           = EtherType{0x80, 0x9B}
	EtherTypeAARP                = EtherType{0x80, 0xF3}
	EtherTypeIPX1                = EtherType{0x81, 0x37}
	EtherTypeIPX2                = EtherType{0x81, 0x38}
	EtherTypeQNXQnet             = EtherType{0x82, 0x04}
	EtherTypeIPv6                = EtherType{0x86, 0xDD}
	EtherTypeEthernetFlowControl = EtherType{0x88, 0x08}
	EtherTypeIEEE802_3           = EtherType{0x88, 0x09}
	EtherTypeCobraNet            = EtherType{0x88, 0x19}
	EtherTypeMPLSUnicast         = EtherType{0x88, 0x47}
	EtherTypeMPLSMulticast       = EtherType{0x88, 0x48}
	EtherTypePPPoEDiscovery      = EtherType{0x88, 0x63}
	EtherTypePPPoESession        = EtherType{0x88, 0x64}
	EtherTypeJumboFrames         = EtherType{0x88, 0x70}
	EtherTypeHomePlug1_0MME      = EtherType{0x88, 0x7B}
	EtherTypeIEEE802_1X          = EtherType{0x88, 0x8E}
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
	EtherTypeIEEE802_1AE         = EtherType{0x88, 0xE5}
	EtherTypeIEEE1588            = EtherType{0x88, 0xF7}
	EtherTypeIEEE802_1ag         = EtherType{0x89, 0x02}
	EtherTypeFCoE                = EtherType{0x89, 0x06}
	EtherTypeFCoEInit            = EtherType{0x89, 0x14}
	EtherTypeRoCE                = EtherType{0x89, 0x15}
	EtherTypeCTP                 = EtherType{0x90, 0x00}
	EtherTypeVeritasLLT          = EtherType{0xCA, 0xFE}
)
