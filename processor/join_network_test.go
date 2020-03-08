package processor

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/encode"
)

const (
	dhG = 199
	dhP = 4001
)

func TestKeyExchange(t *testing.T) {
	defer keyExchangeCache.Clear()

	hdr := &openiot.Header{
		DeviceId:    123,
		KeyExchange: true,
	}

	// Try it few times: every request should override previously saved key for device
	for i := 0; i < 3; i++ {
		transport := &mockTransport{}
		// Generate Key Exchange request
		privateA, publicA := generateDiffieHellman(dhG, dhP)
		keyReq := &openiot.KeyExchangeRequest{
			DhG: dhG,
			DhP: dhP,
			DhA: publicA,
		}
		var buf bytes.Buffer
		encode.WriteSingleMessage(&buf, keyReq)

		// Run / parse response
		err := processKeyExchangeRequest(hdr, &buf, transport)
		require.NoError(t, err)
		hdrResp := &openiot.Header{}
		keyResp := &openiot.KeyExchangeResponse{}
		err = encode.ReadPlain(transport.LastMessage(), hdrResp, keyResp)
		require.NoError(t, err)
		// Validate response header
		assert.Equal(t, hdr.DeviceId, hdrResp.DeviceId)
		assert.True(t, hdrResp.KeyExchange)
		assert.False(t, hdrResp.JoinRequest)

		// Validate key correctness
		key := calculateDiffieHellmanKey(keyReq.DhP, keyResp.DhB, privateA)
		// Ensure that the same key is pending in the list
		cached, ok := keyExchangeCache.Get(hdr.DeviceId)
		assert.True(t, ok)
		assert.Equal(t, key, cached.(*keyExchangeItem).key)
	}

	// Finally only one key should be in cache
	assert.Equal(t, 1, keyExchangeCache.Len())
}

