package device

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	_ "github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/transport"
)

var minimalYamlData = `
id: "0x1020304050"
key: "01020304050506070809000102030405"
`

var yamlData = `id: "0x1020304050"
name: TestDevice
manufacturer: Test
product_url: ""
protobuf_url: Some Url
key: "01020304050506070809000102030405"
sequence_send: 1
sequence_receive: 2
handlers:
- config_handler1
- config_handler2
transport: config_transport1
messages:
- openiot.SystemJoinRequest
- openiot.SystemJoinResponse
`

func TestLoadSaveDevice(t *testing.T) {
	handler1 := &mockHandler{
		name: "config_handler1",
	}
	MustAddHandler(handler1)
	defer DeleteHandler(handler1.name)
	handler2 := &mockHandler{
		name: "config_handler2",
	}
	MustAddHandler(handler2)
	defer DeleteHandler(handler2.name)

	transport1 := &mockTransport{
		name: "config_transport1",
	}
	transport.MustAddTransport(transport1)
	defer transport.DeleteTransport(transport1.name)

	// Load device from YAML
	reader := bytes.NewReader([]byte(yamlData))
	result, err := CreateDeviceFromYaml(reader)
	require.NoError(t, err)

	expected := &Device{
		ID:              0x1020304050,
		IDhex:           "0x1020304050",
		Name:            "TestDevice",
		Manufacturer:    "Test",
		ProductURL:      "",
		ProtobufURL:     "Some Url",
		key:             []byte{1, 2, 3, 4, 5, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5},
		KeyString:       "01020304050506070809000102030405",
		SequenceSend:    1,
		SequenceReceive: 2,
		HandlerNames: []string{
			handler1.name,
			handler2.name,
		},
		handlers: []Handler{
			FindHandlerByName(handler1.name),
			FindHandlerByName(handler2.name),
		},
		TransportName: "config_transport1",
		transport:     transport.FindTransportByName(transport1.name),
		Protobufs: []string{
			"openiot.SystemJoinRequest",
			"openiot.SystemJoinResponse",
		},
	}

	assert.Equal(t, expected, result)

	// Save it back to YAML
	data, err := yaml.Marshal(expected)
	require.NoError(t, err)
	assert.Equal(t, yamlData, string(data))
}

func TestLoadDeviceNegative(t *testing.T) {
	{
		data := "name: \"dddd"
		_, err := CreateDeviceFromYaml(
			bytes.NewReader([]byte(data)),
		)
		assert.EqualError(t, err, "yaml: found unexpected end of stream")
	}
	{
		data := "id: zzz"
		_, err := CreateDeviceFromYaml(
			bytes.NewReader([]byte(data)),
		)
		assert.EqualError(t, err, "strconv.ParseUint: parsing \"zzz\": invalid syntax")
	}
	{
		data := "id: 0x1\nkey: zzz"
		_, err := CreateDeviceFromYaml(
			bytes.NewReader([]byte(data)),
		)
		assert.EqualError(t, err, "encoding/hex: invalid byte: U+007A 'z'")
	}
	{
		data := minimalYamlData + "handlers:\n- qqq"
		_, err := CreateDeviceFromYaml(
			bytes.NewReader([]byte(data)),
		)
		assert.EqualError(t, err, "DeviceHandler 'qqq' not found")
	}
	{
		data := minimalYamlData + "messages:\n- qqq"
		_, err := CreateDeviceFromYaml(
			bytes.NewReader([]byte(data)),
		)
		assert.EqualError(t, err, "Protobuf 'qqq' unknown")
	}
}
