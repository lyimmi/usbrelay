package usbrelay

const (
	DeviceVendorID  uint16 = 0x16c0
	DeviceProductID uint16 = 0x05DF
	UdevVendorID    string = "16c0"
	UdevProductID   string = "05df"
	relayNamePrefix        = "USBRelay"
)

// State represents a relay's state ON or OFF
type State int

// Available relay State(s)
const (
	OFF State = iota
	ON
)

// RelayNumber is the relay's identifier on the device
//
// Valid identifier can be found by calling Device.RelayCount
type RelayNumber int

// Available RelayNumber(s)
const (
	R1 RelayNumber = iota + 1
	R2
	R3
	R4
	R5
	R6
	R7
	R8
	R9
	R10
	R11
	R12
	R_ALL = 1000
)