func TestKeyExchangeDeviceAlreadyRegistered(t *testing.T) {
	defer device.DeleteAllDevices()

	// Register device and then try to perform key exchange
	dev := device.NewDevice(888)
	device.AddDevice(dev)

	hdr := &openiot.Header{
		DeviceId:    888,
		KeyExchange: true,
	}
	err := processKeyExchangeRequest(hdr, nil, nil)
	assert.EqualError(t, err, "Key Exchange request for already registered device 0x378")
	assert.Equal(t, 0, keyExchangeCache.Len())
}
func TestKeyExchangeNegative(t *testing.T) {
	var buf bytes.Buffer
	transport := &mockTransport{}

	// Invalid len of DhA
	keyReq := &openiot.KeyExchangeRequest{
		DhA: []uint32{1, 2, 3},
	}
	encode.WriteSingleMessage(&buf, keyReq)
	err := processKeyExchangeRequest(&openiot.Header{}, &buf, transport)
	assert.EqualError(t, err, "Invalid DhA len, 3")
	assert.True(t, transport.Empty())

	// Malformed payload of KeyExchange protobuf
	buf.WriteString("fsdfsfsdfsd")
	err = processKeyExchangeRequest(&openiot.Header{}, &buf, transport)
	assert.EqualError(t, err, "Invalid message length: 102, max 10")
	assert.True(t, transport.Empty())
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

func TestJoinNoEncryption(t *testing.T) {
	defer device.DeleteAllDevices()
	// When no encryption used device may simply send
	// JoinRequest and get joined into network
	joinReq := &openiot.JoinRequest{
		Name:         "test1",
		Manufacturer: "man1",
		ProductUrl:   "url1",
		ProtobufUrl:  "proto1",
	}
	joinResp, err := performJoinRequest(112233, openiot.EncryptionType_PLAIN, nil, joinReq)
	require.NoError(t, err)
	// Validate JoinResponse
	assert.Equal(t, *flagServerName, joinResp.Name)
	assert.GreaterOrEqual(t, joinResp.Timestamp, time.Now().Unix())
	// Ensure that device has been added / all parameters have been picked up
	dev := device.FindDeviceByID(112233)
	require.NotNil(t, dev)
	assert.Equal(t, joinReq.Name, dev.Name)
	assert.Equal(t, joinReq.Manufacturer, dev.Manufacturer)
	assert.Equal(t, joinReq.ProductUrl, dev.ProductURL)
	assert.Equal(t, joinReq.ProtobufUrl, dev.ProtobufURL)
}

func TestJoinNoEncryptionDeviceExists(t *testing.T) {
	defer device.DeleteAllDevices()

	// Add device to registry
	dev := &device.Device{
		ID:           555,
		Name:         "dummy",
		SequenceSend: 10,
	}
	err := device.AddDevice(dev)
	require.NoError(t, err)

	// Send JoinRequest with the same device id
	joinReq := &openiot.JoinRequest{
		Name:         "test555",
		Manufacturer: "man555",
		ProductUrl:   "url555",
		ProtobufUrl:  "proto555",
	}
	joinResp, err := performJoinRequest(555, dev.EncryptionType, dev.Key(), joinReq)
	require.NoError(t, err)
	// Validate JoinResponse
	assert.Equal(t, *flagServerName, joinResp.Name)
	assert.GreaterOrEqual(t, joinResp.Timestamp, time.Now().Unix())
	// Ensure that device info has updated (no new device created)
	dev = device.FindDeviceByID(555)
	require.NotNil(t, dev)
	assert.Equal(t, joinReq.Name, dev.Name)
	assert.Equal(t, joinReq.Manufacturer, dev.Manufacturer)
	assert.Equal(t, joinReq.ProductUrl, dev.ProductURL)
	assert.Equal(t, joinReq.ProtobufUrl, dev.ProtobufURL)
	assert.Equal(t, uint32(10), dev.SequenceSend)
}

func TestJoinWithEncryption(t *testing.T) {
	defer device.DeleteAllDevices()
	defer keyExchangeCache.Clear()

	// Perform key exchange since network join will be encrypted
	key, err := performKeyExchangeRequest(999, openiot.EncryptionType_AES_ECB)

	// Perform Join Request
	joinReq := &openiot.JoinRequest{
		Name:         "test99",
		Manufacturer: "man99",
		ProductUrl:   "url99",
		ProtobufUrl:  "proto99",
	}
	joinResp, err := performJoinRequest(999, openiot.EncryptionType_AES_ECB, key, joinReq)
	require.NoError(t, err)
	// Validate response
	assert.Equal(t, *flagServerName, joinResp.Name)

	// Ensure that device has been added / all parameters have been picked up
	dev := device.FindDeviceByID(999)
	require.NotNil(t, dev)
	assert.Equal(t, joinReq.Name, dev.Name)
	assert.Equal(t, joinReq.Manufacturer, dev.Manufacturer)
	assert.Equal(t, joinReq.ProductUrl, dev.ProductURL)
	assert.Equal(t, joinReq.ProtobufUrl, dev.ProtobufURL)

	// One more try with invalid key
	key[0] = 0
	joinResp, err = performJoinRequest(999, openiot.EncryptionType_AES_ECB, key, joinReq)
	require.Error(t, err)
}

func TestJoinWithEncryptionDeviceExists(t *testing.T) {
	defer device.DeleteAllDevices()

	// Add device to registry
	dev := &device.Device{
		ID:             321,
		Name:           "dummy",
		SequenceSend:   20,
		EncryptionType: openiot.EncryptionType_AES_ECB,
	}
	dev.SetKey([]byte{11, 22, 33, 44, 55, 66, 77, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	err := device.AddDevice(dev)
	require.NoError(t, err)

	// Send JoinRequest with the same device id
	joinReq := &openiot.JoinRequest{
		Name:         "test555",
		Manufacturer: "man555",
		ProductUrl:   "url555",
		ProtobufUrl:  "proto555",
	}
	joinResp, err := performJoinRequest(321, dev.EncryptionType, dev.Key(), joinReq)
	require.NoError(t, err)
	// Validate JoinResponse
	assert.Equal(t, *flagServerName, joinResp.Name)
	assert.GreaterOrEqual(t, joinResp.Timestamp, time.Now().Unix())
	// Ensure that device info has updated (no new device created)
	dev = device.FindDeviceByID(321)
	require.NotNil(t, dev)
	assert.Equal(t, joinReq.Name, dev.Name)
	assert.Equal(t, joinReq.Manufacturer, dev.Manufacturer)
	assert.Equal(t, joinReq.ProductUrl, dev.ProductURL)
	assert.Equal(t, joinReq.ProtobufUrl, dev.ProtobufURL)
	assert.Equal(t, uint32(20), dev.SequenceSend)
}

// helpers //

func performJoinRequest(
	id uint64, enc openiot.EncryptionType, key []byte, request *openiot.JoinRequest) (
	*openiot.JoinResponse, error) {

	// Prepare Message (hdr + request)
	hdr := &openiot.Header{
		DeviceId:    id,
		JoinRequest: true,
	}
	payload, err := encode.MakeReadyToSendMessage(hdr, enc, key, request)
	if err != nil {
		return nil, fmt.Errorf("MakeReadyToSendMessage failed: %v", err)
	}

	// Send request
	transport := &mockTransport{}
	msg := &Message{
		Source:  transport,
		Payload: payload,
	}
	if err := ProcessMessage(msg); err != nil {
		return nil, fmt.Errorf("ProcessMessage failed: %v", err)
	}

	// Extract response
	hdrResp := &openiot.Header{}
	joinResp := &openiot.JoinResponse{}
	respBuf := transport.LastMessage()
	if err := encode.ReadSingleMessage(respBuf, hdrResp); err != nil {
		return nil, fmt.Errorf("ReadSingleMessage failed: %v", err)
	}
	err = encode.DecryptAndRead(respBuf, enc, key, joinResp)

	return joinResp, err
}

func performKeyExchangeRequest(id uint64, enc openiot.EncryptionType) ([]byte, error) {
	// Generate Diffie Hellman numbers / KeyExchange Request
	privateA, publicA := generateDiffieHellman(dhG, dhP)
	keyReq := &openiot.KeyExchangeRequest{
		DhG:            dhG,
		DhP:            dhP,
		DhA:            publicA,
		EncryptionType: enc,
	}

	// Prepare Message (hdr + request)
	hdr := &openiot.Header{
		DeviceId:    id,
		KeyExchange: true,
	}
	payload, err := encode.MakeReadyToSendMessage(hdr, openiot.EncryptionType_PLAIN, nil, keyReq)
	if err != nil {
		return nil, fmt.Errorf("MakeReadyToSendMessage failed: %v", err)
	}

	// Send KeyExchange Request
	transport := &mockTransport{}
	msg := &Message{
		Source:  transport,
		Payload: payload,
	}
	if err := ProcessMessage(msg); err != nil {
		return nil, fmt.Errorf("ProcessMessage failed: %v", err)
	}

	// Extract response
	hdrResp := &openiot.Header{}
	keyResp := &openiot.KeyExchangeResponse{}
	respBuf := transport.LastMessage()
	if err := encode.ReadSingleMessage(respBuf, hdrResp); err != nil {
		return nil, fmt.Errorf("ReadSingleMessage failed: %v", err)
	}
	err = encode.DecryptAndRead(respBuf, openiot.EncryptionType_PLAIN, nil, keyResp)

	// Calculate encryption key
	return calculateDiffieHellmanKey(keyReq.DhP, keyResp.DhB, privateA), nil
}
