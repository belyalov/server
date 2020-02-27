package processor

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/encode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dhG = 199
	dhP = 4001
)

func TestKeyExchange(t *testing.T) {
	keyExchangeCache.Clear()

	hdr := &openiot.Header{
		DeviceId: 123,
		Crc:      1212,
	}

	// Try it few times: every request should override previously saved key
	for i := 0; i < 5; i++ {
		// Generate Key Exchange request
		privateA, publicA := generateDiffieHellman(dhG, dhP)
		jreq := &openiot.KeyExchangeRequest{
			DhG: dhG,
			DhP: dhP,
			DhA: publicA,
		}
		var buf bytes.Buffer
		encode.WriteSingleMessage(&buf, jreq)

		// Run it
		jresp, err := processKeyExchangeRequest(hdr, &buf)
		require.NoError(t, err)
		require.NotNil(t, jresp)

		// Calculate our key
		key := calculateDiffieHellmanKey(jreq.DhP, jresp.(*openiot.KeyExchangeResponse).DhB, privateA)

		// Ensure that the same key is pending in the list
		cached, ok := keyExchangeCache.Get(hdr.DeviceId)
		assert.True(t, ok)
		assert.Equal(t, key, cached)
	}

	// Finally only one key should be in cache
	assert.Equal(t, 1, keyExchangeCache.Len())
}

func TestGenerateDiffieHellman(t *testing.T) {
	rand.Seed(1)
	private, public := generateDiffieHellman(199, 4001)

	assert.Equal(t,
		[]uint32{1090, 1054, 2958, 2422, 2818, 780, 1650, 1368, 3728, 3912, 2445, 1375, 3909, 4067, 688, 2613},
		private)
	assert.Equal(t,
		[]uint32{121, 3672, 1450, 3760, 394, 1031, 2886, 3418, 625, 1154, 3231, 2055, 755, 1524, 3610, 3203},
		public)
}
