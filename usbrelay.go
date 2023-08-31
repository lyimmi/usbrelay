package usbrelay

import (
	"fmt"
	"github.com/karalabe/hid"
	"strconv"
	"strings"
)

func Enumerate() ([]*Device, error) {

	deviceInfos := hid.Enumerate(cfgVendorID, cfgDeviceID)
	devices := make([]*Device, 0)

	if len(deviceInfos) <= 0 {
		return devices, nil
	}

	for _, info := range deviceInfos {
		if !strings.HasPrefix(info.Product, relayNamePrefix) {
			continue
		}

		numRelaysStr, found := strings.CutPrefix(info.Product, relayNamePrefix)
		if !found {
			continue
		}

		numRelays, err := strconv.Atoi(numRelaysStr)
		if err != nil {
			return devices, err
		}
		if numRelays < 0 || numRelays > 8 {
			return nil, fmt.Errorf("Unknown usbDevice device? num relays=%d\n", numRelays)
		}
		devices = append(devices, newDevice(&info, numRelays))
	}

	return devices, nil
}
