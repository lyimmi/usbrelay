package usbrelay

import (
	"fmt"
	"github.com/karalabe/hid"
	"strconv"
	"strings"
)

func Enumerate() ([]*Device, error) {
	deviceInfos := hid.Enumerate(DeviceVendorID, DeviceProductID)
	devices := make([]*Device, 0)

	if len(deviceInfos) <= 0 {
		return devices, ErrNoDeviceFound
	}

	var (
		err       error
		numRelays int
	)

	for _, info := range deviceInfos {
		if !strings.HasPrefix(info.Product, relayNamePrefix) {
			continue
		}

		numRelaysStr, found := strings.CutPrefix(info.Product, relayNamePrefix)
		if !found {
			continue
		}

		numRelays, err = strconv.Atoi(numRelaysStr)
		if err != nil {
			return devices, err
		}
		if numRelays < 0 && numRelays != R_ALL {
			return nil, fmt.Errorf("%w num relays: %d", ErrInvalidNumberOfRelays, numRelays)
		}

		var device *Device
		device, err = newDevice(&info, numRelays)
		if err != nil {
			break
		}
		err = device.Open(false)
		if err != nil {
			break
		}

		_, err = device.GetSerialNumber()
		if err != nil {
			break
		}

		devices = append(devices, device)
	}

	for _, device := range devices {
		device.Close()
	}

	if err != nil {
		return nil, err
	}

	if len(devices) <= 0 {
		return devices, ErrNoDeviceFound
	}

	return devices, err
}

func GetDeviceBySerialNumber(sn SerialNumber) (*Device, error) {
	devices, err := Enumerate()
	if err != nil {
		return nil, err
	}

	for _, d := range devices {
		if d.serialNumber == sn {
			return d, nil
		}
	}
	return nil, ErrNoDeviceFound
}
