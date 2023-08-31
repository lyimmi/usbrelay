package usbrelay

const (
	cfgVendorID     uint16 = 0x16c0
	cfgDeviceID     uint16 = 0x05DF
	relayNamePrefix        = "USBRelay"
)

type State int

// Relay states
const (
	OFF State = iota
	ON
)

type Relay int

// Relay numbers
const (
	R1 Relay = iota + 1
	R2
	R3
	R4
	R5
	R6
	R7
	R8
	R_ALL
)
