package main

import (
	// Custom
	_ "github.com/open-iot-devices/server/handlers/loveheart"

	// Transports
	_ "github.com/open-iot-devices/server/transport/udp"

	// Device handlers
	_ "github.com/open-iot-devices/server/handlers/influxdb"
	_ "github.com/open-iot-devices/server/handlers/logger"

	// Protobufs
	_ "github.com/open-iot-devices/protobufs/go/openiot"
	_ "github.com/open-iot-devices/protobufs/go/openiot/sensor"
)
