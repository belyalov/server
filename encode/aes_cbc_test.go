package encode

import (
	"bytes"
	"testing"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodeCBC(t *testing.T) {
	msg1 := &openiot.KeyExchangeRequest{
		DhP: 10,
		DhG: 100000,
		DhA: []uint32{1, 2, 3},
	}
	msg2 := &openiot.KeyExchangeResponse{
		DhB: []uint32{11, 22, 33},
	}

	// Encrypt messages
	var buf bytes.Buffer
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	iv := key
	assert.NoError(t, WriteAndEncryptCBC(&buf, key, iv, msg1, msg2))

	// Decrypt previously encrypted bytes
	res1 := &openiot.KeyExchangeRequest{}
	res2 := &openiot.KeyExchangeResponse{}
	assert.NoError(t, DecryptAndReadCBC(&buf, key, iv, res1, res2))

	// Ensure that decoded / deserialized messages match
	res1.XXX_sizecache = msg1.XXX_sizecache
	res2.XXX_sizecache = msg2.XXX_sizecache
	assert.Equal(t, msg1, res1)
}
