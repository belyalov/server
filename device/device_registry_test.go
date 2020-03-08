package device

import (
	"bytes"
	"testing"

	"github.com/open-iot-devices/server/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceRegistry(t *testing.T) {
	DeleteAllDevices()

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

var testConfig = `- id: "0x112233"
  name: Unknown Device
  manufacturer: Unknown
  product_url: www
  protobuf_url: proto_www
  key: "010203040506070809"
  sequence_send: 10
  sequence_receive: 11
  handlers:
  - hmock
  transport: tmock
  messages: []
  encodingtype: 0
- id: "0x556677"
  name: Unknown Device
  manufacturer: Unknown
  product_url: www2
  protobuf_url: proto_www2
  key: 0b16212c37424d5863
  sequence_send: 20
  sequence_receive: 22
  handlers: []
  transport: ""
  messages: []
  encodingtype: 0
`

func TestDeviceRegistryLoadSave(t *testing.T) {
	devicesByID = map[uint64]*Device{}
	handlersByName = map[string]Handler{}

	// Register mocks
	mh := &mockHandler{name: "hmock"}
	mt := &mockTransport{name: "tmock"}
	MustAddHandler(mh)
	transport.AddTransport(mt)
	defer transport.DeleteTransport("tmock")

	// "Load" devices from YAML
	reader := bytes.NewReader([]byte(testConfig))
	require.NoError(t, LoadDevices(reader))

	// Save it back to YAML
	var writer bytes.Buffer
	assert.NoError(t, SaveDevices(&writer))

	assert.Equal(t, testConfig, writer.String())
}
