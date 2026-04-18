//go:build !linux && !darwin && !windows

package hid

// Device is a stub matching the zondax/hid API on platforms where HID
// access is unavailable.
type Device struct {
	DeviceInfo
}

// Supported returns false on platforms that lack a usbhid backend.
func Supported() bool {
	return false
}

// Enumerate returns nil on unsupported platforms.
func Enumerate(vendorID uint16, productID uint16) []DeviceInfo {
	return nil
}

// Open always fails with ErrUnsupportedPlatform.
func (info DeviceInfo) Open() (*Device, error) {
	return nil, ErrUnsupportedPlatform
}

func (dev *Device) Write(b []byte) (int, error) {
	return 0, ErrUnsupportedPlatform
}

func (dev *Device) Read(b []byte) (int, error) {
	return 0, ErrUnsupportedPlatform
}

func (dev *Device) Close() error {
	return nil
}
