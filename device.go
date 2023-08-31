package usbrelay

import (
	"fmt"
	"github.com/google/gousb"
	"sync"
)

func NewDevice(vID int16, pID int16, numRelays uint8) *Device {
	d := &Device{
		vID:       vID,
		pID:       pID,
		mu:        &sync.Mutex{},
		numRelays: numRelays,
	}
	d.state = make(map[int]bool)
	for i := 0; i < int(numRelays); i++ {
		d.state[i] = false
	}
	return d
}

type Device struct {
	vID       int16
	pID       int16
	mu        *sync.Mutex
	numRelays uint8
	state     map[int]bool
}

func (d *Device) onOff(state bool, i int) error {
	if i < 0 || i > int(d.numRelays) {
		return fmt.Errorf("invalid usbDevice number. Must be 0-%d or -1 fro all)", i)
	}
	ctx := gousb.NewContext()

	dev, err := ctx.OpenDeviceWithVIDPID(gousb.ID(d.vID), gousb.ID(d.pID))
	if err != nil {
		return err
	}
	defer ctx.Close()

	dev.SetAutoDetach(true)

	cmdBuffer := make([]byte, 10)
	if state {
		cmdBuffer[0] = 0xFF
	} else {
		cmdBuffer[0] = 0xFD
	}
	cmdBuffer[1] = uint8(i)

	_, err = setReport(dev, cmdBuffer)

	return err
}

func (d *Device) Toggle(i int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if i < 0 || i > int(d.numRelays) {
		return fmt.Errorf("invalid usbDevice number. Must be 0-%d or -1 fro all)", i)
	}

	return d.onOff(!d.state[i], i)
}

func (d *Device) On(i int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.onOff(true, i)
}

func (d *Device) Off(i int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.onOff(false, i)
}

func (d *Device) NumRelays() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return int(d.numRelays)
}

func setReport(dev *gousb.Device, buffer []byte) (int, error) {
	return dev.Control(
		USB_TYPE_CLASS|USB_RECIP_DEVICE|USB_ENDPOINT_OUT,
		USBRQ_HID_SET_REPORT,
		USB_HID_REPORT_TYPE_FEATURE<<8|(0&0xff),
		0,
		buffer,
	)
}

func readStatus(dev *gousb.Device) (uint8, error) {
	buff := make([]byte, 10)

	_, err := dev.Control(
		USB_TYPE_CLASS|USB_RECIP_DEVICE|USB_ENDPOINT_IN,
		USBRQ_HID_SET_REPORT,
		USB_HID_REPORT_TYPE_FEATURE<<8|0,
		0,
		buff,
	)
	return buff[8], err
}
