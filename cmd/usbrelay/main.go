package main

import (
	"errors"
	"fmt"
	"github.com/lyimmi/usbrelay"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"text/tabwriter"
)

// commands
const (
	list         = "list"
	on           = "on"
	off          = "off"
	toggle       = "toggle"
	setSerial    = "setserial"
	setUdevRules = "setudev"
	udevFileName = "/etc/udev/rules.d/70-go.usbrelay-hid.rules"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.StripEscape)
		texts := []string{
			"usage: usbrelay <command> [<args>]",
			"",
			"- Serial numbers are case sensitive.",
			`- Use "all" as a relay number to set all relays at once.`,
			"",
			"Available commands:",
			"",
			"\tCommand\tParams\tDescription",
			"\tlist\t\tList all available devices (add -s flag to print a simplified output)",
			"\ton\t<serial> <relay>\tSet a relay's state to ON",
			"\toff\t<serial> <relay>\tSet a relay's state to OFF",
			"\ttoggle\t<serial> <relay>\tToggle a relay's state",
			"\tsetserial\t<serial> <new serial>\tChange a device's serial number (max 5 ASCII characters)",
			"\tsetudev\t\tSet a udev rule to enable running without sudo (linux only - requires root)",
		}

		for _, text := range texts {
			if _, err := fmt.Fprintln(w, text); err != nil {
				printFatal(err)
			}
		}

		if err := w.Flush(); err != nil {
			printFatal(err)
		}
		return
	}

	cmd := strings.ToLower(args[0])

	switch cmd {
	case list:
		devices, err := usbrelay.Enumerate()
		handleNoDeviceErr(err)

		if len(args) > 1 && args[1] == "-s" {
			for _, device := range devices {
				fmt.Println(device.String())
			}
		} else {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			if _, err = fmt.Fprintln(w, "Serial\tRelays\tVendor\tProduct"); err != nil {
				printFatal(err)
			}
			var (
				serialNumber usbrelay.SerialNumber
				vendorID     int
				productID    int
				relayCount   int
			)
			for _, device := range devices {
				serialNumber, vendorID, productID, relayCount = device.Info()
				if _, err = fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", serialNumber, relayCount, vendorID, productID); err != nil {
					printFatal(err)
				}
			}
			if err = w.Flush(); err != nil {
				printFatal(err)
			}
		}
		return
	case setUdevRules:
		if runtime.GOOS != "linux" {
			printFatal(errors.New("udev rules only available on linux"))
		}
		if err := setUdevRule(); err != nil {
			printFatal(err)
		}
		fmt.Println("Disconnect and reconnect the device to activate the rule!")
		return
	case on, off, toggle, setSerial:
		if len(args) < 3 {
			printFatal(fmt.Errorf("too few arguments to call %s\n", cmd))
		}
		if err := callFunc(cmd, args[1], args[2]); err != nil {
			printFatal(err)
		}
		return
	default:
		printFatal(errors.New("unknown command"))
	}
}

func callFunc(cmd string, serialNumber string, relayNumber string) error {
	var rn usbrelay.RelayNumber
	if relayNumber == "all" {
		rn = usbrelay.R_ALL
	} else if cmd != setSerial {
		rnInt, err := strconv.Atoi(relayNumber)
		if err != nil {
			return err
		}
		rn = usbrelay.RelayNumber(rnInt)
	}
	sn := usbrelay.NewSerialNumber(serialNumber)
	device, err := usbrelay.GetDeviceBySerialNumber(sn)
	handleNoDeviceErr(err)

	if err = device.Open(cmd == toggle); err != nil {
		return err
	}
	defer device.Close()

	switch cmd {
	case on:
		return device.On(rn)
	case off:
		return device.Off(rn)
	case toggle:
		return device.Toggle(rn)
	case setSerial:
		return device.SetSerialNumber(usbrelay.NewSerialNumber(relayNumber))
	}
	return nil
}

func handleNoDeviceErr(err error) {
	if err == nil {
		return
	}
	if errors.Is(err, usbrelay.ErrNoDeviceFound) {
		printFatal(fmt.Errorf("%w, maybe a bad serial number or try running as root (or set udev rules)", err))
	}
	printFatal(err)
}

func setUdevRule() error {
	exists, err := fileExists(udevFileName)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("udev rule already exists")
	}

	data := fmt.Sprintf(
		`SUBSYSTEM=="usb", ATTRS{idVendor}=="%s", ATTRS{idProduct}=="%s", TAG+="uaccess"`,
		usbrelay.UdevVendorID,
		usbrelay.UdevProductID,
	)
	if err := os.WriteFile(udevFileName, []byte(data), 0644); err != nil {
		return err
	}
	cmd := exec.Command("sh", "-c", "udevadm control --reload-rules && udevadm trigger")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, stdoutStderr)
	}

	return nil
}

func fileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

func printFatal(err error) {
	fmt.Println(err)
	os.Exit(1)
}
