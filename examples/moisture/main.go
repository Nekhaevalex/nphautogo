// This example demonstates single moisture value reading from sensor 'MS1'.
package main

import (
	"fmt"
	"log"

	moisture "github.com/Nekhaevalex/nphautogo/sensors"
)

func main() {
	// Creating new connection instance for device "MS1"
	ms, err := moisture.NewMoistureSensor("MS1")
	if err != nil {
		log.Fatal(err)
	}

	// Connecting to the device
	if err := ms.Connect(); err != nil {
		log.Fatal(err)
	}

	// Disconnect when finished
	defer func() {
		if err := ms.Disconnect(); err != nil {
			log.Fatal(err)
		}
	}()

	// Create variable for moisture value
	var data float64

	// Reading single value
	err = ms.Read(&data)
	if err != nil {
		log.Fatal(err)
	}

	// Printing value
	fmt.Print(data)
}
