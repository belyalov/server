package processor

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"reflect"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/encode"
	"github.com/open-iot-devices/server/transport"
)

// Message contains packet payload and source transport
type Message struct {
	Source  transport.Transport
	Payload []byte
}

// ProcessMessage decodes / de-serializes raw packet and calls appropriate handler
func ProcessMessage(message *Message) error {
	buf := bytes.NewBuffer(message.Payload)

	// First message (openiot.Header) is always unencrypted
	hdr := &openiot.Header{}
	if err := encode.ReadSingleMessage(buf, hdr); err != nil {
		return err
	}

	// Check CRC of message payload
	if hdr.Crc != crc32.ChecksumIEEE(buf.Bytes()) {
		return fmt.Errorf("CRC check failed")
	}

	// Process Network Join Requests
	if hdr.KeyExchange {
		return processKeyExchangeRequest(hdr, buf, message.Source)
	}
	if hdr.JoinRequest {
		return processJoinRequest(hdr, buf, message.Source)
	}

	// At this point we serve only registered devices
	dev := device.FindDeviceByID(hdr.DeviceId)
	if dev == nil {
		return fmt.Errorf("Device 0x%x is not registered", hdr.DeviceId)
	}

	// De-Serialize MessageInfo
	info := &openiot.MessageInfo{}
	msgType := proto.MessageType(dev.ProtobufName)
	if msgType == nil {
		return fmt.Errorf("0x%x: Protobuf '%s' is not registered", dev.ID, dev.ProtobufName)
	}

	// De-Serialize device message
	msg := reflect.New(msgType.Elem()).Interface().(proto.Message)
	err := encode.DecryptAndRead(buf, dev.EncryptionType, dev.Key(), info, msg)
	if err != nil {
		return fmt.Errorf("0x%x: decrypt/deserialize failed: %v", dev.ID, err)
	}

	// Drop duplicates
	if info.Sequence <= dev.SequenceReceive {
		return fmt.Errorf("0x%x: drop duplicate packet seq %d (last seq %d)",
			dev.ID,
			info.Sequence,
			dev.SequenceReceive,
		)
	}
	dev.SequenceReceive = info.Sequence

	// Run all associated handlers
	glog.Infof("Message from %s/%s/%s",
		message.Source.GetTypeName(),
		message.Source.GetName(),
		dev.DisplayName,
	)
	for _, handler := range dev.Handlers() {
		handler.ProcessMessage(dev, msg)
	}

	return nil
}
