package processor

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/encode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// Negative: malformed payload
	transport := &mockTransport{}
	buf := bytes.NewBufferString("fsdjfhsdfkds")
	err := processKeyExchangeRequest(hdr, buf, transport)
	assert.Error(t, err)
	assert.True(t, transport.Empty())
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

func TestJoinNoEncryption(t *testing.T) {
	defer device.DeleteAllDevices()

	// When no encryption used device may simply send
	// JoinRequest and get joined into network
	hdr := &openiot.Header{
		DeviceId:    112233,
		JoinRequest: true,
	}
	joinReq := &openiot.JoinRequest{
		Name:         "test1",
		Manufacturer: "man1",
		ProductUrl:   "url1",
		ProtobufUrl:  "proto1",
	}
	var buf bytes.Buffer
	err := encode.WritePlain(&buf, hdr, joinReq)
	require.NoError(t, err)
	// "Send" Join Request
	transport := &mockTransport{}
	msg := &Message{
		Source:  transport,
		Payload: buf.Bytes(),
	}
	err = ProcessMessage(msg)
	require.NoError(t, err)

	// Validate Join Response
	hdrResp := &openiot.Header{}
	joinResp := &openiot.JoinResponse{}
	err = encode.ReadPlain(transport.LastMessage(), hdrResp, joinResp)
	require.NoError(t, err)
	assert.Equal(t, *flagServerName, joinResp.Name)
	// Ensure that device has been added / all parameters have been picked up
	dev := device.FindDeviceByID(hdr.DeviceId)
	require.NotNil(t, dev)
	assert.Equal(t, joinReq.Name, dev.Name)
	assert.Equal(t, joinReq.Manufacturer, dev.Manufacturer)
	assert.Equal(t, joinReq.ProductUrl, dev.ProductURL)
	assert.Equal(t, joinReq.ProtobufUrl, dev.ProtobufURL)

	// Finally ensure that duplicates of JoinRequest do not create new device
	// always respond with JoinResponse
	dev.SequenceSend = 10
	err = ProcessMessage(msg)
	require.NoError(t, err)
	dev = device.FindDeviceByID(hdr.DeviceId)
	assert.Equal(t, uint32(10), dev.SequenceSend)
	assert.Equal(t, 2, len(transport.history))
}

func TestJoinWithEncryption(t *testing.T) {
	defer device.DeleteAllDevices()
	var buf bytes.Buffer

	transport := &mockTransport{}
	msg := &Message{
		Source: transport,
	}

	// Perform Key Exchange for AES ECB
	privateA, publicA := generateDiffieHellman(dhG, dhP)
	hdr := &openiot.Header{
		DeviceId:    999,
		KeyExchange: true,
	}
	keyReq := &openiot.KeyExchangeRequest{
		DhG:            dhG,
		DhP:            dhP,
		DhA:            publicA,
		EncryptionType: openiot.EncryptionType_AES_ECB,
	}
	encode.WritePlain(&buf, hdr, keyReq)
	// Run Process Message
	msg.Payload = buf.Bytes()
	err := ProcessMessage(msg)
	require.NoError(t, err)
	hdrResp := &openiot.Header{}
	keyResp := &openiot.KeyExchangeResponse{}
	err = encode.ReadPlain(transport.LastMessage(), hdrResp, keyResp)
	require.NoError(t, err)
	// Calculate encryption key
	key := calculateDiffieHellmanKey(keyReq.DhP, keyResp.DhB, privateA)

	// Perform Join Request
	hdr = &openiot.Header{
		DeviceId:    999,
		JoinRequest: true,
	}
	joinReq := &openiot.JoinRequest{
		Name:         "test1",
		Manufacturer: "man1",
		ProductUrl:   "url1",
		ProtobufUrl:  "proto1",
	}
	buf.Reset()
	// Encrypt JoinRequest
	var encBuf bytes.Buffer
	err = encode.WriteAndEncrypt(&encBuf, keyReq.EncryptionType, key, joinReq)
	require.NoError(t, err)
	// Write all into one buffer
	err = encode.WriteSingleMessage(&buf, hdr)
	require.NoError(t, err)
	_, err = encBuf.WriteTo(&buf)
	require.NoError(t, err)
	// Run Process Message
	msg.Payload = buf.Bytes()
	err = ProcessMessage(msg)
	require.NoError(t, err)

	// Ensure that device has been added / all parameters have been picked up
	dev := device.FindDeviceByID(hdr.DeviceId)
	require.NotNil(t, dev)
	assert.Equal(t, joinReq.Name, dev.Name)
	assert.Equal(t, joinReq.Manufacturer, dev.Manufacturer)
	assert.Equal(t, joinReq.ProductUrl, dev.ProductURL)
	assert.Equal(t, joinReq.ProtobufUrl, dev.ProtobufURL)
	// // Check that JoinResponse has been sent
	hdrResp = &openiot.Header{}
	joinResp := &openiot.JoinResponse{}
	err = encode.ReadPlain(transport.LastMessage(), hdrResp, joinResp)
	require.NoError(t, err)
	assert.Equal(t, *flagServerName, joinResp.Name)

	// Ensure that duplicates of JoinRequest do not creates new device
	dev.SequenceSend = 10
	err = ProcessMessage(msg)
	require.NoError(t, err)
	dev = device.FindDeviceByID(hdr.DeviceId)
	assert.Equal(t, uint32(10), dev.SequenceSend)
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
