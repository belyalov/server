package main

import (
	"flag"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/golang/glog"
	"github.com/open-iot-devices/server/transport"
)

var flagConfigDir = flag.String("config", ".config", "Configuration directory")
var flagMsgBuffer = flag.Uint("buffer", 32, "Receive message buffer size, in messages")

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Setup config names and create directory
	transportsFilename := path.Join(*flagConfigDir, "transports.yaml")
	devicesFilename := path.Join(*flagConfigDir, "devices.yaml")
	os.MkdirAll(*flagConfigDir, os.ModePerm)

	// Load transports
	if fd, err := os.Open(transportsFilename); err == nil {
		if err := transport.LoadTransports(fd); err != nil {
			glog.Errorf("Unable to LoadTransports: %v", err)
		}
	} else {
		glog.Errorf("Unable to open: %v", err)
	}

	_ = devicesFilename

	// // Create transports from configuration
	// glog.Infof("Loading transports:")
	// for name, config := range transports.UDP {
	// 	glog.Infof("\t%s", name)
	// 	instance, err := udp.NewUDP(name, config)
	// 	if err != nil {
	// 		glog.Errorf("Unable to create UDP transport '%s': %v", name, err)
	// 		continue
	// 	}
	// 	transport.MustAddTransport(instance)
	// }

	// glog.Infof("Loading devices from %s...", devicesDirname)
	// config.LoadDevicesFromDirectory(devicesDirname)

	// glog.Infof("All configuration has been successfully loaded.")

	// // To be able to shutdown server gracefully...
	// var wg sync.WaitGroup
	// doneCh := make(chan interface{})

	// // Start all transports
	// incomingMessagesCh := make(chan *processor.Message, *flagMsgBuffer)
	// glog.Infof("Starting transports...")
	// for name, instance := range transport.GetAllTransports() {
	// 	if err := instance.Start(); err != nil {
	// 		glog.Fatalf("Unable to start transport '%s': %v", name, err)
	// 	}
	// 	wg.Add(1)
	// 	go func(wg *sync.WaitGroup, name string, instance transport.Transport) {
	// 		for {
	// 			select {
	// 			case packet := <-instance.Receive():
	// 				// Forward packet
	// 				incomingMessagesCh <- &processor.Message{
	// 					Source:  instance,
	// 					Payload: packet,
	// 				}
	// 			case <-doneCh:
	// 				instance.Stop()
	// 				glog.Infof("Transport %s terminated", name)
	// 				wg.Done()
	// 				return
	// 			}
	// 		}
	// 	}(&wg, name, instance)
	// }

	// // Setup SIGTERM / SIGINT
	// signalCh := make(chan os.Signal, 1)
	// signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// glog.Info("OpenIoT server ready.")

	// // Save all configuration on exit
	// defer func() {
	// 	if err := config.SaveDevicesToFile(devicesDirname); err != nil {
	// 		glog.Errorf("Failed to save devices config: %v", err)
	// 	}

	// 	defer glog.Flush()
	// }()

	// // Main loop, handle:
	// // - all incoming packets from transports
	// // - ctrl+c
	// for {
	// 	select {
	// 	case message := <-incomingMessagesCh:
	// 		if err := processor.ProcessMessage(message); err != nil {
	// 			glog.Infof("ProcessPacket failed: %v", err)
	// 		}
	// 	case sig := <-signalCh:
	// 		glog.Infof("Got SIG %v, terminating...", sig)
	// 		// Gracefully shutdown everything
	// 		close(doneCh)
	// 		wg.Wait()
	// 		glog.Info("Gracefully terminated.")
	// 		return
	// 	}
	// }
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
