package processor

import (
	"bytes"
	"crypto/aes"
	"fmt"
	"math/rand"

	"github.com/golang/protobuf/proto"
	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/encode"
)

// Temporary map of devices where SystemJoinResponse has sent
var pendingDevices = map[uint64][]byte{}

func handleJoinNetwork(hdr *openiot.Header, buf *bytes.Buffer) (proto.Message, error) {
	// Deserialize / Verify JoinRequest
	jreq := &openiot.SystemJoinRequest{}
	if err := encode.ReadSingleMessage(buf, jreq); err != nil {
		return nil, err
	}
	if len(jreq.DhA) != aes.BlockSize {
		return nil, fmt.Errorf("Invalid DhA len, %d", len(jreq.DhA))
	}
	// Generate server part of Diffie-Hellman key exchange and save it
	private, public := generateDiffieHellman(jreq.DhG, jreq.DhP)
	pendingDevices[hdr.DeviceId] = calculateDiffieHellmanKey(jreq.DhP, jreq.DhA, private)

	return &openiot.SystemJoinResponse{
		DhB: public,
	}, nil
}

// Diffie-Hellman implementation //

func generateDiffieHellman(dhG, dhP uint64) ([]uint32, []uint32) {
	public := make([]uint32, aes.BlockSize)
	private := make([]uint32, aes.BlockSize)
	for i := 0; i < aes.BlockSize; i++ {
		private[i] = rand.Uint32() % 4096
		public[i] = uint32(
			diffieHellmanPowMod(int(dhG), int(private[i]), int(dhP)),
		)
	}
	return private, public
}

func calculateDiffieHellmanKey(dhP uint64, public, private []uint32) []byte {
	key := make([]byte, len(private))
	for index := range private {
		key[index] = byte(diffieHellmanPowMod(
			int(public[index]),
			int(private[index]),
			int(dhP),
		))
	}
	return key
}

// it does math: g**x mod n
func diffieHellmanPowMod(g, x, p int) int {
	var r int
	var y int = 1

	for x > 0 {
		r = x % 2
		// Fast exponention
		if r == 1 {
			y = (y * g) % p
		}
		g = g * g % p
		x = x / 2
	}

	return y
}
