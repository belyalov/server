package belyalov

import (
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"

	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/encode"

	pb "github.com/belyalov/protobufs/go"
)

const handlerName = "belyalov"

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
	// Prepare control message
	ctrl := &pb.Control{
		LoveHeart: &pb.LoveHeartControl{},
		Tulip:     &pb.TulipControl{},
	}
	hour, _, _ := time.Now().Clock()
	if hour >= 10 && hour < 20 {
		ctrl.LoveHeart.EnableAnimation = true
		ctrl.Tulip.EnableAnimation = true
	}
	glog.Infof("Animation is %v", ctrl.LoveHeart.EnableAnimation)

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
