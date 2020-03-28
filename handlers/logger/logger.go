package logger

import (
	"fmt"
	"reflect"

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

func (h *deviceHandler) ProcessMessage(device *device.Device, msg proto.Message) error {
	// Extract and log all proto field name/value pairs
	for _, value := range extractNameValuesFromMessage(msg) {
		glog.Infof("%v", value)
	}
	return nil
}

func (h *deviceHandler) AddDevice(device *device.Device) {}

func extractNameValuesFromMessage(msg proto.Message) []string {
	results := []string{}
	reflected := reflect.Indirect(reflect.ValueOf(msg))

	for i := 0; i < reflected.NumField(); i++ {
		// skip protobuf internals (not marked with "protobuf" fields)
		tag, ok := reflected.Type().Field(i).Tag.Lookup("protobuf")
		if !ok {
			continue
		}
		value := reflected.Field(i)
		// skip empty structs
		if value.Kind() == reflect.Ptr && value.IsNil() {
			continue
		}
		str := fmt.Sprintf("%v: %+v",
			utils.ProtoGetFieldNameFromTag(tag),
			reflected.Field(i).Interface(),
		)
		results = append(results, str)
	}
	// Skip non protobuf fields
	return results
}

// Register device handler
func init() {
	device.MustAddHandler(&deviceHandler{})
}
