//go:build linux || darwin || windows

package ledger

import (
	"errors"

	"rafaelmartins.com/p/usbhid"
)

// hidSupported reports whether this platform has a HID backend.
const hidSupported = true

const (
	ledgerVendorID  = 0x2c97
	ledgerUsagePage = 0xffa0 // the APDU interface; FIDO uses another page
)

// usbhidReports is a HID device opened with usbhid, which needs no CGO.
type usbhidReports struct {
	dev *usbhid.Device
}

func (r usbhidReports) Write(report []byte) (int, error) {
	// Ledger reports carry no report ID.
	if err := r.dev.SetOutputReport(0, report); err != nil {
		return 0, err
	}
	return len(report), nil
}

func (r usbhidReports) Read(report []byte) (int, error) {
	_, data, err := r.dev.GetInputReport()
	if err != nil {
		return 0, err
	}
	return copy(report, data), nil
}

func (r usbhidReports) Close() error {
	return r.dev.Close()
}

// connectHID opens the APDU interface of the first Ledger device. The
// interface is told apart from the device's others (FIDO, ...) by its usage
// page.
func connectHID() (*hidDevice, error) {
	devices, err := usbhid.Enumerate(func(d *usbhid.Device) bool {
		return d.VendorId() == ledgerVendorID && d.UsagePage() == ledgerUsagePage
	})
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, errors.New("no Ledger device found: connect and unlock it, and close other applications using it")
	}
	dev := devices[0]
	if err := dev.Open(false); err != nil {
		return nil, err
	}
	return &hidDevice{dev: usbhidReports{dev: dev}}, nil
}
