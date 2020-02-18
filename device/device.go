package device

import (
	"encoding/hex"
	"fmt"

	"github.com/open-iot-devices/server/transport"
)

// Device is single IoT device
type Device struct {
	// Next parameters will be written to YAML configuration
	ID              uint64 `yaml:"-"`
	IDhex           string `yaml:"id"`
	Name            string
	Manufacturer    string
	ProductURL      string   `yaml:"product_url"`
	ProtobufURL     string   `yaml:"protobuf_url"`
	KeyString       string   `yaml:"key"`
	SequenceSend    uint32   `yaml:"sequence_send"`
	SequenceReceive uint32   `yaml:"sequence_receive"`
	HandlerNames    []string `yaml:"handlers"`
	TransportName   string   `yaml:"transport"`
	Protobufs       []string `yaml:"messages"`

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
func (dev *Device) AddHandler(handler Handler) error {
	// Ensure that new handler is in present yet
	for _, value := range dev.handlers {
		if handler.GetName() == value.GetName() {
			return fmt.Errorf("Handler '%s' already exists in device %x",
				handler.GetName(), dev.ID)
		}
	}

	dev.handlers = append(dev.handlers, handler)
	dev.HandlerNames = append(dev.HandlerNames, handler.GetName())
	handler.AddDevice(dev)

	return nil
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
