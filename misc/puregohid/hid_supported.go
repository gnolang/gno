//go:build linux || darwin || windows

package hid

import (
	"sync"

	"github.com/rafaelmartins/usbhid"
)

// Device represents an opened HID device.
type Device struct {
	DeviceInfo

	dev  *usbhid.Device
	lock sync.Mutex
}

// Supported returns whether HID is supported on this platform.
func Supported() bool {
	return true
}

// Enumerate returns all HID devices matching the given vendor and product IDs.
// A value of 0 for vendorID or productID acts as a wildcard.
func Enumerate(vendorID uint16, productID uint16) []DeviceInfo {
	devices, err := usbhid.Enumerate(func(d *usbhid.Device) bool {
		if vendorID != 0 && d.VendorId() != vendorID {
			return false
		}
		if productID != 0 && d.ProductId() != productID {
			return false
		}
		return true
	})
	if err != nil {
		return nil
	}

	infos := make([]DeviceInfo, 0, len(devices))
	for _, d := range devices {
		infos = append(infos, DeviceInfo{
			Path:         d.Path(),
			VendorID:     d.VendorId(),
			ProductID:    d.ProductId(),
			Release:      d.Version(),
			Serial:       d.SerialNumber(),
			Manufacturer: d.Manufacturer(),
			Product:      d.Product(),
			UsagePage:    d.UsagePage(),
			Usage:        d.Usage(),
			// usbhid does not expose the HID interface number; leave as 0.
			// Ledger devices are matched via UsagePage, so this is a known
			// gap only for the ledger-go fallback detection path.
			Interface: 0,
		})
	}
	return infos
}

// Open connects to the HID device described by this DeviceInfo by
// re-looking it up through usbhid using Path.
func (info DeviceInfo) Open() (*Device, error) {
	d, err := usbhid.Get(func(d *usbhid.Device) bool {
		return d.Path() == info.Path
	}, true, false)
	if err != nil {
		return nil, err
	}
	return &Device{
		DeviceInfo: info,
		dev:        d,
	}, nil
}

// Write sends an output report to the device. Returns the number of bytes
// written (len(b)) on success, matching zondax/hid behavior.
func (dev *Device) Write(b []byte) (int, error) {
	dev.lock.Lock()
	d := dev.dev
	dev.lock.Unlock()

	if d == nil {
		return 0, ErrDeviceClosed
	}
	if err := d.SetOutputReport(0, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

// Read receives an input report from the device into b. Returns the number
// of bytes copied into b.
func (dev *Device) Read(b []byte) (int, error) {
	dev.lock.Lock()
	d := dev.dev
	dev.lock.Unlock()

	if d == nil {
		return 0, ErrDeviceClosed
	}
	_, data, err := d.GetInputReport()
	if err != nil {
		return 0, err
	}
	return copy(b, data), nil
}

// Close releases the HID device handle.
func (dev *Device) Close() error {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	if dev.dev == nil {
		return nil
	}
	err := dev.dev.Close()
	dev.dev = nil
	return err
}
