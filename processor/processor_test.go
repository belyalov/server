package processor

import (
	"bytes"
	"hash/crc32"
	"testing"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/encode"
	"github.com/stretchr/testify/assert"
)

func TestMalformedMessage(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("some_totally_malformed_message_not_protobuf")

	err := ProcessMessage(&Message{Payload: buf.Bytes()})
	assert.EqualError(t, err, "Invalid message length: 115, max 42")
}

func TestInvalidCRC(t *testing.T) {
	// Generate valid header + junk payload so CRC will fail
	hdr := &openiot.Header{
		DeviceId: 111,
		Crc:      111,
	}

	var buf bytes.Buffer
	encode.WriteSingleMessage(&buf, hdr)
	buf.WriteString("somejunkpayload")

	err := ProcessMessage(&Message{Payload: buf.Bytes()})
	assert.EqualError(t, err, "CRC check failed")
}

func TestUnknownDevice(t *testing.T) {
	hdr := &openiot.Header{
		DeviceId: 0xffff,
	}
	msg := &openiot.MessageInfo{
		Sequence: 111,
	}

	payload, err := encode.MakeReadyToSendMessage(hdr, openiot.EncryptionType_PLAIN, nil, msg)
	assert.NoError(t, err)

	err = ProcessMessage(&Message{Payload: payload})
	assert.EqualError(t, err, "Device 0xffff is not registered")
}

func TestUnknownDeviceMessage(t *testing.T) {
	// Add dummy device
	err := device.AddDevice(&device.Device{
		ID:           0xff,
		ProtobufName: "qqqq",
	})
	assert.NoError(t, err)
	defer device.DeleteAllDevices()

	// Send message with dummy device as dst
	hdr := &openiot.Header{
		DeviceId: 0xff,
	}

	var buf bytes.Buffer
	encode.WriteSingleMessage(&buf, hdr)

	err = ProcessMessage(&Message{Payload: buf.Bytes()})
	assert.EqualError(t, err, "0xff: Protobuf 'qqqq' is not registered")
}

func TestDeviceMessageDeserializeError(t *testing.T) {
	// Add dummy device
	err := device.AddDevice(&device.Device{
		ID:           0xff,
		ProtobufName: "openiot.JoinRequest",
	})
	assert.NoError(t, err)
	defer device.DeleteAllDevices()

	// Craft message with valid CRC of device message, but message itself is junk
	hdr := &openiot.Header{
		DeviceId: 0xff,
		Crc:      crc32.ChecksumIEEE([]byte("junk")),
	}
	var buf bytes.Buffer
	encode.WriteSingleMessage(&buf, hdr)
	buf.WriteString("junk")

	err = ProcessMessage(&Message{
		Payload: buf.Bytes(),
		Source:  &mockTransport{},
	})
	assert.EqualError(t, err, "0xff: decrypt/deserialize failed: Invalid message length: 106, max 3")
}

func TestDeviceMessageDuplicate(t *testing.T) {
	// Add dummy device
	dev := &device.Device{
		ID:           0xff,
		ProtobufName: "openiot.JoinRequest",
	}
	err := device.AddDevice(dev)
	assert.NoError(t, err)
	defer device.DeleteAllDevices()

	// Craft device message: header + info + joinrequest
	hdr := &openiot.Header{
		DeviceId: dev.ID,
	}
	info := &openiot.MessageInfo{
		Sequence: 1,
	}
	request := &openiot.JoinRequest{}
	payload, err := encode.MakeReadyToSendMessage(hdr, openiot.EncryptionType_PLAIN, nil, info, request)
	assert.NoError(t, err)

	// Send it 2 times: second packet should be dropped
	err = ProcessMessage(&Message{
		Payload: payload,
		Source:  &mockTransport{},
	})
	assert.NoError(t, err)
	err = ProcessMessage(&Message{
		Payload: payload,
		Source:  &mockTransport{},
	})
	assert.EqualError(t, err, "0xff: drop duplicate packet seq 1 (last seq 1)")
}
