package ledger

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/zondax/hid"
	ledger_go "github.com/zondax/ledger-go"
)

// hidDevice exchanges APDUs with a Ledger device over HID, in ledger-go's
// framing, writing one 64-byte report per call.
//
// ledger-go's own transport passes every report of an APDU to a single write.
// On macOS, after such a write for an APDU spanning several reports (over 57
// bytes), the next write to a device running the Gno app times out, and the
// device stops receiving until it is unplugged.
type hidDevice struct {
	dev reportIO
}

// reportIO is the part of [hid.Device] hidDevice uses.
type reportIO interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Close() error
}

var _ ledger_go.LedgerDevice = (*hidDevice)(nil)

// connectHID opens the APDU interface of the first Ledger device, chosen as
// ledger-go chooses it.
func connectHID() (*hidDevice, error) {
	for _, info := range hid.Enumerate(ledger_go.VendorLedger, 0) {
		if info.UsagePage != ledger_go.UsagePageLedgerNanoS && info.Interface != 0 {
			continue
		}
		dev, err := info.Open()
		if err != nil {
			return nil, err
		}
		return &hidDevice{dev: dev}, nil
	}
	return nil, errors.New("no Ledger device found: it may be locked or in use by another application")
}

func (d *hidDevice) Exchange(apdu []byte) ([]byte, error) {
	if len(apdu) < 5 || int(apdu[4]) != len(apdu)-5 {
		return nil, errors.New("malformed APDU")
	}
	reports, err := ledger_go.WrapCommandAPDU(ledger_go.Channel, apdu, ledger_go.PacketSize)
	if err != nil {
		return nil, err
	}
	for len(reports) > 0 {
		if _, err := d.dev.Write(reports[:ledger_go.PacketSize]); err != nil {
			return nil, err
		}
		reports = reports[ledger_go.PacketSize:]
	}

	resp, err := d.read()
	if err != nil {
		return nil, err
	}
	if len(resp) < 2 {
		return nil, fmt.Errorf("response of %d bytes has no status word", len(resp))
	}
	data, sw := resp[:len(resp)-2], binary.BigEndian.Uint16(resp[len(resp)-2:])
	if sw != 0x9000 {
		// As ledger-go reports it, so that statusWord recovers it.
		return data, errors.New(ledger_go.ErrorMessage(sw))
	}
	return data, nil
}

// read collects the reports of one response.
func (d *hidDevice) read() ([]byte, error) {
	var (
		resp []byte
		size int
		seq  uint16
	)
	for {
		report := make([]byte, ledger_go.PacketSize)
		n, err := d.dev.Read(report)
		if err != nil {
			return nil, err
		}
		chunk, total, _, err := ledger_go.DeserializePacket(ledger_go.Channel, report[:n], seq)
		if err != nil {
			if seq == 0 {
				// Not the start of a response, such as an empty report some
				// devices send on macOS: wait for the next one.
				continue
			}
			return nil, err
		}
		if seq == 0 {
			size = int(total)
		}
		resp = append(resp, chunk...)
		seq++
		if len(resp) >= size {
			return resp[:size], nil
		}
	}
}

func (d *hidDevice) Close() error {
	return d.dev.Close()
}
