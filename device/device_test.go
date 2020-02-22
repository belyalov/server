package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceHandler(t *testing.T) {
	handler := &mockHandler{}

	dev := NewDevice(1111)
	err := dev.AddHandler(handler)
	assert.NoError(t, err)

	// Ensure that set key actually updates 2 fields
	dev.SetKey([]byte{1, 2, 3, 4, 55})
	assert.Equal(t, "0102030437", dev.KeyString)

	// Add the same handler
	err = dev.AddHandler(handler)
	assert.Error(t, err)

	// Device handler's array is correct
	assert.Equal(t, []Handler{handler}, dev.Handlers())

	// Ensure that handler.AddDevice is actually called
	assert.Equal(t, []*Device{dev}, handler.history)
}

func TestDeviceTransport(t *testing.T) {
	dev := NewDevice(123)
	transport := &mockTransport{}

	dev.SetTransport(transport)

	assert.Equal(t, transport, dev.transport)
	assert.Equal(t, transport.GetName(), dev.TransportName)
	assert.Equal(t, transport, dev.Transport())
}

func TestDeviceFixParameters(t *testing.T) {
	{
		device := &Device{
			IDhex: "zzz",
		}
		err := device.fixParameters()
		assert.EqualError(t, err, "strconv.ParseUint: parsing \"zzz\": invalid syntax")
	}
	{
		device := &Device{
			IDhex:     "0x1",
			KeyString: "zzz",
		}
		err := device.fixParameters()
		assert.EqualError(t, err, "encoding/hex: invalid byte: U+007A 'z'")
	}
	{
		device := &Device{
			IDhex:        "0x1",
			KeyString:    "01020304",
			HandlerNames: []string{"qqq"},
		}
		err := device.fixParameters()
		assert.EqualError(t, err, "unknown handler 'qqq'")
	}
	{
		device := &Device{
			IDhex:     "0x1",
			KeyString: "01020304",
			Protobufs: []string{"qqq"},
		}
		err := device.fixParameters()
		assert.EqualError(t, err, "unknown protobuf 'qqq'")
	}
}
