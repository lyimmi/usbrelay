package usbrelay

import "errors"

// Error types
var (
	ErrNoDeviceFound          = errors.New("no device found")
	ErrDeviceNotConnected     = errors.New("device is not connected, call Open()")
	ErrInvalidNumberOfRelays  = errors.New("invalid number of relays found")
	ErrRelayStateNotSet       = errors.New("relay state could not be set")
	ErrInvalidSerialNumberLen = errors.New("invalid serial number length")
	ErrInvalidRelayNumber     = errors.New("invalid relay number")
)
