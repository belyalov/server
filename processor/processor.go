package processor

import (
	"bytes"
	"fmt"
	"math/rand"

	"github.com/open-iot-devices/protobufs/go/openiot"
	"github.com/open-iot-devices/server/encoding"
	"github.com/open-iot-devices/server/transport"
)

// ProcessMessage decodes / deserializes raw packet and calls appropriate handler
func ProcessMessage(transport transport.Transport, packet []byte) error {
	// IoT message contains 2 protobufs in series using "delimited" mode
	// i.e. when message length prepends the message.
	buffer := bytes.NewBuffer(packet)

	// Read HeaderMessage (always non-encrypted)
	hdrMsg := &openiot.HeaderMessage{}
	err := encoding.DeserializeSingleMessage(buffer, hdrMsg)
	if err != nil {
		return err
	}

	// Special case for System Messages:
	if hdrMsg.SystemMessage {
		sysMsg := &openiot.SystemMessage{}
		err = encoding.DeserializeSingleMessage(buffer, sysMsg)
		if err != nil {
			return err
		}
		return processSystemMessage(transport, sysMsg)
	}

	// Device based message

	return nil
}

func processSystemMessage(transport transport.Transport, msg *openiot.SystemMessage) error {
	switch x := msg.Message.(type) {
	case *openiot.SystemMessage_JoinRequest:
		return processJoinRequest(transport, x.JoinRequest)
	}
	return nil
}

func processJoinRequest(transport transport.Transport, request *openiot.JoinRequest) error {
	// Validate request
	if len(request.DhA) != encoding.AesBlockSize {
		return fmt.Errorf("Invalid JoinRequest: wrong dhA size %d, expected %d",
			len(request.DhA), encoding.AesBlockSize)
	}

	// Calculate Diffie-Hellman stuff
	dhB := make([]uint32, encoding.AesBlockSize)
	aesKey := make([]byte, encoding.AesBlockSize)
	for i := 0; i < encoding.AesBlockSize; i++ {
		bPrivate := int(rand.Uint32())
		dhB[i] = uint32(
			diffieHellmanPowMod(
				int(request.DhG),
				bPrivate,
				int(request.DhP),
			),
		)
		aesKey[i] = byte(diffieHellmanPowMod(
			int(request.DhA[i]),
			bPrivate,
			int(request.DhP),
		))
	}

	// Create response
	response := &openiot.SystemMessage{
		Message: &openiot.SystemMessage_JoinResponse{
			JoinResponse: &openiot.JoinResponse{
				DhB: dhB,
			},
		},
	}

	fmt.Println(aesKey)
	fmt.Println(response)

	return nil
}
