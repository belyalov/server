package encoding

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	secure_rand "crypto/rand"
	"encoding/binary"
	"fmt"
	insecure_rand "math/rand"

	"github.com/golang/protobuf/proto"
	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/device"
)

// SerializeAndEncodeMessage makes ready to be send bytes
// by serializing and (optional) encoding of given protobuf.
// It also takes care of preparing OpenIoT message header
func SerializeAndEncodeMessage(dev device.Device, message proto.Message) ([]byte, error) {
	var serialized bytes.Buffer

	hdr := &openiot.HeaderMessage{
		DeviceId: dev.GetDeviceID(),
	}

	// Special handling for JoinRequest and JoinResponse system messages
	// since they should never be encrypted (key exchange).
	// so handling them specially
	if _, ok := message.(*openiot.SystemMessage); ok {
		hdr.SystemMessage = true
		hdr.Encryption = &openiot.HeaderMessage_Plain{Plain: true}
		// Write 2 messages: Header + System
		_, err := SerializeSingleMessage(&serialized, hdr)
		if err != nil {
			return nil, err
		}
		_, err = SerializeSingleMessage(&serialized, message)
		return nil, err
	}

	// Serialize message
	if _, err := SerializeSingleMessage(&serialized, message); err != nil {
		return nil, err
	}

	// Align to AES block size
	if serialized.Len()%aes.BlockSize != 0 {
		for i := 0; i < serialized.Len()%aes.BlockSize; i++ {
			if err := serialized.WriteByte(byte(insecure_rand.Intn(255))); err != nil {
				return nil, err
			}
		}
	}

	// Calculate AES IV (just random values)
	aesIv := make([]byte, aes.BlockSize)
	if _, err := secure_rand.Read(aesIv); err != nil {
		return nil, err
	}

	// Encrypt message
	block, err := aes.NewCipher(dev.GetEncryptionKey())
	if err != nil {
		return nil, err
	}
	encrypted := make([]byte, serialized.Len())
	encryptor := cipher.NewCBCEncrypter(block, aesIv)
	encryptor.CryptBlocks(encrypted, serialized.Bytes())

	// Compose entire message
	var payload bytes.Buffer
	hdr.Encryption = &openiot.HeaderMessage_AesIv{
		AesIv: aesIv,
	}
	// Write message header
	if _, err = SerializeSingleMessage(&payload, hdr); err != nil {
		return nil, err
	}
	// Write encrypted (and previously serialized) message
	payload.Write(encrypted)

	return payload.Bytes(), nil
}

// DeserializeSingleMessage reads and de-serializes protobuf
// from buffer previously encoded using "delimited" approach,
// i.e. message length followed by message payload
func DeserializeSingleMessage(buffer *bytes.Buffer, msg proto.Message) error {
	// Read message length (encoded using "delimited" approach)
	msgLen, err := binary.ReadUvarint(buffer)
	if err != nil {
		return fmt.Errorf("Unable to read message length: %v", err)
	}

	// Ensure that it fits remain buffer
	if int(msgLen) > buffer.Len() {
		return fmt.Errorf("Invalid message length: %d, max %d", msgLen, buffer.Len())
	}

	// De-Serialize
	return proto.Unmarshal(buffer.Next(int(msgLen)), msg)
}

// SerializeSingleMessage serializes single protobuf
// using "delimited" approach into buffer.
// Returns amount of written bytes (including message length) or error, if any
func SerializeSingleMessage(buffer *bytes.Buffer, msg proto.Message) (int, error) {
	// Serialize message
	payload, err := proto.Marshal(msg)
	if err != nil {
		return 0, err
	}

	// Write message length
	tmpBuf := make([]byte, 8)
	lenSize := binary.PutUvarint(tmpBuf, uint64(len(payload)))
	_, err = buffer.Write(tmpBuf[:lenSize])
	if err != nil {
		return 0, err
	}

	// Write serialized message
	_, err = buffer.Write(payload)

	return len(payload) + lenSize, err
}
