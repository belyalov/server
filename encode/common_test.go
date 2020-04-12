package encode

import (
	"bytes"
	"hash/crc32"
	"testing"
	"time"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceMessageNoEncryption(t *testing.T) {
	dev := device.NewDevice(123)
	dev.SequenceSend = 1

	now := time.Now()
	encodedDate := encodeDate(&now)
	encodedTime := encodeTime(&now) & 0xffffff00
	payload, err := MakeReadyToSendDeviceMessage(dev,
		&openiot.JoinRequest{
			Name:         "test1",
			Manufacturer: "man1",
		},
	)
	require.NoError(t, err)
	buf := bytes.NewBuffer(payload)

	// Read message header
	hdr := &openiot.Header{}
	err = ReadSingleMessage(buf, hdr)
	require.NoError(t, err)
	assert.Equal(t, dev.ID, hdr.DeviceId)

	// Validate CRC
	crc := crc32.ChecksumIEEE(buf.Bytes())
	assert.Equal(t, crc, hdr.Crc)

	// Read the rest: MessageInfo, JoinReq
	msgInfo := &openiot.MessageInfo{}
	joinReq := &openiot.JoinRequest{}
	err = DecryptAndRead(buf, dev.EncryptionType, dev.Key(), msgInfo, joinReq)
	require.NoError(t, err)

	// Message Info (send sequence is correct)
	assert.Equal(t, dev.SequenceSend, msgInfo.Sequence)
	assert.Equal(t, encodedDate, msgInfo.Date)
	assert.Equal(t, encodedTime, msgInfo.Time&0xffffff00) // do not compare seconds

	// Actual "device" messages
	assert.Equal(t, "test1", joinReq.Name)
	assert.Equal(t, "man1", joinReq.Manufacturer)
}

func TestDeviceMessageAesECB(t *testing.T) {
	dev := device.NewDevice(333)
	dev.SequenceSend = 10
	// Setup encryption for device
	dev.SetKey([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	dev.EncryptionType = openiot.EncryptionType_AES_ECB

	payload, err := MakeReadyToSendDeviceMessage(dev,
		&openiot.JoinRequest{
			Name:         "test2",
			Manufacturer: "man22",
		},
	)
	require.NoError(t, err)
	buf := bytes.NewBuffer(payload)

	// Read message header
	hdr := &openiot.Header{}
	err = ReadSingleMessage(buf, hdr)
	require.NoError(t, err)
	assert.Equal(t, dev.ID, hdr.DeviceId)

	// Validate CRC
	crc := crc32.ChecksumIEEE(buf.Bytes())
	assert.Equal(t, crc, hdr.Crc)

	// Read the rest: MessageInfo, JoinReq, JoinResp
	msgInfo := &openiot.MessageInfo{}
	joinReq := &openiot.JoinRequest{}
	err = DecryptAndRead(buf, dev.EncryptionType, dev.Key(), msgInfo, joinReq)
	require.NoError(t, err)

	// Message Info (send sequence is correct)
	assert.Equal(t, dev.SequenceSend, msgInfo.Sequence)

	// Actual "device" messages
	assert.Equal(t, "test2", joinReq.Name)
	assert.Equal(t, "man22", joinReq.Manufacturer)
}

func TestIntToBcd(t *testing.T) {
	runs := map[int]uint32{
		0:  0x0,
		1:  0x1,
		9:  0x9,
		10: 0x10,
		16: 0x16,
		20: 0x20,
		99: 0x99,
	}

	for value, expected := range runs {
		assert.Equal(t, expected, intToBcd(value))
	}
}

func TestEncodeDate(t *testing.T) {
	// YY-MM-DD-WD
	// Mid week
	{
		date := time.Date(2011, time.December, 13, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, uint32(0x11121302), encodeDate(&date))
	}
	// Special case for Sunday
	{
		date := time.Date(2011, time.December, 25, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, uint32(0x11122507), encodeDate(&date))
	}
}

func TestEncodeTime(t *testing.T) {
	// HH:MM:SS
	date := time.Date(2011, time.December, 13, 22, 23, 24, 0, time.UTC)
	assert.Equal(t, uint32(0x222324), encodeTime(&date))
}
