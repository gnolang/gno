package upstream

// translator_pb_test.go: focused round-trip tests for VoteToProto /
// ProposalToProto edge cases — specifically the zero-Timestamp case
// (must omit the field on the wire to match upstream Tendermint, which
// otherwise emits a year-0001 protobuf timestamp).

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVoteToProto_ZeroTimestampPreserved(t *testing.T) {
	t.Parallel()

	// tmkms / tendermint-rs reject a SignVoteRequest whose Vote has
	// a missing (nil) Timestamp — the prost decoder produces
	// "missing timestamp field". Send the year-0001 protobuf
	// timestamp explicitly. Both sides canonicalize it identically,
	// so signatures still verify.
	v := &types.Vote{
		Type:   types.PrecommitType,
		Height: 7,
		Round:  1,
	}
	pb, err := VoteToProto(v)
	require.NoError(t, err)
	require.NotNil(t, pb)
	require.NotNil(t, pb.Timestamp,
		"VoteToProto must always emit a Timestamp — tmkms refuses to sign a Vote with a missing Timestamp field")
	assert.Equal(t, int64(-62135596800), pb.Timestamp.Seconds,
		"zero time.Time must serialize as protobuf year-0001 (-62135596800 unix seconds)")
}

func TestVoteToProto_NonZeroTimestampPreserved(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 2, 3, 4, 5, 6, time.UTC)
	v := &types.Vote{
		Type:      types.PrecommitType,
		Height:    7,
		Round:     1,
		Timestamp: now,
	}
	pb, err := VoteToProto(v)
	require.NoError(t, err)
	require.NotNil(t, pb)
	require.NotNil(t, pb.Timestamp)
	assert.Equal(t, now.UTC(), pb.Timestamp.AsTime().UTC())
}

func TestProposalToProto_ZeroTimestampPreserved(t *testing.T) {
	t.Parallel()

	p := &types.Proposal{
		Type:     types.ProposalType,
		Height:   7,
		Round:    1,
		POLRound: -1,
	}
	pb, err := ProposalToProto(p)
	require.NoError(t, err)
	require.NotNil(t, pb)
	require.NotNil(t, pb.Timestamp,
		"ProposalToProto must always emit a Timestamp — tmkms refuses a missing Timestamp field")
	assert.Equal(t, int64(-62135596800), pb.Timestamp.Seconds)
}

func TestProposalToProto_NonZeroTimestampPreserved(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 2, 3, 4, 5, 6, time.UTC)
	p := &types.Proposal{
		Type:      types.ProposalType,
		Height:    7,
		Round:     1,
		POLRound:  -1,
		Timestamp: now,
	}
	pb, err := ProposalToProto(p)
	require.NoError(t, err)
	require.NotNil(t, pb)
	require.NotNil(t, pb.Timestamp)
	assert.Equal(t, now.UTC(), pb.Timestamp.AsTime().UTC())
}
