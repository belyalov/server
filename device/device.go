package device

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/transport"
)

// Device is single IoT device
type Device struct {
	// Next parameters will be written to YAML configuration
	ID              uint64 `yaml:"-"`
	IDhex           string `yaml:"id"`
	Name            string
	DisplayName     string `yaml:"display_name"`
	Manufacturer    string
	ProductURL      string   `yaml:"product_url"`
	KeyString       string   `yaml:"key"`
	SequenceSend    uint32   `yaml:"sequence_send"`
	SequenceReceive uint32   `yaml:"sequence_receive"`
	ProtobufName    string   `yaml:"protobuf_name"`
	HandlerNames    []string `yaml:"handlers"`
	TransportName   string   `yaml:"transport"`
	EncryptionType  openiot.EncryptionType

	key       []byte
	transport transport.Transport
	handlers  []Handler
}

// NewDevice creates "unknown" device.
func NewDevice(id uint64) *Device {
	return &Device{
		ID:           id,
		IDhex:        fmt.Sprintf("0x%x", id),
		Name:         "Unknown Device",
		Manufacturer: "Unknown",
	}
}

// AddHandler sets new device handler
func (dev *Device) AddHandler(name string) {
	for _, value := range dev.HandlerNames {
		if name == value {
			// Duplicate
			return
		}
	}

	dev.HandlerNames = append(dev.HandlerNames, name)

	if handler := FindHandlerByName(name); handler != nil {
		dev.handlers = append(dev.handlers, handler)
		handler.AddDevice(dev)
	}
}

// SetHandler sets device handler (replaces existing)
func (dev *Device) SetHandler(name string) {
	dev.HandlerNames = []string{name}

	if handler := FindHandlerByName(name); handler != nil {
		dev.handlers = []Handler{handler}
		handler.AddDevice(dev)
	}
}

// Handlers return array of associated device's handlers
func (dev *Device) Handlers() []Handler {
	return dev.handlers
}

// SetKey set device's encryption key
func (dev *Device) SetKey(key []byte) {
	dev.key = key
	dev.KeyString = hex.EncodeToString(key)
}

// Key returns current device's encryption key
func (dev *Device) Key() []byte {
	return dev.key
}

// SetTransport sets new transport
func (dev *Device) SetTransport(transport transport.Transport) {
	dev.transport = transport
	dev.TransportName = transport.GetName()
}

// Transport returns device's handler
func (dev *Device) Transport() transport.Transport {
	return dev.transport
}

// fixParameters re-calculates non YAMLified parameters
// e.g. key is stored as string, but bytes are used here
func (dev *Device) fixParameters() error {
	// Set ID from hex string
	if id, err := strconv.ParseUint(dev.IDhex, 0, 64); err == nil {
		dev.ID = id
	} else {
		return err
	}
	// Setup Key from hex string representation
	if key, err := hex.DecodeString(dev.KeyString); err == nil {
		dev.key = key
	} else {
		return err
	}
	// Setup transport
	if transport := transport.FindTransportByName(dev.TransportName); transport != nil {
		dev.SetTransport(transport)
	}
	// Setup handlers
	for _, name := range dev.HandlerNames {
		if handler := FindHandlerByName(name); handler != nil {
			dev.handlers = append(dev.handlers, handler)
		} else {
			return fmt.Errorf("unknown handler '%s'", name)
		}
	}

	return nil
}
