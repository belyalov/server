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

	if dev := device.FindDeviceByID(hdr.DeviceId); dev != nil {
		return fmt.Errorf("Key Exchange request for already registered device 0x%x", hdr.DeviceId)
	}

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

	// JoinRequest maybe encrypted or not:
	// - When encrypted - device must complete KeyExchange before
	// - It maybe duplicate JoinRequest, in this case take device's encryption params
	encParams := &keyExchangeItem{}
	if dev := device.FindDeviceByID(hdr.DeviceId); dev != nil {
		encParams.key = dev.Key()
		encParams.encryptionType = dev.EncryptionType
	} else if keyInfo, ok := keyExchangeCache.Get(hdr.DeviceId); ok {
		encParams = keyInfo.(*keyExchangeItem)
	}

	// Read/Decode JoinRequest
	joinRequest := &openiot.JoinRequest{}
	if err := encode.DecryptAndRead(buf, encParams.encryptionType, encParams.key, joinRequest); err != nil {
		return err
	}

	// Add device into registry / update info if already present
	dev := device.FindDeviceByID(hdr.DeviceId)
	if dev == nil {
		dev = device.NewDevice(hdr.DeviceId)
	}
	if joinRequest.DefaultHandler != "" {
		dev.AddHandler(joinRequest.DefaultHandler)
	}
	dev.SetTransport(transport)
	dev.Name = joinRequest.Name
	dev.Manufacturer = joinRequest.Manufacturer
	dev.ProductURL = joinRequest.ProductUrl
	dev.ProtobufURL = joinRequest.ProtobufUrl
	device.AddDevice(dev) // Ignore "Device Already Exists" error

	// Send response
	joinResp := &openiot.JoinResponse{
		Name:      *flagServerName,
		Timestamp: time.Now().Unix(),
	}
	payload, err := encode.MakeReadyToSendMessage(hdr, encParams.encryptionType, encParams.key, joinResp)
	if err != nil {
		return err
	}
	return transport.Send(payload)
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
