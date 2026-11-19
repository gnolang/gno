package blockchain

import (
	"bytes"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/consensus"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/require"
)

// TestReceiveBoundsLogOfPeerMessage sends messages of about 1 MB that fail
// decoding, and bounds what each puts in the log at debug level.
func TestReceiveBoundsLogOfPeerMessage(t *testing.T) {
	t.Parallel()

	// A consensus message, refused as another package's type. Its bit array
	// declares 2^40 bits over a million words.
	nvb := amino.MustMarshalAny(&consensus.NewValidBlockMessage{
		Height:           1,
		BlockPartsHeader: types.PartSetHeader{Total: 3},
		BlockParts:       &bitarray.BitArray{Bits: 1 << 40, Elems: make([]uint64, 1_000_000)},
	})

	// Its validation error would carry all 500000 bytes of data; the reactor
	// refuses it as another package's type before validation runs.
	rejected := amino.MustMarshalAny(&p2pTesting.RejectedMessage{Data: make([]byte, 500_000)})

	cases := []struct {
		name string
		msg  []byte
		stop bool
	}{
		{"consensus message", nvb, true},
		{"trailing bytes", append(amino.MustMarshalAny(&bcStatusRequestMessage{Height: 1}), make([]byte, 1_000_000)...), true},
		{"non-ASCII type URL", p2pTesting.BadTypeURLMessage(1_000_000), true},
		{"invalid message of another module", rejected, true},
		{"malformed message of another module", p2pTesting.MalformedMessage(1_000_000), true},
	}

	var logged bytes.Buffer
	sw := &p2pTesting.StopRecorder{}

	// Receive reads neither the state nor the block executor for these. The
	// pool logs from its own routines, so it keeps out of the buffer.
	reactor := NewBlockchainReactor(sm.State{}, nil, store.NewBlockStore(memdb.NewMemDB()), true, nil)
	reactor.SetLogger(p2pTesting.DebugLogger(&logged))
	reactor.pool.Logger = log.NewNoopLogger()
	reactor.SetSwitch(sw)

	peer := p2pTesting.NewPeer(t)
	defer peer.Stop()

	// The pool asks the peer for blocks, so Receive decodes the large messages
	// it sends, where it drops those of a peer the pool asked nothing.
	require.NoError(t, reactor.pool.Start())
	defer reactor.pool.Stop()
	reactor.pool.SetPeerHeight(peer.ID(), 3)
	require.Eventually(t, func() bool { return reactor.pool.IsWaitingOn(peer.ID()) }, 5*time.Second, 10*time.Millisecond)

	for _, tc := range cases {
		require.Less(t, len(tc.msg), maxMsgSize, tc.name)
		p2pTesting.RequireBoundedReceive(t, &logged, sw, tc.name, tc.stop, func() {
			reactor.Receive(BlockchainChannel, peer, tc.msg)
		})
	}
}
