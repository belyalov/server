package main

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/golang/glog"

	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/processor"
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

	// To be able to shutdown server gracefully...
	var wg sync.WaitGroup
	doneCh := make(chan interface{})

	// Load transports
	if fd, err := os.Open(transportsFilename); err == nil {
		if err := transport.LoadTransports(fd); err != nil {
			glog.Fatalf("Unable to LoadTransports: %v", err)
		}
	} else {
		glog.Errorf("Unable to open: %v", err)
	}

	// Start device handlers
	glog.Info("Starting device handlers...")
	for name, handler := range device.GetAllHandlers() {
		glog.Infof("\t%s", name)
		if err := handler.Start(); err != nil {
			glog.Fatalf("Unable to start device handler '%s': %v",
				handler.GetName(), err)
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup, handler device.Handler) {
			<-doneCh
			handler.Stop()
			wg.Done()
			glog.Infof("\thandler %s terminated.", handler.GetName())
		}(&wg, handler)
	}

	// Load Devices
	if fd, err := os.Open(devicesFilename); err == nil {
		if err := device.LoadDevices(fd); err != nil {
			glog.Fatalf("Unable to LoadDevices: %v", err)
		}
	} else {
		glog.Errorf("Unable to open: %v", err)
	}
	// Print all devices
	glog.Info("Registered devices:")
	for _, dev := range device.GetAllDevices() {
		glog.Infof("\t%s (0x%x), handlers: %v", dev.Name, dev.ID, dev.HandlerNames)
	}

	glog.Infof("Starting transports...")
	incomingMessagesCh := make(chan *processor.Message, *flagMsgBuffer)
	for _, tr := range transport.GetAllTransports() {
		glog.Infof("\t%s/%s", tr.GetTypeName(), tr.GetName())
		if err := tr.Start(); err != nil {
			glog.Fatalf("Unable to start transport %s/%s: %v",
				tr.GetTypeName(), tr.GetName(), err)
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup, instance transport.Transport) {
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
					wg.Done()
					glog.Infof("\t%s/%s terminated.", instance.GetTypeName(), instance.GetName())
					return
				}
			}
		}(&wg, tr)
	}

	// Setup SIGTERM / SIGINT
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	glog.Info("OpenIoT server ready.")

	// Save all configuration on exit
	defer saveDevicesToFile(devicesFilename)
	defer glog.Flush()

	// Main loop, handle:
	// - all incoming packets from transports
	// - ctrl+c
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			saveDevicesToFile(devicesFilename)

		case message := <-incomingMessagesCh:
			if err := processor.ProcessMessage(message); err != nil {
				glog.Infof("ProcessPacket failed: %v", err)
			}

		case sig := <-signalCh:
			glog.Infof("Got SIG %v, terminating...", sig)
			// Gracefully shutdown everything
			close(doneCh)
			wg.Wait()
			return
		}
	}
}

func saveDevicesToFile(filename string) {
	if fd, err := os.Create(filename); err == nil {
		if err := device.SaveDevices(fd); err != nil {
			glog.Errorf("Unable to SaveDevices: %v", err)
		}
	} else {
		glog.Errorf("Unable to create: %v", err)
	}
}
