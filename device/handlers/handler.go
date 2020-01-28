package handlers

// DeviceHandler defines communication interface between
// system and device message handlers
type DeviceHandler interface {
	ProcessMessage() error
}

type NoHandler struct {
}

func (*NoHandler) ProcessMessage() error {
	return nil
}
