# usbrelay

[![CommandLine](https://github.com/lyimmi/usbrelay/actions/workflows/go-command-line.yml/badge.svg)](https://github.com/lyimmi/usbrelay/actions/workflows/go-command-line.yml) [![Example](https://github.com/lyimmi/usbrelay/actions/workflows/go-example.yml/badge.svg)](https://github.com/lyimmi/usbrelay/actions/workflows/go-example.yml)

A package for controlling USB relay boards with HID API in go, capable of controlling any number of channels on a relay board.

**This package is work in progress, the public interface and behavior may change.**

The package uses [karalabe/hid](https://github.com/karalabe/hid) to access and control the USB device.
karalabe/hid has a simple interface, embeds `libusb` that makes this package `go get`-able.

## Usage

### Requirements

This package requires CGO to be built.

### As a command line tool:

Simply install the tool without cloning it.

```shell
go install github.com/lyimmi/usbrelay/cmd/usbrelay@latest
```

#### Command line usage:

```shell
usbrelay
```

```text
usage: usbrelay <command> [<args>]

- Serial numbers are case sensitive.
- Use "all" as a relay number to set all relays at once.

Available commands:

   Command     Params                  Description
   list                                List all available devices (add -s flag to print a simplified output)
   on          <serial> <relay>        Set a relay's state to ON
   off         <serial> <relay>        Set a relay's state to OFF
   toggle      <serial> <relay>        Toggle a relay's state
   setserial   <serial> <new serial>   Change a device's serial number (max 5 ASCII characters)
   setudev                             Set a udev rule to enable running without sudo (linux only - requires root)

```

### As a package:

```shell
go get -u github.com/lyimmi/usbrelay
```

### Examples

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

## Testing

Testing is done via a mocked virtual USB device with [umockdev](https://github.com/martinpitt/umockdev).

A 4 channel device was used to record the USB communication with [wireshark](https://www.wireshark.org/) and `usbmon` 
while executing the command line tool.

Each command's communication is stored in a separate `pcap` file in the test directory and replayed for the command line
tool with `umockdev-run` using the actual device's stored sysfs device and udev properties.

## Credits

This package is an amalgamation of a few people's previous work:

- https://github.com/pavel-a/usb-relay-hid
- https://github.com/karalabe/hid
- https://github.com/e61983/go-usb-relay
- https://github.com/spetr/hidrelay

## License

Based on [this article](https://en.wikipedia.org/wiki/Open-source_license#Types) this package is under the [MIT license](https://github.com/lyimmi/usbrelay/blob/main/LICENSE).
