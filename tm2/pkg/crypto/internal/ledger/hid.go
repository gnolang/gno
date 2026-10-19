package ledger

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// device exchanges APDUs with a Ledger app. Exchange returns the response
// data; a status word other than 0x9000 is returned as a *statusError.
type device interface {
	Exchange(apdu []byte) ([]byte, error)
	Close() error
}

// swDeviceLocked is the status word the device answers with while locked,
// whichever app is open.
const swDeviceLocked = 0x5515

// statusError is a status word other than 0x9000 and the data sent with it.
type statusError struct {
	sw   uint16
	data []byte
}

func (e *statusError) Error() string {
	return fmt.Sprintf("Ledger status word 0x%04X", e.sw)
}

// HID framing of APDUs, as Ledger devices expect it: each 64-byte report
// carries the channel, a tag, a sequence number and, in the first report, the
// length of the whole APDU.
const (
	hidReportSize = 64
	hidChannel    = 0x0101
	hidTag        = 0x05
)

// frameAPDU splits apdu into the reports that carry it.
func frameAPDU(apdu []byte) [][]byte {
	var reports [][]byte
	rest := apdu
	for seq := uint16(0); seq == 0 || len(rest) > 0; seq++ {
		report := make([]byte, hidReportSize)
		binary.BigEndian.PutUint16(report, hidChannel)
		report[2] = hidTag
		binary.BigEndian.PutUint16(report[3:], seq)
		header := 5
		if seq == 0 {
			binary.BigEndian.PutUint16(report[5:], uint16(len(apdu)))
			header = 7
		}
		n := copy(report[header:], rest)
		rest = rest[n:]
		reports = append(reports, report)
	}
	return reports
}

// hidDevice exchanges APDUs over HID, writing one report per call.
//
// ledger-go passed every report of an APDU to a single write. On macOS,
// after such a write for an APDU spanning several reports (over 57 bytes),
// the next write to a device running the Gno.land app timed out, and the
// device stopped receiving until it was unplugged. Pure-Go HID backends
// refuse writes longer than a report outright.
type hidDevice struct {
	dev reportIO
}

// reportIO reads and writes single HID reports.
type reportIO interface {
	Write(report []byte) (int, error)
	Read(report []byte) (int, error)
	Close() error
}

var _ device = (*hidDevice)(nil)

func (d *hidDevice) Exchange(apdu []byte) ([]byte, error) {
	if len(apdu) < 5 || int(apdu[4]) != len(apdu)-5 {
		return nil, errors.New("malformed APDU")
	}
	for _, report := range frameAPDU(apdu) {
		if _, err := d.dev.Write(report); err != nil {
			return nil, err
		}
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
		return data, &statusError{sw: sw, data: data}
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
		report := make([]byte, hidReportSize)
		n, err := d.dev.Read(report)
		if err != nil {
			return nil, err
		}
		report = report[:n]
		header := 5
		if seq == 0 {
			header = 7
		}
		if len(report) < header || binary.BigEndian.Uint16(report) != hidChannel ||
			report[2] != hidTag || binary.BigEndian.Uint16(report[3:]) != seq {
			if seq == 0 {
				// Not the start of a response, such as an empty report some
				// devices send on macOS: wait for the next one.
				continue
			}
			return nil, fmt.Errorf("unexpected HID report %d of a response", seq)
		}
		if seq == 0 {
			size = int(binary.BigEndian.Uint16(report[5:]))
		}
		resp = append(resp, report[header:]...)
		seq++
		if len(resp) >= size {
			return resp[:size], nil
		}
	}
}

func (d *hidDevice) Close() error {
	return d.dev.Close()
}
