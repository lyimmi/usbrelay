package main

import (
	"fmt"
	"github.com/lyimmi/usbrelay"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

// commands
const (
	list   = "list"
	on     = "on"
	off    = "off"
	toggle = "toggle"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "usage: usbrelay <command> [<args>]\n")
		fmt.Fprintln(w, "Serial numbers are case sensitive.")
		fmt.Fprintln(w, `Use "all" as a relay number to set all relays at once.`)
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Available commands:")
		fmt.Fprintln(w, "\tlist\tList all available devices (add -s flag to print a simplified output)")
		fmt.Fprintln(w, "\ton\t<serial> <relay no.>\tset a relay's state to ON")
		fmt.Fprintln(w, "\toff\t<serial> <relay no.>\tset a relay's state to OFF")
		fmt.Fprintln(w, "\ttoggle\t<serial> <relay no.>\ttoggle a relay's state")
		w.Flush()
		return
	}

	switch args[0] {
	case list:
		devices, err := usbrelay.Enumerate()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if len(devices) < 1 {
			fmt.Println(usbrelay.ErrNoDeviceFound)
			os.Exit(1)
		}

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
	case on, off, toggle:
		err := callFunc(args[0], args[1], args[2])
		if err != nil {
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

	cmd = strings.ToLower(cmd)
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
	if err != nil {
		return err
	}

	err = device.Open(cmd == toggle)
	if err != nil {
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
