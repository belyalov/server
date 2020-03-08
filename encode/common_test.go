package encode

import (
	"bytes"
	"hash/crc32"
	"testing"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceMessageNoEncryption(t *testing.T) {
	dev := device.NewDevice(123)
	dev.SequenceSend = 1

	payload, err := MakeReadyToSendDeviceMessage(dev,
		&openiot.JoinRequest{
			Name:         "test1",
			Manufacturer: "man1",
		},
		&openiot.JoinResponse{
			Timestamp: 555,
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
	joinResp := &openiot.JoinResponse{}
	err = DecryptAndRead(buf, dev.EncryptionType, dev.Key(), msgInfo, joinReq, joinResp)
	require.NoError(t, err)

	// Message Info (send sequence is correct)
	assert.Equal(t, dev.SequenceSend, msgInfo.Sequence)

	// Actual "device" messages
	assert.Equal(t, "test1", joinReq.Name)
	assert.Equal(t, "man1", joinReq.Manufacturer)
	assert.Equal(t, int64(555), joinResp.Timestamp)
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
