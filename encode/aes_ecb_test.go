package encode

import (
	"bytes"
	"testing"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodeECB(t *testing.T) {
	msg1 := &openiot.SystemJoinRequest{
		DhP: 10,
		DhG: 100000,
		DhA: []uint32{1, 2, 3},
	}
	msg2 := &openiot.SystemJoinResponse{
		DhB: []uint32{11, 22, 33},
	}

	// Encrypt messages
	var buf bytes.Buffer
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	assert.NoError(t, WriteAndEncryptECB(&buf, key, msg1, msg2))

	// Decrypt previously encrypted bytes
	res1 := &openiot.SystemJoinRequest{}
	res2 := &openiot.SystemJoinResponse{}
	assert.NoError(t, DecryptAndReadECB(&buf, key, res1, res2))

	// Ensure that decoded / deserialized messages match
	res1.XXX_sizecache = msg1.XXX_sizecache
	res2.XXX_sizecache = msg2.XXX_sizecache
	assert.Equal(t, msg1, res1)
}
