package encode

import (
	"bytes"
	"hash/crc32"
	"time"

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

	now := time.Now()

	info := &openiot.MessageInfo{
		Sequence: dev.SequenceSend,
		Date:     encodeDate(&now),
		Time:     encodeTime(&now),
	}

	return MakeReadyToSendMessage(hdr, dev.EncryptionType, dev.Key(), info, msg)
}

// MakeReadyToSendMessage makes message ready to be send, it does:
// - Serialize all msgs
// - Calculate CRC
// - Optionally encrypt messages
// - Write all messages into buffer
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

func encodeDate(date *time.Time) uint32 {
	// YY-MM-DD-WD
	var res uint32
	weekday := date.Weekday()
	if weekday == time.Sunday {
		res = 0x07
	} else {
		res = uint32(weekday)
	}
	year, month, day := date.Date()
	res |= intToBcd(day) << 8
	res |= intToBcd(int(month)) << 16
	res |= intToBcd(year-2000) << 24

	return res
}

func encodeTime(date *time.Time) uint32 {
	// HH:MM:SS
	var res uint32 = intToBcd(date.Second())
	res |= intToBcd(date.Minute()) << 8
	res |= intToBcd(date.Hour()) << 16

	return res
}

func intToBcd(val int) uint32 {
	if val > 99 {
		return 0
	}
	return uint32((val / 10 << 4) | (val % 10))
}
