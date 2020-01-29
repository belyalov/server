package transport

import (
	"fmt"
	"sync"
)

var transportByName map[string]Transport
var transportInit sync.Once

// FindTransportByName lookups device by name. Returns nil if not found.
func FindTransportByName(name string) Transport {
	return transportByName[name]
}

// MustAddTransport adds new device handler into registry
// Panics in case of error
func MustAddTransport(transport Transport) {
	transportInit.Do(func() {
		transportByName = make(map[string]Transport)
	})
	if _, ok := transportByName[transport.GetName()]; ok {
		panic(fmt.Sprintf("Device Handler '%s' already exists.", transport.GetName()))
	}

	transportByName[transport.GetName()] = transport
}
