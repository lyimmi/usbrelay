package usbrelay

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
*/
import "C"
import (
	"fmt"
	"github.com/google/gousb"
	"strconv"
	"strings"
)

func Enumerate() ([]*Device, error) {
	// Initialize a new Context.
	ctx := gousb.NewContext()
	defer ctx.Close()

	ctx.Debug(0)
	// Iterate through available Devices, finding all that match a known VID/PID.
	//vid, pid := gousb.ID(CfgVendorID), gousb.ID(CfgDeviceID)
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// this function is called for every device present.
		// Returning true means the device should be opened.
		return true
	})
	// All returned devices are now open and will need to be closed.
	for _, d := range devs {
		defer d.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("OpenDevices(): %v", err)
	}

	devices := make([]*Device, 0)

	for _, d := range devs {
		p, _ := d.Product()

		if len(p) != len(RelayNamePref)+1 {
			continue
		}

		if !strings.HasPrefix(p, RelayNamePref) {
			continue
		}

		numRelaysStr, found := strings.CutPrefix(p, RelayNamePref)
		if !found {
			continue
		}

		numRelays, err := strconv.Atoi(numRelaysStr)
		if err != nil {

		}
		if numRelays < 0 || numRelays > 8 {
			return nil, fmt.Errorf("Unknown usbDevice device? num relays=%d\n", numRelays)
		}
		devices = append(devices, NewDevice(int16(d.Desc.Vendor), int16(d.Desc.Product), uint8(numRelays)))
	}

	return devices, nil
}
