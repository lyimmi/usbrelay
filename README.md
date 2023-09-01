# usbrelay

A package for controlling USB relay boards with HID API in go.
This implementation is capable of controlling 1-12 channel relay boards.

**This repository is work in progress, the public interface and behavior may change.**

This packages uses [karalabe/hid](https://github.com/karalabe/hid) to access and control the USB device, karalabe/hid 
has a simple interface, embeds `libusb` and that makes this package `go get`-able.

## Usage

```shell
go get -u github.com/lyimmi/usbrelay
```

### Requirements

This package requires CGO to be built.

### Examples

Simple example:

```golang
import (
    "time"
    "github.com/lyimmi/usbrelay"
)

func main() {
    devices, _ := usbrelay.Enumerate()
    device := devices[0]
	
    device.Open(true)
    defer device.Close()
	
    device.On(usbrelay.R1)
    time.Sleep(500 * time.Millisecond)
    device.Off(usbrelay.R1)
}
```

A more detailed example can be found in the [example directory](https://github.com/lyimmi/usbrelay/blob/main/example/main.go).

### Permissions

On linux USB access needs root permissions by default. If you don't want to run your code as root, it can be done via 
udev rules:

TLDR:
1. Create a rule file: 
   - `/etc/udev/rules.d/xx-my-rule.rules`
   - xx is any number > 50 (the defaults are in 50, and higher numbers take priority)
   - my-rule is whatever you want to call it 
   - must end in .rules
2. Insert the following rule:
   - `SUBSYSTEM=="usb", ATTRS{idVendor}=="16c0", ATTRS{idProduct}=="05df", TAG+="uaccess"`
3. Save and run:
   - `sudo udevadm control --reload-rules && sudo udevadm trigger`
4. Make sure to unplug and replug the USB device!

## Credits

This package is an amalgamation of a few people's previous work:

- https://github.com/pavel-a/usb-relay-hid
- https://github.com/karalabe/hid
- https://github.com/e61983/go-usb-relay
- https://github.com/spetr/hidrelay

## License

Based on [this article](https://en.wikipedia.org/wiki/Open-source_license) this package is under the [MIT license](https://github.com/lyimmi/usbrelay/blob/main/LICENSE).
