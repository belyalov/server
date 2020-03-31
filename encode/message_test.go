package encode

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadWriteSingleMessage(t *testing.T) {
	original := &openiot.MessageInfo{
		Sequence: 3333,
	}

	// Write message
	var buffer bytes.Buffer
	require.NoError(t, WriteSingleMessage(&buffer, original))

	// Read it
	result := &openiot.MessageInfo{}
	require.NoError(t, ReadSingleMessage(&buffer, result))
	result.XXX_sizecache = original.XXX_sizecache

	// Validate
	require.Equal(t, original, result)
}

func TestWriteSingleMessage(t *testing.T) {
	msg := &openiot.Header{
		DeviceId: 10000,
		Crc:      1232312,
	}

	// Serialize it using "delimited" approach
	var result bytes.Buffer
	require.NoError(t, WriteSingleMessage(&result, msg))

	// Ensure that result is valid
	var expected bytes.Buffer

	serialized, err := proto.Marshal(msg)
	require.NoError(t, err)

	err = writeMessageLen(&expected, len(serialized))
	assert.NoError(t, err)
	expected.Write(serialized)

	require.Equal(t, expected.Bytes(), result.Bytes())
}

func TestReadSingleMessage(t *testing.T) {
	original := &openiot.MessageInfo{
		Sequence: 11111,
	}

	var buffer bytes.Buffer

	// Serialize it manually
	serialized, err := proto.Marshal(original)
	require.NoError(t, err)
	err = writeMessageLen(&buffer, len(serialized))
	assert.NoError(t, err)
	buffer.Write(serialized)

	// Validate
	result := &openiot.MessageInfo{}
	require.NoError(t, ReadSingleMessage(&buffer, result))
	result.XXX_sizecache = original.XXX_sizecache
	require.Equal(t, original, result)
}

func TestReadSingleMessageNegative(t *testing.T) {
	// Empty buffer
	buffer := &bytes.Buffer{}
	err := ReadSingleMessage(buffer, nil)
	assert.EqualError(t, err, "EOF")

	// Malformed varInt encoding
	buffer = bytes.NewBuffer([]byte{0x80})
	err = ReadSingleMessage(buffer, nil)
	assert.EqualError(t, err, "EOF")

	// Invalid message len
	buffer = bytes.NewBuffer([]byte{0x10, 0})
	err = ReadSingleMessage(buffer, nil)
	assert.EqualError(t, err, "Invalid message length: 16, max 1")

	// Invalid message len (too big, 64bit)
	buffer = bytes.NewBuffer([]byte{177, 237, 128, 130, 62, 177, 249})
	err = ReadSingleMessage(buffer, nil)
	assert.EqualError(t, err, "Invalid message length: 16647206577, max 2")
}

func TestAddPadding(t *testing.T) {
	var buf bytes.Buffer

	//  Empty buffer
	assert.NoError(t, addPadding(&buf))
	assert.Equal(t, 0, buf.Len())

	// Single byte
	buf.WriteRune('1')
	assert.NoError(t, addPadding(&buf))
	assert.Equal(t, 16, buf.Len())

	//  Buffer exactly 16 bytes (one block)
	assert.NoError(t, addPadding(&buf))
	assert.Equal(t, 16, buf.Len())

	// Buffer is 31 byte (1 byte of padding needed)
	tmp := make([]byte, 15)
	buf.Write(tmp)
	assert.NoError(t, addPadding(&buf))
	assert.Equal(t, 32, buf.Len())
}
