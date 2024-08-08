// Package moisture provides module for working with NPH Automation Moisture
// Sensor.
package moisture

import (
	"encoding/binary"
	"errors"
	"log"
	"math"

	"tinygo.org/x/bluetooth"
)

// Moisture sensor UUIDs
const ServiceUUID string = "a255a717-1559-4ded-9a62-0b285f9a5c8a"
const CharacteristicUUID string = "6fe14a79-6e62-491b-b2e0-174762f9ac9b"

// Errors
var ErrDevNotSetUp = errors.New("device was not set up")

// Local Bluetooth adapter
var adapter = bluetooth.DefaultAdapter

// MoistureSensor represents moisture sensor device.
type MoistureSensor struct {
	name           string                         // Device local name
	address        bluetooth.Address              // Sensor's Bluetooth address
	device         bluetooth.Device               // Sensor connection
	service        bluetooth.DeviceService        // BLE Service
	characteristic bluetooth.DeviceCharacteristic // BLE characteristic
	connected      bool                           // Flag if sensor was connected
}

// Connects specified sensor
func (ms *MoistureSensor) Connect() error {
	device, err := adapter.Connect(ms.address, bluetooth.ConnectionParams{})
	if err != nil {
		return err
	}
	ms.device = device

	err = ms.setupSensor()
	if err != nil {
		return err
	}

	ms.connected = true
	return nil
}

// Disconnects sensor
func (ms *MoistureSensor) Disconnect() error {
	err := ms.device.Disconnect()
	if err != nil {
		return err
	}
	ms.connected = false
	return nil
}

// Sets up sensor characteristics & services
func (ms *MoistureSensor) setupSensor() error {
	msService, err := bluetooth.ParseUUID(ServiceUUID)
	if err != nil {
		return err
	}

	services, err := ms.device.DiscoverServices([]bluetooth.UUID{msService})
	if err != nil {
		return err
	}

	ms.service = services[0]

	msChar, err := bluetooth.ParseUUID(CharacteristicUUID)
	if err != nil {
		return err
	}

	chars, err := ms.service.DiscoverCharacteristics([]bluetooth.UUID{msChar})
	if err != nil {
		return err
	}

	ms.characteristic = chars[0]

	return nil
}

// Reads data from sensor to result variable
func (ms MoistureSensor) Read(result *float64) error {
	var emptyService bluetooth.DeviceService
	var emptyChar bluetooth.DeviceCharacteristic

	if ms.service == emptyService {
		return ErrDevNotSetUp
	}

	if ms.characteristic == emptyChar {
		return ErrDevNotSetUp
	}

	buffer := make([]byte, 8)

	_, err := ms.characteristic.Read(buffer)
	if err != nil {
		return err
	}

	rawData := binary.LittleEndian.Uint64(buffer)

	*result = math.Float64frombits(rawData)

	return nil
}

// Finds device with specified local name and returns as ScanResult
func findMoistureSensor(targetName string) (bluetooth.ScanResult, error) {
	var device bluetooth.ScanResult = bluetooth.ScanResult{}

	if err := adapter.Enable(); err != nil {
		return device, err
	}

	defer func() {
		if err := adapter.StopScan(); err != nil {
			log.Println(err)
		} else {
			log.Println("scanning stopped")
		}
	}()

	log.Println("searching...")

	devCh := make(chan int, 1)

	go func() {
		<-devCh
		adapter.StopScan()
	}()

	if err := adapter.Scan(func(a *bluetooth.Adapter, sr bluetooth.ScanResult) {
		log.Println(sr)
		if sr.LocalName() == targetName {
			log.Printf("found device %s\n", targetName)
			device = sr
			devCh <- 1
		}
	}); err != nil {
		return device, err
	}

	return device, nil
}

// Finds and allocates moisture sensor with specified local name
func NewMoistureSensor(name string) (*MoistureSensor, error) {
	sr, err := findMoistureSensor(name)
	if err != nil {
		return nil, err
	}

	return &MoistureSensor{
		name:      name,
		address:   sr.Address,
		connected: false,
	}, nil
}
