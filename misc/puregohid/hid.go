// Package hid provides a pure-Go shim implementing the github.com/zondax/hid
// API on top of github.com/rafaelmartins/usbhid. This eliminates the CGO
// dependency required by the original zondax/hid package.
//
// On platforms not supported by usbhid the package compiles with stub
// implementations, matching the upstream zondax/hid behavior where
// Supported() returns false and device access fails gracefully.
package hid

import "errors"

// Error sentinels matching zondax/hid.
var (
	ErrDeviceClosed        = errors.New("hid: device closed")
	ErrUnsupportedPlatform = errors.New("hid: unsupported platform")
)

// DeviceInfo contains information about a discovered HID device.
// It is a plain exported struct so it can be copied or reconstructed
// freely — Open() looks the device up again by Path.
type DeviceInfo struct {
	Path         string
	VendorID     uint16
	ProductID    uint16
	Release      uint16
	Serial       string
	Manufacturer string
	Product      string
	UsagePage    uint16
	Usage        uint16
	Interface    int
}
