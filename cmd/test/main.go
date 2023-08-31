package main

import (
	"time"
	"usbrelay"
)

func main() {
	devices, err := usbrelay.Enumerate()
	device := devices[0]

	for i := 0; i <= device.NumRelays(); i++ {
		err = device.On(i)
		if err != nil {
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	for i := 0; i <= device.NumRelays(); i++ {
		err = device.Off(i)
		if err != nil {
			return
		}

		time.Sleep(100 * time.Millisecond)
	}
}
