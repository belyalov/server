package belyalov

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/open-iot-devices/server/device"
	"github.com/open-iot-devices/server/encode"
	"github.com/open-iot-devices/server/utils/sun"

	pb "github.com/belyalov/protobufs/go"
)

const (
	handlerName      = "belyalov"
	asteriskDeviceID = uint64(0x343534340029001e)
)

type deviceHandler struct{}

func (h *deviceHandler) GetName() string {
	return handlerName
}

func (h *deviceHandler) Start() error {
	return nil
}

func (h *deviceHandler) Stop() {
}

func (h *deviceHandler) AddDevice(dev *device.Device) {

}

func (h *deviceHandler) ProcessMessage(dev *device.Device, rawMsg proto.Message) error {
	msg, ok := rawMsg.(*pb.Status)
	if !ok {
		return fmt.Errorf("Got %T, but belyalov custom handler expects only pb.Status", rawMsg)
	}
	if msg.MishaControl != nil {
		return h.processMishaMessage(dev, msg)
	}
	// Prepare control message
	ctrl := &pb.Control{
		LoveHeart: &pb.LoveHeartControl{},
		Tulip:     &pb.TulipControl{},
		WallSpotlight: &pb.WallSpotlightControl{
			HourOn:  uint32(sun.GetSunset().Hour()),
			HourOff: uint32(sun.GetSunrise().Hour() + 1),
		},
		WallCtrlSpotlight: &pb.WallCtrlSpotlightControl{
			HourOn:  17,
			HourOff: 23,
		},
	}
	hour, _, _ := time.Now().Clock()
	if hour >= 10 && hour < 20 {
		ctrl.LoveHeart.EnableAnimation = true
		ctrl.Tulip.EnableAnimation = true
	}

	// Send it
	payload, err := encode.MakeReadyToSendDeviceMessage(dev, ctrl)
	if err == nil {
		err = dev.Transport().Send(payload)
	}
	return err
}

func (h *deviceHandler) processMishaMessage(dev *device.Device, msg *pb.Status) error {
	// Button2 points to "Asterisk" wallSpotlight which has only white LEDs
	if msg.MishaControl.Button2 {
		ctrl := &pb.Control{
			WallCtrlSpotlight: &pb.WallCtrlSpotlightControl{
				HourOn:  17,
				HourOff: 23,
				W:       true,
			},
		}
		asteriskDev := device.FindDeviceByID(asteriskDeviceID)
		if asteriskDev == nil {
			return fmt.Errorf("Device asterisk (%x) not found", asteriskDeviceID)
		}
		payload, err := encode.MakeReadyToSendDeviceMessage(asteriskDev, ctrl)
		if err == nil {
			err = dev.Transport().Send(payload)
		}
		return err
	}

	return nil
}

// Register device handler
func init() {
	device.MustAddHandler(&deviceHandler{})
}
