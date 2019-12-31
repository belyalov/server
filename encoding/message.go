package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/golang/protobuf/proto"
)

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
func SerializeSingleMessage(buffer *bytes.Buffer, msg proto.Message) error {
	// Serialize message
	payload, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	// Write message length
	tmpBuf := make([]byte, 8)
	lenSize := binary.PutUvarint(tmpBuf, uint64(len(payload)))
	_, err = buffer.Write(tmpBuf[:lenSize])
	if err != nil {
		return err
	}
	// Write serialized message
	_, err = buffer.Write(payload)
	return err
}
