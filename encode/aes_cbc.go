package encode

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/golang/protobuf/proto"
)

// WriteAndEncryptCBC serializes all messages using "delimited"
// approach, then encodes result with AES-CBC using provided key and iv.
func WriteAndEncryptCBC(buffer *bytes.Buffer, key, iv []byte, msgs ...proto.Message) error {
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

// DecryptAndReadCBC decrypts buffer using AES-CBC with provided key and IV,
// then deserializes all messages using "delimited" approach.
func DecryptAndReadCBC(buffer *bytes.Buffer, key, iv []byte, msgs ...proto.Message) error {
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
