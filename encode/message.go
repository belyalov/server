package encode

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/golang/protobuf/proto"
)

// EncryptAndWriteMessages serializes all messages using "delimited"
// approach, then encodes result with AES using key and IV.
func EncryptAndWriteMessages(buffer *bytes.Buffer, key, iv []byte, msgs ...proto.Message) error {
	// Serialize all messages into continuos buffer
	var serializedBuf bytes.Buffer
	for _, msg := range msgs {
		if err := WriteSingleMessage(&serializedBuf, msg); err != nil {
			return err
		}
	}

	// AES operates only with blocks aligned to aes.BlockSize
	if err := addPadding(&serializedBuf); err != nil {
		return err
	}

	// Encrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	encrypted := make([]byte, serializedBuf.Len())
	encryptor := cipher.NewCBCEncrypter(block, iv)
	encryptor.CryptBlocks(encrypted, serializedBuf.Bytes())

	_, err = buffer.Write(encrypted)

	return err
}

// DecryptAndReadMessages decrypts buffer using key and IV, then deserializes all messages
// using "delimited" approach.
func DecryptAndReadMessages(buffer *bytes.Buffer, key, iv []byte, msgs ...proto.Message) error {
	// AES encrypted message must be aligned to AES block size
	if buffer.Len()%aes.BlockSize != 0 {
		return fmt.Errorf("Buffer is not aligned to AES block size")
	}

	// Decode buffer
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	decryptor := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, buffer.Len())
	decryptor.CryptBlocks(decrypted, buffer.Bytes())

	// Deserialize messages
	tmpBuf := bytes.NewBuffer(decrypted)
	for _, msg := range msgs {
		if err := ReadSingleMessage(tmpBuf, msg); err != nil {
			return err
		}
	}

	return nil
}

// ReadSingleMessage reads and de-serializes protobuf
// from buffer previously encoded using "delimited" approach,
// i.e. message length followed by message payload
func ReadSingleMessage(buffer *bytes.Buffer, msg proto.Message) error {
	messageLen, err := readMessageLen(buffer)
	if err != nil {
		return err
	}

	// Check boundaries
	if messageLen > buffer.Len() {
		return fmt.Errorf("Invalid message length: %d, max %d", messageLen, buffer.Len())
	}

	// De-serialize
	return proto.Unmarshal(buffer.Next(messageLen), msg)
}

// WriteSingleMessage serializes single protobuf
// using "delimited" approach into buffer.
func WriteSingleMessage(buffer *bytes.Buffer, msg proto.Message) error {
	// Serialize message
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	// Write message len
	if err := writeMessageLen(buffer, len(payload)); err != nil {
		return err
	}

	// Write serialized message
	_, err = buffer.Write(payload)

	return err
}

func writeMessageLen(buffer *bytes.Buffer, messageLen int) error {
	tmpBuf := make([]byte, 8)
	written := binary.PutUvarint(tmpBuf, uint64(messageLen))
	_, err := buffer.Write(tmpBuf[:written])

	return err
}

func readMessageLen(buffer *bytes.Buffer) (int, error) {
	msgLen, err := binary.ReadUvarint(buffer)

	return int(msgLen), err
}

// addPadding checks / adds random bytes padding
// to make buffer aligned with aes.BlockSize
func addPadding(buf *bytes.Buffer) error {
	if buf.Len()%aes.BlockSize == 0 {
		// Perfectly aligned to AES block size
		return nil
	}

	padding := make([]byte, aes.BlockSize-buf.Len()%aes.BlockSize)
	if _, err := rand.Read(padding); err != nil {
		return err
	}

	_, err := buf.Write(padding)

	return err
}
