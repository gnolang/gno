package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/stretchr/testify/require"
)

// MaxReceiveLog bounds what one peer message may put in a reactor's log, and
// in the error the switch logs again when it stops the peer.
const MaxReceiveLog = 1024

// BadTypeURLMessage returns an Any whose type URL is n bytes of 0xff, which
// is not ASCII text.
func BadTypeURLMessage(n int) []byte {
	return typeURLMessage(bytes.Repeat([]byte{0xff}, n))
}

// UnknownTypeURLMessage returns an Any whose type URL is n printable bytes
// naming no registered type.
func UnknownTypeURLMessage(n int) []byte {
	return typeURLMessage(bytes.Repeat([]byte{'a'}, n))
}

func typeURLMessage(typeURL []byte) []byte {
	return append(binary.AppendUvarint([]byte{0x0a}, uint64(len(typeURL))), typeURL...)
}

// StopRecorder is a switch that records why each peer was stopped.
type StopRecorder struct {
	p2p.Switch
	Errs []error
}

func (s *StopRecorder) StopPeerForError(_ p2p.PeerConn, err error) {
	s.Errs = append(s.Errs, err)
}

// DebugLogger logs every level to w as text, the way a node at debug does.
func DebugLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// RequireBoundedReceive runs receive and requires its log, and the error of
// the stop it causes when stop is set, to stay within MaxReceiveLog.
func RequireBoundedReceive(t *testing.T, logged *bytes.Buffer, sw *StopRecorder, name string, stop bool, receive func()) {
	t.Helper()

	logged.Reset()
	stops := len(sw.Errs)

	receive()

	require.LessOrEqual(t, logged.Len(), MaxReceiveLog, "%s: log", name)
	if !stop {
		require.Len(t, sw.Errs, stops, "%s: peer stopped", name)
		return
	}
	require.Len(t, sw.Errs, stops+1, "%s: peer not stopped", name)
	require.LessOrEqual(t, len(fmt.Sprintf("%+v", sw.Errs[stops])), MaxReceiveLog, "%s: stop error", name)
}

// UnhandledMessage is a valid message that no reactor handles, being of
// another package than every reactor's.
type UnhandledMessage struct {
	Data []byte
}

func (*UnhandledMessage) ValidateBasic() error { return nil }

// RejectedMessage is a message whose ValidateBasic fails with an error
// carrying all of Data.
type RejectedMessage struct {
	Data []byte
}

func (m *RejectedMessage) ValidateBasic() error {
	return errors.Wrap(errors.New("rejected"), string(m.Data))
}

// MalformedMessage returns a message of n bytes of data that never decodes,
// with an error quoting all of them, as params.Param's does for a malformed
// entry.
func MalformedMessage(n int) []byte {
	return amino.MustMarshalAny(&malformedMessage{Data: strings.Repeat("x", n)})
}

type malformedMessage struct {
	Data string
}

func (m malformedMessage) MarshalAmino() (string, error) { return m.Data, nil }

func (m *malformedMessage) UnmarshalAmino(data string) error {
	return fmt.Errorf("malformed %q", data)
}

func (malformedMessage) ValidateBasic() error { return nil }

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/internal/p2p",
	"p2ptesting",
	amino.GetCallersDirname(),
).WithTypes(
	&UnhandledMessage{},
	&RejectedMessage{},
	&malformedMessage{},
))

// Deliver sends a message of n bytes on the channel desc describes, between
// two connections built with desc, and returns nil once it arrives, or the
// error the receiving connection stops with.
func Deliver(t *testing.T, desc conn.ChannelDescriptor, n int) error {
	t.Helper()

	received := make(chan struct{}, 1)
	failed := make(chan error, 1)

	cfg := conn.DefaultMConnConfig()
	cfg.SendRate, cfg.RecvRate = 1<<30, 1<<30 // so the largest message takes no time

	local, remote := net.Pipe()
	recv, other := desc, desc
	receiver := conn.NewMConnectionWithConfig(local, []*conn.ChannelDescriptor{&recv},
		func(byte, []byte) { received <- struct{}{} },
		func(err error) { failed <- err },
		cfg,
	)
	sender := conn.NewMConnectionWithConfig(remote, []*conn.ChannelDescriptor{&other},
		func(byte, []byte) {}, func(error) {}, cfg)

	require.NoError(t, receiver.Start())
	defer receiver.Stop()
	require.NoError(t, sender.Start())
	defer sender.Stop()

	require.True(t, sender.Send(desc.ID, make([]byte, n)))

	select {
	case <-received:
		return nil
	case err := <-failed:
		return err
	case <-time.After(10 * time.Second):
		t.Fatalf("no delivery or error for a %d-byte message", n)
		return nil
	}
}
