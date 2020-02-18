package device

import (
	"fmt"
	"sync"
)

var handlersByName map[string]Handler
var handlersInit sync.Once

// FindHandlerByName lookups device by name. Returns nil if not found.
func FindHandlerByName(name string) Handler {
	return handlersByName[name]
}

// MustAddHandler adds new device handler into registry
// Panics in case of error
func MustAddHandler(dev Handler) {
	handlersInit.Do(func() {
		handlersByName = make(map[string]Handler)
	})
	if _, ok := handlersByName[dev.GetName()]; ok {
		panic(fmt.Sprintf("Device Handler '%s' already exists.", dev.GetName()))
	}

	handlersByName[dev.GetName()] = dev
}

// DeleteHandler deletes registered device handler.
// It is not being used in production, just for tests.
func DeleteHandler(name string) {
	delete(handlersByName, name)
}

// GetAllHandlers returns all registered device handlers
func GetAllHandlers() map[string]Handler {
	return handlersByName
}
