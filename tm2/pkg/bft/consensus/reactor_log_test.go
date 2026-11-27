package consensus

import (
	"bytes"
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/stretchr/testify/require"
)

// wideBitArray declares 2^40 bits over a million words, about 1 MB on the
// wire, so it fails every ValidateBasic that checks a bit array.
func wideBitArray() *bitarray.BitArray {
	return &bitarray.BitArray{Bits: 1 << 40, Elems: make([]uint64, 1_000_000)}
}

// TestReceiveBoundsLogOfPeerMessage sends messages that fail decoding or
// validation, and valid ones carrying the largest bit array validation admits,
// and bounds what each puts in the log at debug level.
func TestReceiveBoundsLogOfPeerMessage(t *testing.T) {
	t.Parallel()

	wide := wideBitArray()
	nvb := amino.MustMarshalAny(&NewValidBlockMessage{
		Height:           1,
		BlockPartsHeader: types.PartSetHeader{Total: 3},
		BlockParts:       wide,
	})

	// Its validation error would carry all 500000 bytes of data; the reactor
	// refuses it as another package's type before validation runs.
	rejected := amino.MustMarshalAny(&p2pTesting.RejectedMessage{Data: make([]byte, 500_000)})

	// Valid messages whose bit array renders one character per vote, 10000
	// characters from about 200 bytes on the wire.
	pol := amino.MustMarshalAny(&ProposalPOLMessage{
		Height:      1,
		ProposalPOL: bitarray.NewBitArray(types.MaxVotesCount),
	})
	vsb := amino.MustMarshalAny(&VoteSetBitsMessage{
		Height: 1,
		Type:   types.PrevoteType,
		Votes:  bitarray.NewBitArray(types.MaxVotesCount),
	})

	cases := []struct {
		name string
		chID byte
		msg  []byte
		stop bool
	}{
		{"invalid NewValidBlockMessage", StateChannel, nvb, true},
		{"invalid ProposalPOLMessage", DataChannel, amino.MustMarshalAny(&ProposalPOLMessage{
			Height:      1,
			ProposalPOL: wide,
		}), true},
		{"invalid VoteSetBitsMessage", VoteSetBitsChannel, amino.MustMarshalAny(&VoteSetBitsMessage{
			Height: 1,
			Type:   types.PrevoteType,
			Votes:  wide,
		}), true},
		{"trailing bytes", StateChannel, slices.Concat(nvb, []byte{0x08, 0x01}), true},
		{"non-ASCII type URL", StateChannel, p2pTesting.BadTypeURLMessage(1_000_000), true},
		{"unknown type URL", StateChannel, p2pTesting.UnknownTypeURLMessage(1_000_000), true},
		{"invalid message of another module", StateChannel, rejected, true},
		{"malformed message of another module", StateChannel, p2pTesting.MalformedMessage(1_000_000), true},
		{"valid ProposalPOLMessage", DataChannel, pol, false},
		{"valid VoteSetBitsMessage", VoteSetBitsChannel, vsb, false},
		{"VoteSetBitsMessage on VoteChannel", VoteChannel, vsb, false},
	}

	thisConfig, _ := ResetConfig("consensus_reactor_log_test")
	defer os.RemoveAll(thisConfig.RootDir)

	state, privVals := randGenesisState(1, false, 10)
	cs := newConsensusStateWithConfig(thisConfig, state, privVals[0], kvstore.NewKVStoreApplication())

	var logged bytes.Buffer
	sw := &p2pTesting.StopRecorder{}

	// fastSync leaves the consensus state stopped, and Receive drops what the
	// data, vote and vote-set-bits channels carry, logging it first.
	reactor := NewConsensusReactor(cs, true)
	reactor.SetLogger(p2pTesting.DebugLogger(&logged))
	reactor.SetSwitch(sw)

	peer := p2pTesting.NewPeer(t)
	defer peer.Stop()
	reactor.InitPeer(peer)

	// A reactor not running yet logs the message it drops at debug level.
	reactor.Receive(StateChannel, peer, nvb)

	require.NotZero(t, logged.Len(), "reactor not running")
	require.LessOrEqual(t, logged.Len(), p2pTesting.MaxReceiveLog, "reactor not running: log")
	require.Empty(t, sw.Errs, "reactor not running: peer stopped")

	require.NoError(t, reactor.Start())
	defer reactor.Stop()

	for _, tc := range cases {
		require.Less(t, len(tc.msg), maxMsgSize, tc.name)
		p2pTesting.RequireBoundedReceive(t, &logged, sw, tc.name, tc.stop, func() {
			reactor.Receive(tc.chID, peer, tc.msg)
		})
	}

	// A valid message's line names the height and round it is about.
	logged.Reset()
	reactor.Receive(VoteSetBitsChannel, peer, amino.MustMarshalAny(&VoteSetBitsMessage{
		Height: 7,
		Round:  3,
		Type:   types.PrevoteType,
		Votes:  bitarray.NewBitArray(1),
	}))
	require.Contains(t, logged.String(), "height=7 round=3")
}

// TestHandleMsgBoundsLogOfRejectedBlockPart hands the consensus state a block
// part from a peer that it refuses, carrying the 100 aunts validation admits,
// and bounds the line that reports it.
func TestHandleMsgBoundsLogOfRejectedBlockPart(t *testing.T) {
	t.Parallel()

	thisConfig, _ := ResetConfig("consensus_state_log_test")
	defer os.RemoveAll(thisConfig.RootDir)

	state, privVals := randGenesisState(1, false, 10)
	cs := newConsensusStateWithConfig(thisConfig, state, privVals[0], kvstore.NewKVStoreApplication())

	var logged bytes.Buffer
	cs.SetLogger(slog.New(slog.NewTextHandler(&logged, nil)))

	// The proposal has one part, and this part's index is past it.
	cs.ProposalBlockParts = types.NewPartSetFromHeader(types.PartSetHeader{Total: 1})

	aunts := make([][]byte, 100)
	for i := range aunts {
		aunts[i] = make([]byte, tmhash.Size)
	}
	part := &types.Part{
		Index: 1,
		Proof: merkle.SimpleProof{Total: 1, LeafHash: make([]byte, tmhash.Size), Aunts: aunts},
	}
	require.NoError(t, part.ValidateBasic())

	cs.handleMsg(msgInfo{&BlockPartMessage{Height: cs.Height, Round: cs.Round, Part: part}, "peer"})

	require.NotZero(t, logged.Len())
	require.LessOrEqual(t, logged.Len(), p2pTesting.MaxReceiveLog)
}
