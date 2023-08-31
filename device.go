package usbrelay

import (
	"fmt"
	"github.com/karalabe/hid"
	"runtime"
	"sync"
)

type Device struct {
	vID        int16
	pID        int16
	mu         *sync.Mutex
	numRelays  Relay
	state      map[Relay]State
	deviceInfo *hid.DeviceInfo
	device     *hid.Device
}

func newDevice(info *hid.DeviceInfo, numRelays int) *Device {
	d := &Device{
		deviceInfo: info,
		mu:         &sync.Mutex{},
		numRelays:  Relay(numRelays),
	}
	d.state = make(map[Relay]State)
	for i := 0; i < numRelays; i++ {
		d.state[Relay(i+1)] = OFF
	}
	return d
}

func (d *Device) Open(readState bool) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.device, err = d.deviceInfo.Open()
	if err != nil {
		return
	}

	if readState {
		_, err = d.getStatus()
	}
	return
}

func (d *Device) Close() error {
	return d.device.Close()
}
func (d *Device) getStatus() (map[Relay]State, error) {
	buf := make([]byte, 9)
	_, err := d.device.GetFeatureReport(buf)
	if err != nil {
		return nil, err
	}

	// Remove HID report ID on Windows, others OSes don't need it.
	if runtime.GOOS == "windows" {
		buf = buf[1:]
	}

	resMap := make(map[Relay]State)
	var (
		state State
		relay Relay
	)
	for i := 0; i < len(d.state); i++ {
		state = State(buf[8] >> i & 0x01)
		relay = Relay(i + 1)
		d.state[relay] = state
		resMap[relay] = state
	}

	return resMap, err
}

func (d *Device) GetStatus() (map[Relay]State, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.getStatus()
}

func (d *Device) onOff(s State, ch Relay) error {
	if (ch < 0 || ch > d.numRelays) && ch != R_ALL {
		return fmt.Errorf("invalid channel number. Must be 1-%d)", R_ALL-1)
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
	return err
}

func (d *Device) Toggle(ch Relay) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	s := d.state[ch]
	switch s {
	case ON:
		s = OFF
		break
	case OFF:
		s = ON
		break
	}

	return d.onOff(s, ch)
}

func (d *Device) On(ch Relay) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.onOff(ON, ch)
}

func (d *Device) Off(ch Relay) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.onOff(OFF, ch)
}

func (d *Device) NumRelays() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return int(d.numRelays)
}
