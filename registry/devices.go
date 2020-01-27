package registry

import (
	"fmt"
	"sync"

	"github.com/open-iot-devices/server/device"
)

var devicesByID map[uint64]device.Device
var lock sync.RWMutex

func init() {
	devicesByID = make(map[uint64]device.Device)
}

// FindDeviceByID looks up device by id.
// Returns Device or nil, if not found
func FindDeviceByID(id uint64) device.Device {
	lock.RLock()
	defer lock.RUnlock()

	if dev, ok := devicesByID[id]; ok {
		return dev
	}

	return nil
}

// AddDevice adds new device into registry
// Returns error or nil in case of success
func AddDevice(device device.Device) error {
	lock.Lock()
	defer lock.Unlock()

	id := device.GetDeviceID()
	if _, ok := devicesByID[id]; ok {
		return fmt.Errorf("Device with ID %x already exists in registry", id)
	}

	devicesByID[id] = device

	return nil
}

// DeleteDevice deletes device from registry
// Returns nil on success, error otherwise
func DeleteDevice(id uint64) error {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := devicesByID[id]; !ok {
		return fmt.Errorf("Device with ID %x not found", id)
	}

	delete(devicesByID, id)

	return nil
}
