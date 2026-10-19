package ledger

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeReports records written reports and serves queued ones.
type fakeReports struct {
	written [][]byte
	queued  [][]byte
}

func (f *fakeReports) Write(b []byte) (int, error) {
	f.written = append(f.written, bytes.Clone(b))
	return len(b), nil
}

func (f *fakeReports) Read(b []byte) (int, error) {
	if len(f.queued) == 0 {
		return 0, errors.New("no report queued")
	}
	n := copy(b, f.queued[0])
	f.queued = f.queued[1:]
	return n, nil
}

func (*fakeReports) Close() error { return nil }

// respond queues the reports carrying data and sw. A response is framed as a
// command is.
func (f *fakeReports) respond(data []byte, sw uint16) {
	f.queued = append(f.queued, frameAPDU(binary.BigEndian.AppendUint16(bytes.Clone(data), sw))...)
}

func TestFrameAPDU(t *testing.T) {
	t.Parallel()

	// The framing ledger-go and Ledger's own tools use: channel 0x0101, tag
	// 0x05, a big-endian sequence number, and the APDU length first.
	getAppName := []byte{gnoCLA, gnoInsGetAppName, 0, 0, 0}
	reports := frameAPDU(getAppName)
	require.Len(t, reports, 1)
	want := make([]byte, hidReportSize)
	copy(want, []byte{0x01, 0x01, 0x05, 0x00, 0x00, 0x00, byte(len(getAppName))})
	copy(want[7:], getAppName)
	assert.Equal(t, want, reports[0])

	// 57 bytes fill the first report; each next one carries 59.
	for _, tc := range []struct{ size, reports int }{
		{57, 1}, {58, 2}, {57 + 59, 2}, {57 + 59 + 1, 3}, {260, 5},
	} {
		apdu := make([]byte, tc.size)
		for i := range apdu {
			apdu[i] = byte(i)
		}
		reports := frameAPDU(apdu)
		require.Len(t, reports, tc.reports, "APDU of %d bytes", tc.size)
		var got []byte
		for seq, r := range reports {
			require.Len(t, r, hidReportSize)
			assert.Equal(t, []byte{0x01, 0x01, 0x05}, r[:3])
			assert.Equal(t, uint16(seq), binary.BigEndian.Uint16(r[3:]))
			if seq == 0 {
				assert.Equal(t, uint16(tc.size), binary.BigEndian.Uint16(r[5:]))
				got = append(got, r[7:]...)
			} else {
				got = append(got, r[5:]...)
			}
		}
		assert.Equal(t, apdu, got[:tc.size], "APDU of %d bytes", tc.size)
		assert.Equal(t, make([]byte, len(got)-tc.size), got[tc.size:], "padding")
	}
}

func TestHIDDeviceWritesOneReportPerCall(t *testing.T) {
	t.Parallel()

	f := &fakeReports{}
	d := &hidDevice{dev: f}
	payload := make([]byte, 255)
	for i := range payload {
		payload[i] = byte(i)
	}
	apdu := append([]byte{gnoCLA, gnoInsSignTx, gnoP1SignNext, gnoP2SignMore, 255}, payload...)
	f.respond(nil, 0x9000)

	_, err := d.Exchange(apdu)
	require.NoError(t, err)
	assert.Equal(t, frameAPDU(apdu), f.written)
	assert.Len(t, f.written, 5)
}

func TestHIDDeviceReadsMultiReportResponses(t *testing.T) {
	t.Parallel()

	f := &fakeReports{}
	d := &hidDevice{dev: f}
	data := make([]byte, 200)
	for i := range data {
		data[i] = byte(i * 7)
	}
	// An empty report before the response is skipped.
	f.queued = append(f.queued, make([]byte, hidReportSize))
	f.respond(data, 0x9000)

	got, err := d.Exchange([]byte{gnoCLA, gnoInsGetAppName, 0, 0, 0})
	require.NoError(t, err)
	assert.Equal(t, data, got)
	assert.Empty(t, f.queued)
}

func TestHIDDeviceStatusWords(t *testing.T) {
	t.Parallel()

	f := &fakeReports{}
	d := &hidDevice{dev: f}
	f.respond([]byte("detail"), 0xB00F)

	data, err := d.Exchange([]byte{gnoCLA, gnoInsGetAppName, 0, 0, 0})
	var status *statusError
	require.ErrorAs(t, err, &status)
	assert.Equal(t, uint16(0xB00F), status.sw)
	assert.Equal(t, []byte("detail"), status.data)
	assert.Equal(t, []byte("detail"), data)

	_, err = d.Exchange([]byte{gnoCLA, gnoInsGetAppName, 0, 0, 1})
	assert.ErrorContains(t, err, "malformed APDU")
}

func TestHIDDeviceRejectsOutOfSequenceReports(t *testing.T) {
	t.Parallel()

	f := &fakeReports{}
	d := &hidDevice{dev: f}
	reports := frameAPDU(make([]byte, 120))
	binary.BigEndian.PutUint16(reports[1][3:], 5)
	f.queued = reports

	_, err := d.Exchange([]byte{gnoCLA, gnoInsGetAppName, 0, 0, 0})
	assert.ErrorContains(t, err, "unexpected HID report")
}
