package exu

type CapabilityStatus int

const (
	// CapabilityStatusPass means the request was not handled by the capability
	// and should be passed to the next capability
	CapabilityStatusPass CapabilityStatus = iota
	// CapabilityStatusFail means the request was dropped by the capability
	CapabilityStatusFail
	// CapabilityStatusDone means the request was handled by the capability
	CapabilityStatusDone
)

type Capability interface {
	// HandleRequest handles a request and returns a status code
	HandleRequest(port *VPort, data *EthernetFrame) CapabilityStatus

	// Match returns true if the capability can handle the data
	Match(port *VPort, data *EthernetFrame) bool
}
