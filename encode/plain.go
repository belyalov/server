package encode

import (
	"bytes"

	"github.com/golang/protobuf/proto"
)

// WritePlain serializes all messages using "delimited"
// it does not performs any type of encryption, just writes messages into buffer
func WritePlain(buffer *bytes.Buffer, msgs ...proto.Message) error {
	// Serialize all messages into continuos buffer
	for _, msg := range msgs {
		if err := WriteSingleMessage(buffer, msg); err != nil {
			return err
		}
	}

	return nil
}

// ReadPlain de-serializes all messages using "delimited" approach.
// It does not perform any decryption
func ReadPlain(buffer *bytes.Buffer, msgs ...proto.Message) error {
	// Deserialize messages
	for _, msg := range msgs {
		if err := ReadSingleMessage(buffer, msg); err != nil {
			return err
		}
	}

	return nil
}
