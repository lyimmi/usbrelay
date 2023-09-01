package usbrelay

import (
	"fmt"
	"github.com/karalabe/hid"
	"math"
	"runtime"
	"sync"
	"unicode"
)

// SerialNumber is the device's unique identifier
type SerialNumber string

// NewSerialNumber returns an array of characters with the length of 5
func NewSerialNumber(s string) SerialNumber {
	snLen := int(math.Min(5, float64(len(s))))
	newSerial := make([]byte, 5)
	copy(newSerial, s[:snLen])
	return SerialNumber(newSerial)
}

// Device represents a USB HID relay device
type Device struct {
	mu           *sync.Mutex
	connected    bool
	relayCount   RelayNumber
	device       *hid.Device
	deviceInfo   *hid.DeviceInfo
	serialNumber SerialNumber
	state        map[RelayNumber]State
}

func newDevice(info *hid.DeviceInfo, relayCount int) *Device {
	d := &Device{
		deviceInfo: info,
		mu:         &sync.Mutex{},
		relayCount: RelayNumber(relayCount),
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

	d.device, err = d.deviceInfo.Open()
	if err != nil {
		return
	}

	d.connected = true

	if readState {
		_, err = d.readStates()
	}
	return
}

// Close closes the USB connection to the device
func (d *Device) Close() error {
	d.connected = false
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
	for i := 0; i < int(d.relayCount); i++ {
		state = State(buf[8] >> i & 0x01)
		relay = RelayNumber(i + 1)
		d.state[relay] = state
		resMap[relay] = state
	}

	return resMap, err
}

// changeState changes the state of one or all of the relays on the device
func (d *Device) changeState(s State, ch RelayNumber) error {
	if (ch < 0 || ch > d.relayCount) && ch != R_ALL {
		return fmt.Errorf("%w must be 1-%d", ErrInvalidRelayNumber, d.relayCount)
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
		for i := 0; i < int(d.relayCount); i++ {
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

	if !d.connected {
		return nil, ErrDeviceNotConnected
	}

	return d.readStates()
}

// Toggle the state of one or all of the relays on the device
func (d *Device) Toggle(ch RelayNumber) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrDeviceNotConnected
	}

	if ch == R_ALL {
		for i := 1; i <= int(d.relayCount); i++ {
			r := RelayNumber(i)
			s := switchState(d.state[r])
			if err := d.changeState(s, r); err != nil {
				return err
			}
		}
		return nil
	}

	s := switchState(d.state[ch])
	return d.changeState(s, ch)
}

// On sets one or all of the relays state on the device to ON
func (d *Device) On(ch RelayNumber) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrDeviceNotConnected
	}

	return d.changeState(ON, ch)
}

// Off sets one or all of the relays state on the device to OFF
func (d *Device) Off(ch RelayNumber) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrDeviceNotConnected
	}

	return d.changeState(OFF, ch)
}

// RelayCount returns the number of relays found on the device
func (d *Device) RelayCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return int(d.relayCount)
}

// SetSerialNumber writes the serial number on the device
func (d *Device) SetSerialNumber(sn SerialNumber) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrDeviceNotConnected
	}

	if len(sn) > 5 {
		return fmt.Errorf("%w %s is longer than 5 characters", ErrInvalidSerialNumber, sn)
	}

	if !isASCII(string(sn)) {
		return fmt.Errorf("%w %s is longer than 5 characters", ErrInvalidSerialNumber, sn)
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

	if !d.connected {
		return "", ErrDeviceNotConnected
	}

	buf := make([]byte, 9)
	_, err = d.device.GetFeatureReport(buf)
	if err != nil {
		return
	}
	buf = buf[1:]
	sn = SerialNumber(buf[:5])
	d.serialNumber = sn
	return
}

// Info returns the basic information about the device as a tuple
func (d *Device) Info() (serialNumber SerialNumber, vendorID int, productID int, relayCount int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.serialNumber, int(d.deviceInfo.VendorID), int(d.deviceInfo.ProductID), int(d.relayCount)
}

// String returns the basic information about the device as a string
func (d *Device) String() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	return fmt.Sprintf("%s:%d:%d:%d", d.serialNumber, d.relayCount, d.deviceInfo.VendorID, d.deviceInfo.ProductID)
}

func switchState(s State) State {
	switch s {
	case ON:
		return OFF
	case OFF:
		return ON
	}
	return OFF
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
