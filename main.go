package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"

	"github.com/open-iot-devices/server/config"
	"github.com/open-iot-devices/server/handlers"
	"github.com/open-iot-devices/server/processor"
	"github.com/open-iot-devices/server/transport"
	"github.com/open-iot-devices/server/transport/udp"
)

var flagConfig = flag.String("config", "config.yaml", "Server Configuration Filename")
var flagDevices = flag.String("devices", "devices.yaml", "Registered Devices Filename")
var flagMsgBuffer = flag.Uint("buffer", 32, "Receive message buffer size, in messages")

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Print registered handlers
	glog.Info("Registered Device Handlers:")
	for name := range handlers.GetAllHandlers() {
		glog.Infof("\t%s", name)
	}

	// Load base configuration
	configuration, err := config.LoadConfigFromFile(*flagConfig)
	if err != nil {
		glog.Fatalf("Unable to read config file %s: %v", *flagConfig, err)
	}

	// Create transports from configuration
	glog.Infof("Creating transports:")
	for name, config := range configuration.UDP {
		glog.Infof("\t%s", name)
		instance, err := udp.NewUDP(name, config)
		if err != nil {
			glog.Fatalf("Unable to create UDP transport '%s': %v", name, err)
		}
		transport.MustAddTransport(instance)
	}

	mt := proto.MessageType("openiot.SystemJoinRequest")
	// tp := &openiot.SystemJoinRequest{}
	// inst := reflect.New(tp).Elem().Interface()
	// inst := reflect.Zero(tp).Interface()
	inst := reflect.New(mt.Elem()).Interface().(proto.Message)

	// sm := inst.(*openiot.SystemJoinRequest)
	// sm.DhP = 1
	// sm.DhG = 1
	// sm := inst.Elem().Interface()
	// sm := make(inst)
	// sm.DhP = 1
	// inst.Interface()
	// sm := inst.
	fmt.Printf("inst %T, %v\n", inst, inst)

	// Load devices
	err = config.LoadDevicesFromFile(*flagDevices)
	if err != nil {
		glog.Fatalf("Unable to load devices: %v", err)
	}
	glog.Infof("All configuration has been successfully loaded.")

	// To be able to shutdown server gracefully...
	var wg sync.WaitGroup
	doneCh := make(chan interface{})

	// Start all transports
	incomingMessagesCh := make(chan *processor.Message, *flagMsgBuffer)
	glog.Infof("Starting transports...")
	for name, instance := range transport.GetAllTransports() {
		if err := instance.Start(); err != nil {
			glog.Fatalf("Unable to start transport '%s': %v", name, err)
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup, name string, instance transport.Transport) {
			for {
				select {
				case packet := <-instance.Receive():
					// Forward packet
					incomingMessagesCh <- &processor.Message{
						Source:  instance,
						Payload: packet,
					}
				case <-doneCh:
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

	glog.Info("OpenIoT server ready.")

	// Save Devices state on exit
	defer glog.Flush()
	defer saveDevicesToFile()

	// Main loop, handle:
	// - all incoming packets from transports
	// - ctrl+c
	// - periodically save devices to disk
	devicesTicker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case message := <-incomingMessagesCh:
			if err := processor.ProcessMessage(message); err != nil {
				glog.Infof("ProcessPacket failed: %v", err)
			}
		case <-devicesTicker.C:
			saveDevicesToFile()
		case sig := <-signalCh:
			glog.Infof("Got SIG %v, terminating...", sig)
			// Gracefully shutdown everything
			close(doneCh)
			wg.Wait()
			glog.Info("Gracefully terminated.")
			return
		}
	}
}

func saveDevicesToFile() {
	if err := config.SaveDevicesToFile(*flagDevices); err != nil {
		glog.Errorf("Failed to save devices to file: %v", err)
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
