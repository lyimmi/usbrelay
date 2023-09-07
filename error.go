package usbrelay

import "errors"

// Error types
var (
	ErrNoDeviceFound         = errors.New("no device found")
	ErrDeviceInfoNotFound    = errors.New("cannot connect to device, device information not found")
	ErrDeviceNotConnected    = errors.New("device is not connected, call Open()")
	ErrInvalidNumberOfRelays = errors.New("invalid number of relays found")
	ErrRelayStateNotSet      = errors.New("relay state could not be set")
	ErrInvalidSerialNumber   = errors.New("invalid serial number")
	ErrInvalidRelayNumber    = errors.New("invalid relay number")
)
