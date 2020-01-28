package registry

import (
	"fmt"

	"github.com/open-iot-devices/server/transport"
)

var transportsByName map[string]transport.Transport

func init() {
	transportsByName = make(map[string]transport.Transport)
}

// FindTransportByName lookups transport by name. Returns nil if not found.
func FindTransportByName(name string) transport.Transport {
	if transport, ok := transportsByName[name]; ok {
		return transport
	}
	return nil
}

// MustAddTransport adds new transport into registry or panics
// if it is already exist.
func MustAddTransport(instance transport.Transport) {
	if _, ok := transportsByName[instance.GetName()]; ok {
		panic(fmt.Sprintf("Transport '%s' already exists", instance.GetName()))
	}

	transportsByName[instance.GetName()] = instance
}

// GetAllTransports returns all registered transports
func GetAllTransports() map[string]transport.Transport {
	return transportsByName
}
