package config

import (
	"bytes"
	"testing"

	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/handlers"
	"github.com/open-iot-devices/server/transport"
	"github.com/stretchr/testify/assert"
)

func TestDevices(t *testing.T) {
	handler1 := &mockHandler{
		name: "config_handler1",
	}
	handlers.MustAddDeviceHandler(handler1)
	handler2 := &mockHandler{
		name: "config_handler2",
	}
	handlers.MustAddDeviceHandler(handler2)

	transport1 := &mockTransport{
		name: "config_transport1",
	}
	transport.MustAddTransport(transport1)
	transport2 := &mockTransport{
		name: "config_transport2",
	}
	transport.MustAddTransport(transport2)

	devices := []*device.Device{
		&device.Device{
			ID:              111,
			Name:            "dev111",
			Manufacturer:    "noname",
			ProductURL:      "http://product",
			ProtobufURL:     "http://protobuf",
			Key:             []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			SequenceSend:    11,
			SequenceReceive: 111,
		},
		&device.Device{
			ID:              222,
			Name:            "dev222",
			Manufacturer:    "shmooble",
			ProductURL:      "http://shmooble",
			ProtobufURL:     "http://shmooble_proto",
			Key:             []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			SequenceSend:    22,
			SequenceReceive: 222,
		},
		&device.Device{
			ID: 333,
		},
	}

	// Setup handlers
	devices[0].AddDeviceHandler(handler1)
	devices[1].AddDeviceHandler(handler1)
	devices[1].AddDeviceHandler(handler2)

	// Setup transport
	devices[0].SetTransport(transport1)
	devices[1].SetTransport(transport2)

	// Serialize
	var buf bytes.Buffer
	err := saveDevicesToYaml(&buf, devices)
	assert.NoError(t, err)

	// Deserialize back
	newDevices, err := loadDevicesFromYaml(&buf)
	assert.NoError(t, err)

	// Ensure that devices equal
	assert.Equal(t, devices[0], newDevices[0])
}
