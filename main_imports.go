package main

import (
	// Transports
	_ "github.com/open-iot-devices/server/transport/udp"

	// Device handlers
	_ "github.com/open-iot-devices/server/handlers/logger"

	// Protobufs
	_ "github.com/open-iot-devices/protobufs/go/openiot"
)
