package transport

import (
	"fmt"
	"io"
	"sync"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

type transportCreateFunc func(name string) Transport

var transportTypes = map[string]transportCreateFunc{}

var transportByName = map[string]Transport{}
var transportLock sync.RWMutex

// FindTransportByName lookups device by name. Returns nil if not found.
func FindTransportByName(name string) Transport {
	return transportByName[name]
}

// AddTransport adds new transport into registry
func AddTransport(transport Transport) error {
	transportLock.Lock()
	defer transportLock.Unlock()

	if _, ok := transportByName[transport.GetName()]; ok {
		return fmt.Errorf("transport '%s' already exists", transport.GetName())
	}
	transportByName[transport.GetName()] = transport

	return nil
}

// DeleteTransport deletes previously registered transport.
func DeleteTransport(name string) error {
	transportLock.Lock()
	defer transportLock.Unlock()

	if _, ok := transportByName[name]; !ok {
		return fmt.Errorf("transport '%s' does not exists", name)
	}
	delete(transportByName, name)

	return nil
}

// GetAllTransports returns all registered transports
func GetAllTransports() []Transport {
	ret := make([]Transport, len(transportByName))

	index := 0
	for _, transport := range transportByName {
		ret[index] = transport
	}

	return ret
}

// Save / Load //

// SaveTransports writes all registered transports in YAML
// format using writer
func SaveTransports(writer io.Writer) error {
	placeHolder := map[string]map[string]interface{}{}

	// Make transports to be structured like:
	// typeName:
	//   transport1:
	//     ...
	//   transport2:
	//     ...
	transportLock.RLock()
	for name, value := range transportByName {
		typeName := value.GetTypeName()
		if _, ok := placeHolder[typeName]; !ok {
			placeHolder[typeName] = make(map[string]interface{})
		}
		placeHolder[typeName][name] = value
	}
	transportLock.RUnlock()

	encoder := yaml.NewEncoder(writer)
	return encoder.Encode(placeHolder)
}

// LoadTransports reads and parses YAML configuration from file
func LoadTransports(reader io.Reader) error {
	placeHolder := map[string]map[string]interface{}{}

	// Decode YAML
	decoder := yaml.NewDecoder(reader)
	decoder.SetStrict(true)
	if err := decoder.Decode(placeHolder); err != nil {
		return err
	}

	// Create transports
	for typeName, transports := range placeHolder {
		newTransport, ok := transportTypes[typeName]
		if !ok {
			return fmt.Errorf("transportType '%s' is unknown", typeName)
		}
		for name, params := range transports {
			transport := newTransport(name)
			mapstructure.Decode(params, transport)
			if err := AddTransport(transport); err != nil {
				return err
			}
		}
	}

	return nil
}

// MustAddTransportType register new transport type
func MustAddTransportType(typeName string, f transportCreateFunc) {
	if _, ok := transportTypes[typeName]; ok {
		panic(fmt.Sprintf("Transport type '%s' already registered", typeName))
	}
	transportTypes[typeName] = f
}
