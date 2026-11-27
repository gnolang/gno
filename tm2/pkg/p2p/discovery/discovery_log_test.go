package discovery_test

import (
	"bytes"
	"encoding/binary"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bech32"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/gnolang/gno/tm2/pkg/p2p/discovery"
	"github.com/gnolang/gno/tm2/pkg/p2p/mock"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/require"
)

// responseWithAddress returns a Response holding one address string, which
// fails to decode into a NetAddress with an error quoting the string.
func responseWithAddress(addr string) []byte {
	field := func(num byte, bz []byte) []byte {
		return append(binary.AppendUvarint([]byte{num<<3 | 2}, uint64(len(bz))), bz...)
	}

	return append(field(1, []byte(amino.GetTypeURL(&discovery.Response{}))), field(2, field(1, []byte(addr)))...)
}

// TestReceiveBoundsLogOfPeerMessage sends 1 MB messages the reactor does not
// take, and bounds what each puts in the log.
func TestReceiveBoundsLogOfPeerMessage(t *testing.T) {
	t.Parallel()

	addrPrefix := string(types.GenerateNodeKey().ID()) + "@1.2.3.4:"

	// An ID with a valid checksum under a 1 MB bech32 prefix of quotes, which
	// a log handler escapes to two bytes each.
	longPrefixID, err := bech32.Encode(strings.Repeat(`"`, 1_000_000), make([]byte, 20))
	require.NoError(t, err)

	cases := []struct {
		name string
		msg  []byte
	}{
		{"non-ASCII type URL", p2pTesting.BadTypeURLMessage(1_000_000)},
		{"unhandled message", amino.MustMarshalAny(&p2pTesting.UnhandledMessage{Data: make([]byte, 1_000_000)})},
		{"rejected message", amino.MustMarshalAny(&p2pTesting.RejectedMessage{Data: make([]byte, 1_000_000)})},
		{"malformed message", p2pTesting.MalformedMessage(1_000_000)},
		{"control bytes in an address", responseWithAddress(addrPrefix + strings.Repeat("\x01", 1_000_000))},
		{"multibyte runes in an address", responseWithAddress(addrPrefix + strings.Repeat("\U000E0001", 250_000))},
		{"long bech32 prefix in an address ID", responseWithAddress(longPrefixID + "@1.2.3.4:26656")},
	}

	store, err := discovery.NewStore(filepath.Join(t.TempDir(), "peers.json"), types.NetAddress{})
	require.NoError(t, err)

	var logged bytes.Buffer

	r := discovery.NewReactor(store)
	r.SetLogger(p2pTesting.DebugLogger(&logged))

	for _, tc := range cases {
		logged.Reset()

		r.Receive(discovery.Channel, &mock.Peer{}, tc.msg)

		require.NotZero(t, logged.Len(), tc.name)
		require.LessOrEqual(t, logged.Len(), p2pTesting.MaxReceiveLog, tc.name)
	}
}
