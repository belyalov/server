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
	"github.com/open-iot-devices/server/registry"
	"github.com/open-iot-devices/server/transport"
	"github.com/open-iot-devices/server/transport/udp"
)

var flagConfig = flag.String("config", "config.yaml", "Server Configuration Filename")
var flagDevices = flag.String("devices", "devices.yaml", "Registered Devices Filename")

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Load configuration
	configuration, err := ConfigLoadFromFile(*flagConfig)
	if err != nil {
		glog.Fatalf("Unable to read config file %s: %v", *flagConfig, err)
	}

	// Create transports
	for name, config := range configuration.UDP {
		glog.Infof("Creating UDP '%s' transport...", name)
		if instance, err := udp.NewUDP(name, config); err == nil {
			registry.MustAddTransport(instance)
		} else {
			glog.Fatalf("Unable to create UDP transport '%s': %v", name, err)
		}
	}

	// Create devices
	// TODO

	// To be able to shutdown server gracefully...
	var wg sync.WaitGroup
	ctx, ctxCancel := context.WithCancel(context.Background())

	// Start all transports
	incomingMessageCh := make(chan *processor.Message)
	for name, instance := range registry.GetAllTransports() {
		glog.Infof("Starting %s transport...", name)
		if err := instance.Start(); err != nil {
			glog.Fatalf("Unable to start transport '%s': %v", name, err)
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup, name string, instance transport.Transport) {
			for {
				select {
				case packet := <-instance.Receive():
					// Forward packet
					incomingMessageCh <- &processor.Message{
						Source:  instance,
						Payload: packet,
					}
				case <-ctx.Done():
					instance.Stop()
					glog.Infof("Transport %s terminated", name)
					wg.Done()
					return
				}
			}
		}(&wg, name, instance)
	}

	// Setup SIGTERM / SIGINT
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	glog.Info("OpeIoT server ready.")

	// Handle all incoming packets
	for {
		select {
		case message := <-incomingMessageCh:
			if err := processor.ProcessMessage(message); err != nil {
				glog.Infof("ProcessPacket failed: %v", err)
			}
		case sig := <-signalCh:
			glog.Infof("Got SIG %v", sig)
			// Cancel context and wait until all jobs done
			ctxCancel()
			wg.Wait()
			// time.Sleep(1 * time.Second)
			// Save devices configuration
			// err := devices.SaveToFile(*flagDevices)
			// if err != nil {
			// 	glog.Fatalf("Save devices config failed: %v", err)
			// }
			glog.Info("Gracefully terminated")
			glog.Flush()
			return
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
