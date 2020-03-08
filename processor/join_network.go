package processor

import (
	"bytes"
	"crypto/aes"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang/groupcache/lru"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/encode"
	"github.com/open-iot-devices/server/transport"
)

type keyExchangeItem struct {
	key            []byte
	encryptionType openiot.EncryptionType
}

var flagServerName = flag.String("server.name", "Open IoT Server", "Name of this server")

// Temporary map of devices where SystemJoinResponse was sent
var keyExchangeCache = lru.New(128)

func processKeyExchangeRequest(
	hdr *openiot.Header, buf *bytes.Buffer, transport transport.Transport) error {

	// Deserialize KeyExchangerequest
	request := &openiot.KeyExchangeRequest{}
	if err := encode.ReadSingleMessage(buf, request); err != nil {
		return err
	}
	if len(request.DhA) != aes.BlockSize {
		return fmt.Errorf("Invalid DhA len, %d", len(request.DhA))
	}
	if _, ok := openiot.EncryptionType_name[int32(request.EncryptionType)]; !ok {
		return fmt.Errorf("Unknown encoding %v", request.EncryptionType)
	}
	// Generate controller's part of Diffie-Hellman key exchange
	// and keep it until JoinRequest arrives
	private, public := generateDiffieHellman(request.DhG, request.DhP)
	entry := &keyExchangeItem{
		key:            calculateDiffieHellmanKey(request.DhP, request.DhA, private),
		encryptionType: request.EncryptionType,
	}
	keyExchangeCache.Add(hdr.DeviceId, entry)

	// Send KeyExchangeResponse: always un-encrypted
	response := &openiot.KeyExchangeResponse{
		DhB: public,
	}
	var sendBuf bytes.Buffer
	if err := encode.WritePlain(&sendBuf, hdr, response); err != nil {
		return err
	}

	return transport.Send(sendBuf.Bytes())
}

func processJoinRequest(
	hdr *openiot.Header, buf *bytes.Buffer, transport transport.Transport) error {

	var err error
	joinRequest := &openiot.JoinRequest{}

	// JoinRequest maybe encrypted or not:
	// - When encrypted - device must complete KeyExchange before
	// - Otherwise it is considered as possible un-encrypted join request
	keyInfo, ok := keyExchangeCache.Get(hdr.DeviceId)
	if ok {
		entry := keyInfo.(*keyExchangeItem)
		err = encode.DecryptAndRead(buf, entry.encryptionType, entry.key, joinRequest)
	} else {
		err = encode.ReadPlain(buf, joinRequest)
	}
	if err != nil {
		// If decrypt / de-serialize failed
		return err
	}

	// Add / Update device into registry
	dev := device.FindDeviceByID(hdr.DeviceId)
	if dev == nil {
		dev = device.NewDevice(hdr.DeviceId)
	}
	dev.Name = joinRequest.Name
	dev.Manufacturer = joinRequest.Manufacturer
	dev.ProductURL = joinRequest.ProductUrl
	dev.ProtobufURL = joinRequest.ProtobufUrl
	// Ignore "Device Already Exists" error
	_ = device.AddDevice(dev)

	// Send response
	response := &openiot.JoinResponse{
		Name:      *flagServerName,
		Timestamp: time.Now().Unix(),
	}
	var sendBuf bytes.Buffer
	if err := encode.WritePlain(&sendBuf, hdr, response); err != nil {
		return err
	}

	return transport.Send(sendBuf.Bytes())
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
