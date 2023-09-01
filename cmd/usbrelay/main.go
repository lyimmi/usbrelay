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
	setUdevRules = "setudev"
	udevFileName = "/etc/udev/rules.d/70-go.usbrelay-hid.rules"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.StripEscape)
		fmt.Fprintln(w, "usage: usbrelay <command> [<args>]")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "- Serial numbers are case sensitive.")
		fmt.Fprintln(w, `- Use "all" as a relay number to set all relays at once.`)
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Available commands:")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "\tCommand\tParams\tDescription")
		fmt.Fprintln(w, "\tlist\t\tList all available devices (add -s flag to print a simplified output)")
		fmt.Fprintln(w, "\ton\t<serial> <relay>\tSet a relay's state to ON")
		fmt.Fprintln(w, "\toff\t<serial> <relay>\tSet a relay's state to OFF")
		fmt.Fprintln(w, "\ttoggle\t<serial> <relay>\tToggle a relay's state")
		fmt.Fprintln(w, "\tsetudev\t\tSet a udev rule to enable running without sudo (linux only - requires root)")
		w.Flush()
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
			fmt.Fprintln(w, "Serial\tRelays\tVendor\tProduct")
			var (
				serialNumber usbrelay.SerialNumber
				vendorID     int
				productID    int
				relayCount   int
			)
			for _, device := range devices {
				serialNumber, vendorID, productID, relayCount = device.Info()
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", serialNumber, relayCount, vendorID, productID)
			}
			w.Flush()
		}
		return
	case setUdevRules:
		if runtime.GOOS != "linux" {
			fmt.Println("udev rules only available on linux")
			os.Exit(1)
		}
		if err := setUdevRule(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Disconnect and reconnect the device to activate the rule!")
		return
	case on, off, toggle:
		if err := callFunc(cmd, args[1], args[2]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	default:
		fmt.Println("unknown command")
		os.Exit(1)
	}
}

func callFunc(cmd string, serialNumber string, relayNumber string) error {
	var rn usbrelay.RelayNumber
	if relayNumber == "all" {
		rn = usbrelay.R_ALL
	} else {
		rnInt, err := strconv.Atoi(relayNumber)
		if err != nil {
			return err
		}
		rn = usbrelay.RelayNumber(rnInt)
	}

	device, err := usbrelay.GetDeviceBySerialNumber(usbrelay.NewSerialNumber(serialNumber))
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
	}
	return nil
}

func handleNoDeviceErr(err error) {
	if err == nil {
		return
	}
	if errors.Is(err, usbrelay.ErrNoDeviceFound) {
		fmt.Println(fmt.Errorf("%w (try running as root, or set udev rules)", err))
		os.Exit(1)
	}
	fmt.Println(err)
	os.Exit(1)
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
