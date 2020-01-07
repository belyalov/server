package device

// Device is top level interface for all IoT devices
type Device interface {
	SetEncryptionKey(key []byte)
	GetEncryptionKey() []byte

	SetDeviceID(id uint64)
	GetDeviceID() uint64
}

// BaseDevice partially implements common methods
// of Device interface
type BaseDevice struct {
	id  uint64
	key []byte
}

// UnknownDevice represent unsupported device
// Most common use case is just joined device.
type UnknownDevice struct {
	BaseDevice
}

// NewUknownDevice creates instance of UnknownDevice
func NewUknownDevice(id uint64) Device {
	return &UnknownDevice{
		BaseDevice: BaseDevice{
			id: id,
		},
	}
}

// SetEncryptionKey sets encryption key
func (b *BaseDevice) SetEncryptionKey(key []byte) {
	b.key = key
}

// GetEncryptionKey returns currentl encryption key
func (b *BaseDevice) GetEncryptionKey() []byte {
	return b.key
}

// SetDeviceID sets encryption key
func (b *BaseDevice) SetDeviceID(id uint64) {
	b.id = id
}

// GetDeviceID returns currentl encryption key
func (b *BaseDevice) GetDeviceID() uint64 {
	return b.id
}
