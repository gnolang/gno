package ledger

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ledger_go "github.com/zondax/ledger-go"
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

// respond queues the reports carrying data and sw, as a device frames them.
func (f *fakeReports) respond(t *testing.T, data []byte, sw uint16) {
	t.Helper()
	resp := binary.BigEndian.AppendUint16(bytes.Clone(data), sw)
	// A response is framed as a command is: length, then the bytes.
	framed, err := ledger_go.WrapCommandAPDU(ledger_go.Channel, resp, ledger_go.PacketSize)
	require.NoError(t, err)
	for len(framed) > 0 {
		f.queued = append(f.queued, framed[:ledger_go.PacketSize])
		framed = framed[ledger_go.PacketSize:]
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
	f.respond(t, nil, 0x9000)

	_, err := d.Exchange(apdu)
	require.NoError(t, err)

	want, err := ledger_go.WrapCommandAPDU(ledger_go.Channel, apdu, ledger_go.PacketSize)
	require.NoError(t, err)
	require.Len(t, f.written, len(want)/ledger_go.PacketSize)
	assert.Len(t, f.written, 5)
	for i, report := range f.written {
		assert.Equal(t, want[i*ledger_go.PacketSize:(i+1)*ledger_go.PacketSize], report, "report %d", i)
	}
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
	f.queued = append(f.queued, make([]byte, ledger_go.PacketSize))
	f.respond(t, data, 0x9000)

	got, err := d.Exchange([]byte{gnoCLA, gnoInsGetAppName, 0, 0, 0})
	require.NoError(t, err)
	assert.Equal(t, data, got)
	assert.Empty(t, f.queued)
}

func TestHIDDeviceStatusWords(t *testing.T) {
	t.Parallel()

	f := &fakeReports{}
	d := &hidDevice{dev: f}
	f.respond(t, nil, 0xB00F)

	_, err := d.Exchange([]byte{gnoCLA, gnoInsGetAppName, 0, 0, 0})
	sw, ok := statusWord(err)
	require.True(t, ok)
	assert.Equal(t, uint16(0xB00F), sw)

	_, err = d.Exchange([]byte{gnoCLA, gnoInsGetAppName, 0, 0, 1})
	assert.ErrorContains(t, err, "malformed APDU")
}
