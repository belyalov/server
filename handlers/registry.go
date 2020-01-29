package handlers

import (
	"fmt"
	"sync"

	"github.com/open-iot-devices/server/device"
)

var deviceHandlersByName map[string]device.Handler
var deviceHandlersInit sync.Once

// FindDeviceHandler lookups device by name. Returns nil if not found.
func FindDeviceHandler(name string) device.Handler {
	return deviceHandlersByName[name]
}

// MustFindDeviceHandler lookups device by name. Panics if not found.
func MustFindDeviceHandler(name string) device.Handler {
	if handler := FindDeviceHandler(name); handler == nil {
		panic(fmt.Sprintf("Device handler '%s' not found.", name))
	} else {
		return handler
	}
}

// MustAddDeviceHandler adds new device handler into registry
// Panics in case of error
func MustAddDeviceHandler(dev device.Handler) {
	deviceHandlersInit.Do(func() {
		deviceHandlersByName = make(map[string]device.Handler)
	})
	if _, ok := deviceHandlersByName[dev.GetName()]; ok {
		panic(fmt.Sprintf("Device Handler '%s' already exists.", dev.GetName()))
	}

	deviceHandlersByName[dev.GetName()] = dev
}
