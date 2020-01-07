package encoding

import (
// "bytes"
// "crypto/aes"
// "crypto/cipher"
// "crypto/rand"
// "errors"
)

// // Load your secret key from a safe place and reuse it across multiple
// // NewCipher calls. (Obviously don't use this example key for anything
// // real.) If you want to convert a passphrase to a key, use a suitable
// // package like bcrypt or scrypt.
// key, _ := hex.DecodeString("6368616e676520746869732070617373")
// plaintext := []byte("some plaintext")

// block, err := aes.NewCipher(key)
// if err != nil {
// 	panic(err)
// }

// // The IV needs to be unique, but not secure. Therefore it's common to
// // include it at the beginning of the ciphertext.
// ciphertext := make([]byte, aes.BlockSize+len(plaintext))
// iv := ciphertext[:aes.BlockSize]
// if _, err := io.ReadFull(rand.Reader, iv); err != nil {
// 	panic(err)
// }

// stream := cipher.NewOFB(block, iv)
// stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

// // It's important to remember that ciphertexts must be authenticated
// // (i.e. by using crypto/hmac) as well as being encrypted in order to
// // be secure.

// // OFB mode is the same for both encryption and decryption, so we can
// // also decrypt that ciphertext with NewOFB.

// plaintext2 := make([]byte, len(plaintext))
// stream = cipher.NewOFB(block, iv)
// stream.XORKeyStream(plaintext2, ciphertext[aes.BlockSize:])

// fmt.Printf("%s\n", plaintext2)

// func AesEncryptCBC(buffer *bytes.Buffer, key, payload []byte) error {
// 	// Init cipher
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return err
// 	}
// 	// Allocate buffer for AES IV + space for encrypted payload
// 	buffer := make([]byte, aes.BlockSize+alignPacketLength(len(payload)))
// 	_, err = rand.Read(buffer)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// AESencryptCBC encrypts packet with AES-CBC using
// random generated AES initialization vector
// func AESencryptCBC(key, packet []byte) ([]byte, error) {
// 	// LoRa limits messages to 256 bytes long
// 	// packetLen := len(packet)
// 	// if packetLen > 255 {
// 	// 	return nil, errors.New("Packet too long")
// 	// }
// 	// Allocate buffer for encryption:
// 	// - IV
// 	// - Aligned to aes.blocksize packet
// 	alignedLength := alignPacketLength(len(packet))
// 	buf := make([]byte, alignedLength+aes.BlockSize)
// 	// Initialize buffer with random values (first 16 bytes will be used as IV)
// 	_, err := rand.Read(buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// First byte carries actual payload length:
// 	// buf[0] = byte(packetLen)
// 	// Copy packet into aes.blocksize aligned buffer
// 	alignedPacket := make([]byte, alignedLength)
// 	copy(alignedPacket, packet)
// 	// Encrypt using AES CBC
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	encryptor := cipher.NewCBCEncrypter(block, buf[:aes.BlockSize])
// 	encryptor.CryptBlocks(buf[aes.BlockSize:], alignedPacket)

// 	return buf, nil
// }

// // AESdecryptCBC decrypts message encrypted with AES-CBC.
// // Message must be prepended with AES-IV (initialization vector)
// // AESBlock size long (usually 16 bytes)
// func AESdecryptCBC(key, packet []byte) ([]byte, error) {
// 	packetLen := len(packet)
// 	// Messages less that 2 AES block size are invalid (IV + 1 block)
// 	if packetLen < aes.BlockSize*2 {
// 		return nil, errors.New("Packet too short")
// 	}
// 	// AES encrypted message must be multiple of AES block size
// 	if packetLen%aes.BlockSize != 0 {
// 		return nil, errors.New("Invalid packet length")
// 	}
// 	// First 16 bytes (AES block size) is IV (AES Initial Value)
// 	// Also, first byte carries actual payload length
// 	iv := packet[:aes.BlockSize]
// 	payloadLen := int(iv[0])
// 	// Ensure that payload length is correct: less that packet size - IV
// 	if payloadLen > packetLen-aes.BlockSize {
// 		return nil, errors.New("Invalid packet payload size")
// 	}
// 	// Decrypt using AES CBC
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	decryptor := cipher.NewCBCDecrypter(block, iv)
// 	decryptedPacket := make([]byte, packetLen-aes.BlockSize)
// 	decryptor.CryptBlocks(decryptedPacket, packet[aes.BlockSize:])

// 	return decryptedPacket[:payloadLen], nil
// }

// func alignPacketLength(packetLen int) int {
// 	if packetLen%aes.BlockSize != 0 {
// 		return (packetLen/aes.BlockSize + 1) * aes.BlockSize
// 	}
// 	return packetLen
// }
