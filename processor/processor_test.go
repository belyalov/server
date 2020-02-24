package processor

// func TestDeviceJoin(t *testing.T) {
// 	device.DeleteAllDevices()
// 	pendingDevices = map[uint64][]byte{}

// 	// Generate "device" DH request
// 	privateA, publicA := generateDiffieHellman(dhG, dhP)

// 	// Craft Join Request message
// 	hdr := &openiot.Header{
// 		DeviceId: 1,
// 		Encryption: &openiot.Header_Plain{
// 			Plain: true,
// 		},
// 	}
// 	jreq := &openiot.SystemJoinRequest{
// 		DhG: dhG,
// 		DhP: dhP,
// 		DhA: publicA,
// 	}
// 	var reqBuf bytes.Buffer
// 	encode.WritePlain(&reqBuf, hdr, jreq)

// 	// "Send" it with mock source transport
// 	mockTr := &mockTransport{}
// 	err := ProcessMessage(&Message{
// 		Payload: reqBuf.Bytes(),
// 		Source:  mockTr,
// 	})
// 	assert.NoError(t, err)

// 	// Ensure that device has been added into temporary map / response sent
// 	assert.Equal(t, 1, len(pendingDevices))
// 	require.Equal(t, 1, len(mockTr.history))

// 	// De-serialize response
// 	respBuf := bytes.NewBuffer(mockTr.history[0])
// 	jresp := &openiot.SystemJoinResponse{}
// 	err = encode.ReadPlain(respBuf, hdr, jresp)
// 	require.NoError(t, err)
// 	// Calculate diffie-hellman key for "device"
// 	key := calculateDiffieHellmanKey(jreq.DhP, jresp.DhB, privateA)
// 	fmt.Println("clie key", key)

// 	t.Fail()
// }
