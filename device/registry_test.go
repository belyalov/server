package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceRegistry(t *testing.T) {
	ReplaceAllDevicesWith(nil)

	dev := NewDevice(111)

	// Add / Lookup / Delete device
	assert.NoError(t, AddDevice(dev))
	assert.Error(t, AddDevice(dev)) // already exists
	assert.Equal(t, dev, FindDeviceByID(111))
	assert.Equal(t, []*Device{dev}, GetAllDevices())
	assert.NoError(t, DeleteDeviceByID(111))
	// Lookup again (device has been deleted)
	assert.Nil(t, FindDeviceByID(111))

	// Negative: no such device
	assert.Nil(t, FindDeviceByID(55666666))
	assert.Error(t, DeleteDeviceByID(55666666))
}
