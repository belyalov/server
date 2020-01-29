package device

import (
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
	Key             []byte   `yaml:"-" mapstructure:"-"`
	SequenceSend    uint32   `yaml:"sequence_send"`
	SequenceReceive uint32   `yaml:"sequence_receive"`
	HandlerNames    []string `yaml:"handlers"`
	TransportName   string   `yaml:"transport"`

	transport transport.Transport
	handlers  []Handler
}

// NewDevice creates "unknown" device.
func NewDevice(id uint64) *Device {
	return &Device{
		ID:           id,
		Name:         "Unknown Device",
		Manufacturer: "Unknown",
	}
}

// AddDeviceHandler sets new device handler
func (dev *Device) AddDeviceHandler(handler Handler) error {
	// Ensure that new handler is in present yet
	for _, value := range dev.handlers {
		if handler.GetName() == value.GetName() {
			return fmt.Errorf("Handler '%s' already exists in device %x",
				handler.GetName(), dev.ID)
		}
	}

	dev.handlers = append(dev.handlers, handler)
	handler.AddDevice(dev)

	return nil
}

// SetTransport sets new transport
func (dev *Device) SetTransport(transport transport.Transport) {
	dev.transport = transport
}

// Handlers return array of associated device's handlers
func (dev *Device) Handlers() []Handler {
	return dev.handlers
}

// Transport returns device's handler
func (dev *Device) Transport() transport.Transport {
	return dev.transport
}
