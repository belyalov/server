package encode

import (
	"bytes"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/open-iot-devices/protobufs/go/openiot"
)

// DecryptAndRead decrypts buffer using encoding from encType,
// then de-serializes all messages using "delimited" approach.
func DecryptAndRead(
	buffer *bytes.Buffer, encType openiot.EncryptionType, key []byte, msgs ...proto.Message) error {

	switch encType {
	case openiot.EncryptionType_PLAIN:
		return ReadPlain(buffer, msgs...)
	case openiot.EncryptionType_AES_ECB:
		return DecryptAndReadECB(buffer, key, msgs...)
	}

	return fmt.Errorf("Encoding %v is not supported", encType)
}

// WriteAndEncrypt serializes all messages using "delimited"
// approach, then encodes result with encoding from encType.
func WriteAndEncrypt(
	buffer *bytes.Buffer, encType openiot.EncryptionType, key []byte, msgs ...proto.Message) error {

	switch encType {
	case openiot.EncryptionType_PLAIN:
		return WritePlain(buffer, msgs...)
	case openiot.EncryptionType_AES_ECB:
		return WriteAndEncryptECB(buffer, key, msgs...)
	}

	return fmt.Errorf("Encoding %v is not supported", encType)
}
