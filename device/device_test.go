package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceHandler(t *testing.T) {
	handler := &mockHandler{
		name: "mock",
	}
	MustAddHandler(handler)
	defer DeleteHandler(handler.name)

	// Add existing / non existing handlers
	dev := NewDevice(1111)
	dev.AddHandler("non_existing")
	dev.AddHandler("mock")
	assert.Equal(t, []string{"non_existing", "mock"}, dev.HandlerNames)
	assert.Equal(t, []Handler{handler}, dev.handlers)

	dev.SetKey([]byte{1, 2, 3, 4, 55})
	// Ensure that set key actually updates 2 fields
	assert.Equal(t, "0102030437", dev.KeyString)
	assert.Equal(t, []byte{1, 2, 3, 4, 55}, dev.Key())

	// Add the same handlers, should be ignored
	dev.AddHandler("mock")
	dev.AddHandler("non_existing")
	assert.Equal(t, []string{"non_existing", "mock"}, dev.HandlerNames)
	assert.Equal(t, []Handler{handler}, dev.handlers)

	// Test SetHandler (overrides)
	dev.SetHandler("mock")
	assert.Equal(t, []string{"mock"}, dev.HandlerNames)
	assert.Equal(t, []Handler{handler}, dev.handlers)
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
}
