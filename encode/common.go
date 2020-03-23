package encode

import (
	"bytes"
	"hash/crc32"

	"github.com/golang/protobuf/proto"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/device"
)

// MakeReadyToSendDeviceMessage makes ready to be send bytes using device's settings and proto message
// It does:
// - Serializes protobuf message
// - Calculates CRC
// - Makes MessageInfo header
// - Encrypts all (if enabled on device)
// - Makes Header
// - Writes all messages into buffer and returns bytes
func MakeReadyToSendDeviceMessage(dev *device.Device, msg proto.Message) ([]byte, error) {
	// Increase send sequence: In order to be able to filter duplicates
	// remove device tracks last received sequence and ignores messages
	// that has already been processed.
	dev.SequenceSend++

	hdr := &openiot.Header{
		DeviceId: dev.ID,
	}
	info := &openiot.MessageInfo{
		Sequence: dev.SequenceSend,
	}

	return MakeReadyToSendMessage(hdr, dev.EncryptionType, dev.Key(), info, msg)
}

// MakeReadyToSendMessage makes message ready to be send
// It does:
// - Serializes all msgs
// - Calculates CRC
// - Optionally encrypts messages
// - Writes all messages into buffer
func MakeReadyToSendMessage(
	hdr *openiot.Header, enc openiot.EncryptionType, key []byte, msgs ...proto.Message) ([]byte, error) {
	// Serialize (with optional encryption) all messages:
	var msgBuf bytes.Buffer
	if err := WriteAndEncrypt(&msgBuf, enc, key, msgs...); err != nil {
		return nil, err
	}

	// Update CRC
	hdr.Crc = crc32.ChecksumIEEE(msgBuf.Bytes())
	// Write header + all [optionally encrypted] messages into one buffer
	var buf bytes.Buffer
	if err := WriteSingleMessage(&buf, hdr); err != nil {
		return nil, err
	}
	if _, err := msgBuf.WriteTo(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
