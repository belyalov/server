package processor

import (
	"bytes"
	"fmt"
	"hash/crc32"

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
	// glog.Infof("Got packet from %s", message.Source.GetName())
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

	// Send response, if provided by handler
	// if response != nil {
	// 	message.Source.Send(encode.MakeReadyToSendDeviceMessage())
	// 	// message.Source.Send()
	// }

	// Create "placeholder" for all possible
	// proto messages for this particular device
	// msgs := make([]proto.Message, len(dev.Protobufs)+1)
	// msgs[0] = &openiot.Message{}
	// for index, name := range dev.Protobufs {
	// 	msg, err := createProtoFromName(name)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	msgs[index+1] = msg
	// }
	// // Decrypt (if needed) then read all other messages
	// switch x := hdr.Encryption.(type) {
	// case *openiot.Header_Plain:
	// 	if err := encode.ReadPlain(buffer, msgs...); err != nil {
	// 		return err
	// 	}
	// case *openiot.Header_AesEcb:
	// 	if err := encode.DecryptAndReadECB(buffer, dev.Key(), msgs...); err != nil {
	// 		return err
	// 	}
	// default:
	// 	return fmt.Errorf("Unsupported encryption type %T", x)
	// }
	// Decrypt (if needed) then read all other messages
	// fmt.Println(dev.MessageNames)
	// mt := proto.MessageType("openiot.SystemJoinRequest")
	// // tp := &openiot.SystemJoinRequest{}
	// // inst := reflect.New(tp).Elem().Interface()
	// // inst := reflect.Zero(tp).Interface()
	// inst := reflect.New(mt.Elem()).Interface().(proto.Message)
	// // De-serialize all followed messages and run them through all device's handler
	// // Ensure that all messages are known to the system
	// for _, name := range msg.Names {
	// 	if proto.MessageType(name) == nil {
	// 		return fmt.Errorf("OpenIoT Message '%s' is not supported", name)
	// 	}
	// }

	return nil
}
