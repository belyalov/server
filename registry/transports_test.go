package registry

import (
	"testing"

	"github.com/open-iot-devices/server/transport"
	"github.com/stretchr/testify/assert"
)

func TestTransport(t *testing.T) {
	// Reset transports
	transportsByName = make(map[string]transport.Transport)

	// Add one
	tr := transport.NewMockTransport("test1")
	MustAddTransport(tr)

	// Add it one more time
	assert.Panics(t, func() {
		MustAddTransport(tr)
	})

	// Lookup it
	assert.NotNil(t, FindTransportByName("test1"))
	assert.Contains(t, GetAllTransports(), "test1")

	// Negative - not found
	assert.Nil(t, FindTransportByName("test111"))
}
