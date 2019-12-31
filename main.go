package main

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang/glog"

	"github.com/open-iot-devices/server/processor"
	"github.com/open-iot-devices/server/transport/udp"
)

var flagConfig = flag.String("config", "config.yaml", "Server Configuration Filename")
var flagDevices = flag.String("devices", "devices.yaml", "Registered Devices Filename")

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Load configuration
	cfg, err := ConfigLoadFromFile(*flagConfig)
	if err != nil {
		glog.Fatalf("Unable to read config file %s: %v", *flagConfig, err)
	}

	// Setup services
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	// caps := &devices.Capabilities{}

	// Start UDP transport
	udpTransport, err := udp.NewUDP(cfg.Udp)
	if err != nil {
		glog.Fatalf("Unable to create UDP transport: %v", err)
	}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		err := udpTransport.Run(ctx)
		if err != nil {
			glog.Fatalf("UDP transport failed to run: %v", err)
		}
		wg.Done()
	}(&wg)

	// Setup SIGTERM / SIGINT
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		// Wait for packet from any transport
		var err error
		select {
		case packet := <-udpTransport.Receive():
			err := processor.ProcessMessage(udpTransport, packet)
			if err != nil {
				glog.Infof("ProcessPacket: %v", err)
			}
		case sig := <-signalCh:
			glog.Infof("Got SIG %v", sig)
			// Cancel context and wait until all jobs done
			cancel()
			// wg.Wait()
			// Save devices configuration
			// err := devices.SaveToFile(*flagDevices)
			// if err != nil {
			// 	glog.Fatalf("Save devices config failed: %v", err)
			// }
			glog.Info("Gracefully terminated")
			glog.Flush()
			return
		}
		// Handle all errors in one place
		if err != nil {
			glog.Infof("ProcessPacket failed: %v", err)
		}
	}

}

// // Start InfluxDB
// caps.InfluxDb, err = influxdb.NewInfluxDB(cfg.InfluxDb)
// if err != nil {
// 	glog.Fatalf("InfluxDB failed: %v", err)
// }

// // Start MQTT client
// caps.Mqtt, err = mqtt.NewMqttClient(cfg.Mqtt)
// wg.Add(1)
// go func(wg *sync.WaitGroup) {
// 	err := caps.Mqtt.Run(ctx)
// 	if err != nil {
// 		glog.Fatalf("MQTT failed: %v", err)
// 	}
// 	wg.Done()
// }(&wg)

// Load / register devices
// time.Sleep(100 * time.Millisecond) // Find better solution?
// err = devices.LoadFromFile(*flagDevices, caps)
// if err != nil {
// 	glog.Fatalf("Unable to load devices: %v", err)
// }

// // Start all devices
// err = devices.StartAllDevices(ctx)
// if err != nil {
// 	glog.Fatalf("Unable to start devices: %v", err)
// }
