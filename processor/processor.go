package processor

import (
	"github.com/open-iot-devices/server/transport"
)

// ProcessMessage decodes / deserializes raw packet and calls appropriate handler
func ProcessMessage(wire transport.Transport, packet []byte) error {
	// // IoT message contains 2 protobufs in series using "delimited" mode
	// // i.e. when message length prepends the message.
	// buffer := bytes.NewBuffer(packet)

	// // Read HeaderMessage (always non-encrypted)
	// header := &openiot.HeaderMessage{}
	// err := encoding.DeserializeSingleMessage(buffer, header)
	// if err != nil {
	// 	return err
	// }

	// // Zero DeviceID is not valid
	// if header.DeviceId == 0 {
	// 	return fmt.Errorf("Invalid Packet Header: zero DeviceID")
	// }

	// // Special case for System Messages:
	// if header.SystemMessage {
	// 	sysMsg := &openiot.SystemMessage{}
	// 	err = encoding.DeserializeSingleMessage(buffer, sysMsg)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return processSystemMessage(wire, header, sysMsg)
	// }

	// // Device based message

	return nil
}

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
