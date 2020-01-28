package device

import "github.com/open-iot-devices/server/device/handlers"

// Device is single IoT device
type Device struct {
	ID           uint64
	Name         string
	Manufacturer string
	ProductURL   string `yaml:"product_url"`
	ProtobufURL  string `yaml:"protobuf_url"`

	// Encryption key
	KeyString string `yaml:"key"`
	Key       []byte `yaml:"-"`

	// Device handler
	HandlerName string                 `yaml:"handler"`
	Handler     handlers.DeviceHandler `yaml:"-"`

	// Sequences used to skip duplicated messages
	SequenceSend    uint32 `yaml:"sequence_send"`
	SequenceReceive uint32 `yaml:"sequence_receive"`
}

// NewUnknownDevice creates unknown device:
// - names set to "unknown"
// - device handler set to "no handler"
func NewUnknownDevice(id uint64) *Device {
	return &Device{
		ID:           id,
		Name:         "Unknown Device",
		Manufacturer: "Unknown",
		Handler:      &handlers.NoHandler{},
		HandlerName:  "NoHandler",
	}
}

// IsUnknown checks device's handler and returns true if it is set to NoHandler
func (d *Device) IsUnknown() bool {
	if _, ok := d.Handler.(*handlers.NoHandler); !ok {
		return true
	}
	return false
}
