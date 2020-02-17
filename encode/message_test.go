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
	original := &openiot.Message{
		Crc:         12222,
		MessageType: 1,
		Sequence:    3333,
	}

	// Write message
	var buffer bytes.Buffer
	require.NoError(t, WriteSingleMessage(&buffer, original))

	// Read it
	result := &openiot.Message{}
	require.NoError(t, ReadSingleMessage(&buffer, result))
	result.XXX_sizecache = original.XXX_sizecache

	// Validate
	require.Equal(t, original, result)
}

func TestWriteSingleMessage(t *testing.T) {
	msg := &openiot.Header{
		DeviceId: 10000,
		Encryption: &openiot.Header_Plain{
			Plain: true,
		},
	}

	// Serialize it using "delimited" approach
	var result bytes.Buffer
	require.NoError(t, WriteSingleMessage(&result, msg))

	// Ensure that result is valid
	var expected bytes.Buffer

	serialized, err := proto.Marshal(msg)
	require.NoError(t, err)

	expected.WriteByte(byte(len(serialized)))
	expected.Write(serialized)

	require.Equal(t, expected.Bytes(), result.Bytes())
}

func TestReadSingleMessage(t *testing.T) {
	original := &openiot.Message{
		Crc:         12,
		MessageType: 11111,
		Sequence:    11111,
	}

	var buffer bytes.Buffer

	// Serialize it manually
	serialized, err := proto.Marshal(original)
	require.NoError(t, err)
	buffer.WriteByte(byte(len(serialized)))
	buffer.Write(serialized)

	// Validate
	result := &openiot.Message{}
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
