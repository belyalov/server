package encode

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/golang/protobuf/proto"
)

// ReadSingleMessage reads and de-serializes protobuf
// from buffer previously encoded using "delimited" approach,
// i.e. message length followed by message payload
func ReadSingleMessage(buffer *bytes.Buffer, msg proto.Message) error {
	msgLen, err := binary.ReadUvarint(buffer)
	if err != nil {
		return err
	}

	// Check boundaries
	if msgLen > uint64(buffer.Len()) {
		return fmt.Errorf("Invalid message length: %d, max %d", msgLen, buffer.Len())
	}

	// De-serialize
	return proto.Unmarshal(buffer.Next(int(msgLen)), msg)
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
