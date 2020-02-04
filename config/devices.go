package config

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/handlers"
	"github.com/open-iot-devices/server/transport"
)

// LoadDevicesFromFile reads and parses YAML devices configuration
func LoadDevicesFromFile(filename string) error {
	// Non existing config file is not an error
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
	}

	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	devices, err := loadDevicesFromYaml(bufio.NewReader(fd))
	if err != nil {
		return err
	}

	device.ReplaceAllDevicesWith(devices)
	return nil
}

func loadDevicesFromYaml(reader io.Reader) ([]*device.Device, error) {
	// Deserialize YAML
	var devices []*device.Device
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&devices); err != nil {
		return nil, err
	}

	// Update devices: setup key / transport / handlers
	for _, dev := range devices {
		// Set ID from hex string
		if id, err := strconv.ParseUint(dev.IDhex, 0, 64); err == nil {
			dev.ID = id
		} else {
			return nil, err
		}
		// Setup key from hex string representation
		if key, err := hex.DecodeString(dev.KeyString); err == nil {
			dev.Key = key
		} else {
			return nil, err
		}
		// Setup transport
		if transport := transport.FindTransportByName(dev.TransportName); transport != nil {
			dev.SetTransport(transport)
		}
		// Setup handlers
		for _, name := range dev.HandlerNames {
			handler := handlers.FindDeviceHandler(name)
			if handler == nil {
				return nil, fmt.Errorf("DeviceHandler '%s' not found. (Referenced in device %x)",
					name, dev.ID)
			}
			dev.AddDeviceHandler(handler)
		}
	}

	return devices, nil
}

// SaveDevicesToFile dumps device's data into filename on disk
func SaveDevicesToFile(filename string) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(fd)
	defer fd.Close()
	defer writer.Flush()

	// YAML header (to avoid EOF when no devices added)
	writer.WriteString("---\n")
	// Write YAML
	return saveDevicesToYaml(writer, device.GetAllDevices())
}

func saveDevicesToYaml(writer io.Writer, devices []*device.Device) error {
	for _, dev := range devices {
		// Device ID is shown to user as HEX string -
		// preserve the same idea for config as well
		dev.IDhex = fmt.Sprintf("0x%x", dev.ID)
		// By default []byte gets encoded into YAML array.
		// KeyString gets encoded into "key" yaml field with holds hexed key
		dev.KeyString = hex.EncodeToString(dev.Key)
		// dev.transport is internal object, using transportName
		// to store name of transport into YAML
		transport := dev.Transport()
		if transport != nil {
			dev.TransportName = dev.Transport().GetName()
		}
		// The same idea for handlers
		handlers := dev.Handlers()
		dev.HandlerNames = make([]string, len(handlers))
		for index, handler := range handlers {
			dev.HandlerNames[index] = handler.GetName()
		}
	}

	return yaml.NewEncoder(writer).Encode(devices)
}
