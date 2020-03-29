package logger

import (
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"

	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/utils"
)

const handlerName = "logger"

type deviceHandler struct{}

func (h *deviceHandler) GetName() string {
	return handlerName
}

func (h *deviceHandler) Start() error {
	return nil
}

func (h *deviceHandler) Stop() {
}

func (h *deviceHandler) AddDevice(device *device.Device) {

}

func (h *deviceHandler) ProcessMessage(device *device.Device, msg proto.Message) error {
	// Extract and log all proto field name/value pairs
	for name, value := range utils.ExtractAllNameValuesFromProtobuf(msg) {
		glog.Infof("%v: %v", name, value)
	}
	return nil
}

// Register device handler
func init() {
	device.MustAddHandler(&deviceHandler{})
}
