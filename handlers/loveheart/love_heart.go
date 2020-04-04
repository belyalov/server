package love_heart

import (
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"

	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/encode"

	love_pb "github.com/belyalov/love_heart/Proto/go"
)

const handlerName = "love_heart"

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
	// Prepare LoveHeart control message
	ctrl := &love_pb.Control{}
	hour, _, _ := time.Now().Clock()
	if hour >= 10 && hour < 20 {
		ctrl.EnableAnimation = true
	}
	glog.Infof("Animation is %v", ctrl.EnableAnimation)

	// Send it
	payload, err := encode.MakeReadyToSendDeviceMessage(device, ctrl)
	if err == nil {
		err = device.Transport().Send(payload)
	}
	return err
}

// Register device handler
func init() {
	device.MustAddHandler(&deviceHandler{})
}
