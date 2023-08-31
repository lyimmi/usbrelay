package main

import (
	"log"
	"time"
	"usbrelay"
)

func main() {
	devices, err := usbrelay.Enumerate()

	if len(devices) < 1 {
		log.Fatalln("no device detected")
	}

	device := devices[0]

	err = device.Open(true)
	if err != nil {
		log.Fatal(err)
	}
	defer device.Close()

	if err = device.On(usbrelay.R_ALL); err != nil {
		log.Fatalln(err)
	}

	time.Sleep(1 * time.Second)

	if err = device.Off(usbrelay.R_ALL); err != nil {
		log.Fatalln(err)
	}

	time.Sleep(1 * time.Second)

	for i := 0; i <= device.RelayCount(); i++ {
		ch := usbrelay.RelayNumber(i)

		if err = device.On(ch); err != nil {
			log.Fatalln(err)
		}

		time.Sleep(150 * time.Millisecond)

		if err := device.Off(ch); err != nil {
			log.Fatalln(err)
		}
	}

	time.Sleep(1 * time.Second)

	for i := device.RelayCount(); i >= 0; i-- {
		ch := usbrelay.RelayNumber(i)

		if err = device.Toggle(ch); err != nil {
			log.Fatalln(err)
		}

		time.Sleep(150 * time.Millisecond)

		if err := device.Toggle(ch); err != nil {
			log.Fatalln(err)
		}
	}
}
