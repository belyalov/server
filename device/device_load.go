package device

import (
	"encoding/hex"
	"fmt"
	"io"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/open-iot-devices/server/transport"
	"gopkg.in/yaml.v2"
)

// CreateDeviceFromYaml creates Device from YAML definition
func CreateDeviceFromYaml(reader io.Reader) (*Device, error) {
	// Deserialize YAML
	dev := &Device{}
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(dev); err != nil {
		return nil, err
	}
	// Set ID from hex string
	if id, err := strconv.ParseUint(dev.IDhex, 0, 64); err == nil {
		dev.ID = id
	} else {
		return nil, err
	}
	// Setup key from hex string representation
	if key, err := hex.DecodeString(dev.KeyString); err == nil {
		dev.key = key
	} else {
		return nil, err
	}
	// Setup transport
	if transport := transport.FindTransportByName(dev.TransportName); transport != nil {
		dev.SetTransport(transport)
	}
	// Check that device's protobufs are registered
	for _, name := range dev.Protobufs {
		if proto.MessageType(name) == nil {
			return nil, fmt.Errorf("Protobuf '%s' unknown", name)
		}
	}
	// Setup handlers: 2 steps: ensure that all handlers exists / then setup them all
	for _, name := range dev.HandlerNames {
		if handler := FindHandlerByName(name); handler == nil {
			return nil, fmt.Errorf("DeviceHandler '%s' not found", name)
		}
	}
	for _, name := range dev.HandlerNames {
		handler := FindHandlerByName(name)
		dev.handlers = append(dev.handlers, handler)
		handler.AddDevice(dev)
	}

	return dev, nil
}
