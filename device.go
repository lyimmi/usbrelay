package usbrelay

import (
	"fmt"
	"github.com/karalabe/hid"
	"runtime"
	"sync"
)

type SerialNumber string

// NewSerialNumber returns an array of characters with the length of 5
func NewSerialNumber(s string) SerialNumber {
	var sn []byte
	copy(sn, s[:5])
	return SerialNumber(sn)
}

// Device represents a USB HID relay device
type Device struct {
	vID          int16
	pID          int16
	mu           *sync.Mutex
	numRelays    int
	state        map[RelayNumber]State
	deviceInfo   *hid.DeviceInfo
	device       *hid.Device
	serialNumber SerialNumber
	isOpen       bool
}

func newDevice(info *hid.DeviceInfo, relayCount int) *Device {
	d := &Device{
		deviceInfo: info,
		mu:         &sync.Mutex{},
		numRelays:  relayCount,
	}
	d.state = make(map[RelayNumber]State, relayCount)
	for i := 0; i < relayCount; i++ {
		d.state[RelayNumber(i+1)] = OFF
	}
	return d
}

// Open connects to the USB device
func (d *Device) Open(readState bool) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.isOpen = true

	d.device, err = d.deviceInfo.Open()
	if err != nil {
		return
	}

	if readState {
		_, err = d.readStates()
	}
	return
}

// Close closes the USB connection to the device
func (d *Device) Close() error {
	d.isOpen = false
	return d.device.Close()
}

// readStates reads the state of all the relays on the device
func (d *Device) readStates() (map[RelayNumber]State, error) {
	buf := make([]byte, 9)
	_, err := d.device.GetFeatureReport(buf)
	if err != nil {
		return nil, err
	}

	// Remove HID report ID on Windows, others OSes don't need it.
	if runtime.GOOS == "windows" {
		buf = buf[1:]
	}

	resMap := make(map[RelayNumber]State)
	var (
		state State
		relay RelayNumber
	)
	for i := 0; i < len(d.state); i++ {
		state = State(buf[8] >> i & 0x01)
		relay = RelayNumber(i + 1)
		d.state[relay] = state
		resMap[relay] = state
	}

	return resMap, err
}

// changeState changes the state of one or all of the relays on the device
func (d *Device) changeState(s State, ch RelayNumber) error {
	if (ch < 0 || int(ch) > d.numRelays) && ch != R_ALL {
		return fmt.Errorf("%w must be 1-%d", ErrInvalidRelayNumber, R_ALL-1)
	}

	if s == d.state[ch] {
		return nil
	}

	cmdBuffer := make([]byte, 9)
	cmdBuffer[0] = 0x0
	if ch == R_ALL {
		if s == ON {
			cmdBuffer[1] = 0xFE
		} else {
			cmdBuffer[1] = 0xFC
		}
	} else {
		if s == ON {
			cmdBuffer[1] = 0xFF
		} else {
			cmdBuffer[1] = 0xFD
		}
		cmdBuffer[2] = byte(ch)
	}
	d.state[ch] = s

	_, err := d.device.SendFeatureReport(cmdBuffer)

	states, err := d.readStates()
	if err != nil {
		return err
	}

	if ch == R_ALL {
		for i := 0; i < d.numRelays; i++ {
			if states[RelayNumber(i)] != d.state[RelayNumber(i)] {
				return fmt.Errorf("%w relay: %d state: %d", ErrRelayStateNotSet, i, s)
			}
		}
	} else if states[ch] != d.state[ch] {
		return fmt.Errorf("%w relay: %d state: %d", ErrRelayStateNotSet, ch, s)
	}

	return err
}

// States returns the state of all the relays on the device
func (d *Device) States() (map[RelayNumber]State, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return nil, ErrDeviceNotConnected
	}

	return d.readStates()
}

// Toggle the state of one or all of the relays on the device
func (d *Device) Toggle(ch RelayNumber) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return ErrDeviceNotConnected
	}

	s := d.state[ch]
	switch s {
	case ON:
		s = OFF
		break
	case OFF:
		s = ON
		break
	}

	return d.changeState(s, ch)
}

// On sets one or all of the relays state on the device to ON
func (d *Device) On(ch RelayNumber) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return ErrDeviceNotConnected
	}

	return d.changeState(ON, ch)
}

// Off sets one or all of the relays state on the device to OFF
func (d *Device) Off(ch RelayNumber) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return ErrDeviceNotConnected
	}

	return d.changeState(OFF, ch)
}

// RelayCount returns the number of relays found on the device
func (d *Device) RelayCount() (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return 0, ErrDeviceNotConnected
	}

	return d.numRelays, nil
}

// SetSerialNumber writes the serial number on the device
func (d *Device) SetSerialNumber(sn SerialNumber) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return ErrDeviceNotConnected
	}

	if len(sn) > 5 {
		err = fmt.Errorf("%w %s is large than 5 bytes", ErrInvalidSerialNumberLen, sn)
		return
	}
	cmd := make([]byte, 9)
	cmd[0] = 0x00
	cmd[1] = 0xFA
	copy(cmd[2:], sn)
	_, err = d.device.SendFeatureReport(cmd)
	if err == nil {
		d.serialNumber = sn
	}
	return
}

// GetSerialNumber reads the serial number from the device
func (d *Device) GetSerialNumber() (sn SerialNumber, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.serialNumber != "" {
		return d.serialNumber, nil
	}

	if !d.isOpen {
		return "", ErrDeviceNotConnected
	}

	cmd := make([]byte, 9)
	_, err = d.device.GetFeatureReport(cmd)
	if err != nil {
		return
	}
	if runtime.GOOS == "windows" {
		cmd = cmd[1:]
	}
	sn = SerialNumber(cmd[:5])
	d.serialNumber = sn
	return
}
