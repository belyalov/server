package processor

import (
	"bytes"
	"fmt"

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

// ProcessMessage decodes / deserializes raw packet and calls appropriate handler
func ProcessMessage(message *Message) error {
	// glog.Infof("Got packet from %s", message.Source.GetName())
	buffer := bytes.NewBuffer(message.Payload)

	// First message is always unencrypted openiot.Header
	hdr := &openiot.Header{}
	if err := encode.ReadSingleMessage(buffer, hdr); err != nil {
		return err
	}

	var response proto.Message
	var err error

	// Handle message
	if dev := device.FindDeviceByID(hdr.DeviceId); dev != nil {
		// Registered devices
	} else {
		// If device is unknown - it may indicate new device which is trying to
		// join network
		response, err = handleJoinNetwork(hdr, buffer)
	}

	// Process message handle errors here
	if err != nil {
		return nil
	}

	// Send response, if needed
	if response != nil {
		// send
	}

	return err
}



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

	// msgs := dev.Messages()
	// fmt.Println(msgs)

	// return fmt.Errorf("ok")

	// // De-serialize openiot.Message - basically metadata of followed message(s)
	// msg := &openiot.Message{}
	// if err := encode.ReadSingleMessage(buffer, msg); err != nil {
	// 	return err
	// }
	// // De-serialize all followed messages and run them through all device's handler
	// // Ensure that all messages are known to the system
	// for _, name := range msg.Names {
	// 	if proto.MessageType(name) == nil {
	// 		return fmt.Errorf("OpenIoT Message '%s' is not supported", name)
	// 	}
	// }

	// return nil
}

// func createProtoFromName(name string) (proto.Message, error) {
// 	if msgType := proto.MessageType(name); msgType != nil {
// 		return reflect.New(msgType.Elem()).Interface().(proto.Message), nil
// 	}
// 	return nil, fmt.Errorf("Proto message '%s' is not registered", name)
// }

// func processSystemMessage(
// 	wire transport.Transport, header *openiot.HeaderMessage, msg *openiot.SystemMessage) error {

// 	switch x := msg.Message.(type) {
// 	case *openiot.SystemMessage_JoinRequest:
// 		return processJoinRequest(wire, header, x.JoinRequest)
// 	}
// 	return nil
// }

// func processJoinRequest(
// 	wire transport.Transport, header *openiot.HeaderMessage, request *openiot.JoinRequest) error {

// 	// Validate request
// 	if len(request.DhA) != aes.BlockSize {
// 		return fmt.Errorf("Invalid JoinRequest: wrong dhA size %d, expected %d",
// 			len(request.DhA), aes.BlockSize)
// 	}

// 	// Ensure that DeviceID is valid in terms of:
// 	// - it is not registered yet, or
// 	// - if already been registered ensure that it type is unknown
// 	//   so it is safe to re-add it
// 	if dev := registry.FindDeviceByID(header.DeviceId); dev != nil {
// 		if x, ok := dev.(*device.UnknownDevice); !ok {
// 			return fmt.Errorf("Attempt to re-register known device: %T, id %x", x, header.DeviceId)
// 		}
// 		// Device already exists. It maybe just re-transmit from device side
// 		// Since this is unknown device and we're in network inclusion mode
// 		// Just re-add device and regenerate keys
// 		registry.DeleteDevice(header.DeviceId)
// 		glog.Infof("Possible re-transmit of JoinRequest (device already registered), so re-registered with new keys, device %x",
// 			header.DeviceId)
// 	}

// 	// Calculate Diffie-Hellman stuff
// 	dhB := make([]uint32, aes.BlockSize)
// 	aesKey := make([]byte, aes.BlockSize)
// 	for i := 0; i < aes.BlockSize; i++ {
// 		bPrivate := int(rand.Uint32())
// 		dhB[i] = uint32(
// 			diffieHellmanPowMod(
// 				int(request.DhG),
// 				bPrivate,
// 				int(request.DhP),
// 			),
// 		)
// 		aesKey[i] = byte(diffieHellmanPowMod(
// 			int(request.DhA[i]),
// 			bPrivate,
// 			int(request.DhP),
// 		))
// 	}

// 	// Create response
// 	response := &openiot.SystemMessage{
// 		Message: &openiot.SystemMessage_JoinResponse{
// 			JoinResponse: &openiot.JoinResponse{
// 				DhB: dhB,
// 			},
// 		},
// 	}

// 	// Create unknown device. It will be replaced with actual one
// 	// once we get first DeviceInfo message
// 	device := device.NewUknownDevice(header.DeviceId)
// 	device.SetEncryptionKey(aesKey)
// 	err := registry.AddDevice(device)
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println(aesKey)
// 	fmt.Println(response)

// 	return nil
// }
