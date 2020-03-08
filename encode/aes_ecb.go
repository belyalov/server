package encode

import (
	"bytes"
	"crypto/aes"
	"fmt"

	"github.com/golang/protobuf/proto"
)

// WriteAndEncryptECB serializes all messages using "delimited"
// approach, then encodes result with AES ECB using provided key.
func WriteAndEncryptECB(buffer *bytes.Buffer, key []byte, msgs ...proto.Message) error {
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
	for i := 0; serializedBuf.Len() > 0; i += aes.BlockSize {
		block.Encrypt(encrypted[i:], serializedBuf.Next(aes.BlockSize))
	}

	_, err = buffer.Write(encrypted)
	return err
}

// DecryptAndReadECB decrypts buffer using AES-ECB with provided key,
// then de-serializes all messages using "delimited" approach.
func DecryptAndReadECB(buffer *bytes.Buffer, key []byte, msgs ...proto.Message) error {
	// AES encrypted message must be aligned to AES block size
	if buffer.Len()%aes.BlockSize != 0 {
		return fmt.Errorf("Buffer is not aligned to AES block size")
	}
	// Decode buffer
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	decrypted := make([]byte, buffer.Len())
	for i := 0; buffer.Len() > 0; i += aes.BlockSize {
		block.Decrypt(decrypted[i:], buffer.Next(aes.BlockSize))
	}
	// Deserialize messages
	tmpBuf := bytes.NewBuffer(decrypted)
	for _, msg := range msgs {
		if err := ReadSingleMessage(tmpBuf, msg); err != nil {
			return err
		}
	}

	return nil
}
