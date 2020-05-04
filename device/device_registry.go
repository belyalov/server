package device

import (
	"fmt"
	"io"
	"sync"

	"gopkg.in/yaml.v2"
)

var devicesByID = map[uint64]*Device{}
var deviceLock sync.RWMutex

// FindDeviceByID looks up device by id.
// Returns Device or nil if not found
func FindDeviceByID(id uint64) *Device {
	deviceLock.RLock()
	defer deviceLock.RUnlock()

	return devicesByID[id]
}

// AddDevice adds new device into registry
func AddDevice(device *Device) error {
	deviceLock.Lock()
	defer deviceLock.Unlock()

	if _, ok := devicesByID[device.ID]; ok {
		return fmt.Errorf("Device with ID %x already exists in registry", device.ID)
	}

	devicesByID[device.ID] = device

	return nil
}

// DeleteAllDevices deletes all registered devices
func DeleteAllDevices() {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	devicesByID = map[uint64]*Device{}
}

// DeleteDeviceByID deletes device from registry
func DeleteDeviceByID(id uint64) error {
	deviceLock.Lock()
	defer deviceLock.Unlock()

	if _, ok := devicesByID[id]; !ok {
		return fmt.Errorf("Device with ID %x not found", id)
	}

	delete(devicesByID, id)

	return nil
}

// GetAllDevices returns all registered devices in array
func GetAllDevices() []*Device {
	var index int
	res := make([]*Device, len(devicesByID))

	deviceLock.RLock()
	defer deviceLock.RUnlock()

	for _, dev := range devicesByID {
		res[index] = dev
		index++
	}

	return res
}

// SaveDevices writes all registered transports in YAML
// format using writer
func SaveDevices(writer io.Writer) error {
	deviceLock.RLock()
	deviceLock.RUnlock()

	// Flatten devices map into array
	placeholder := make([]*Device, len(devicesByID))
	index := 0
	for _, dev := range devicesByID {
		placeholder[index] = dev
		index++
	}

	encoder := yaml.NewEncoder(writer)
	return encoder.Encode(placeholder)
}

// LoadDevices reads and parses YAML configuration from file
func LoadDevices(reader io.Reader) error {
	// Decode YAML
	var devices []*Device
	decoder := yaml.NewDecoder(reader)
	decoder.SetStrict(true)
	if err := decoder.Decode(&devices); err != nil {
		return err
	}

	// Restore some non YAMLified parameters
	for _, dev := range devices {
		if err := dev.fixParameters(); err != nil {
			return err
		}
	}

	// Replace registry with new set of devices
	deviceLock.Lock()
	devicesByID = map[uint64]*Device{}
	for _, dev := range devices {
		devicesByID[dev.ID] = dev
	}
	deviceLock.Unlock()

	return nil
}
