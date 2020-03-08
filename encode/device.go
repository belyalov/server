package encode

import (
	"bytes"
	"hash/crc32"

	"github.com/golang/protobuf/proto"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/device"
)

// MakeReadyToSendDeviceMessage makes ready to be send message using device's settings.
// It does:
// - Serializes all msgs
// - Calculates CRC
// - Makes MessageInfo header
// - Encrypts all (if enabled on device)
// - Makes MessageHeader
// - Writes all messages into buffer
func MakeReadyToSendDeviceMessage(dev *device.Device, msgs ...proto.Message) ([]byte, error) {
	// Increase send sequence: In order to be able to filter duplicates
	// remove device tracks last received sequence and ignores messages
	// that has already been processed.
	dev.SequenceSend++
	info := &openiot.MessageInfo{
		Sequence: dev.SequenceSend,
	}

	// Serialize (with optional encryption) all messages:
	// info, msgs...
	var serializeBuf bytes.Buffer
	allMsgs := append([]proto.Message{info}, msgs...)
	if err := WriteAndEncrypt(&serializeBuf, dev.EncodingType, dev.Key(), allMsgs...); err != nil {
		return nil, err
	}
	// Make message header
	hdr := &openiot.Header{
		DeviceId: dev.ID,
		Crc:      crc32.ChecksumIEEE(serializeBuf.Bytes()),
	}
	// Write all messages into one buffer
	var buf bytes.Buffer
	if err := WriteSingleMessage(&buf, hdr); err != nil {
		return nil, err
	}
	if _, err := serializeBuf.WriteTo(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
