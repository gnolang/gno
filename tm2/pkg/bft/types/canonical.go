package types

import (
	"time"

	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
)

// Canonical* wraps the structs in types for amino encoding them for use in SignBytes / the Signable interface.

// TimeFormat is used for generating the sigs
const TimeFormat = time.RFC3339Nano

// Canonical types are defined to be wire-byte-compatible with upstream
// Tendermint v0.34's canonical.proto. Field order, types, and tags here
// determine the bytes that get signed; any divergence breaks signature
// verification against tmkms and other upstream-protocol consumers.

type CanonicalBlockID struct {
	Hash        []byte
	PartsHeader CanonicalPartSetHeader
}

type CanonicalPartSetHeader struct {
	Total uint32 // upstream: field 1, uint32
	Hash  []byte // upstream: field 2, bytes
}

type CanonicalProposal struct {
	Type      SignedMsgType // type alias for byte
	Height    int64         `binary:"fixed64"`
	Round     int64         `binary:"fixed64"`
	POLRound  int64         `binary:"varint"` // upstream: int64 (plain varint), not sfixed64
	BlockID   CanonicalBlockID
	Timestamp time.Time
	ChainID   string
}

type CanonicalVote struct {
	Type      SignedMsgType // type alias for byte
	Height    int64         `binary:"fixed64"`
	Round     int64         `binary:"fixed64"`
	BlockID   CanonicalBlockID
	Timestamp time.Time
	ChainID   string
}

//-----------------------------------
// Canonicalize the structs

func CanonicalizeBlockID(blockID BlockID) CanonicalBlockID {
	return CanonicalBlockID{
		Hash:        blockID.Hash,
		PartsHeader: CanonicalizePartSetHeader(blockID.PartsHeader),
	}
}

func CanonicalizePartSetHeader(psh PartSetHeader) CanonicalPartSetHeader {
	if psh.Total < 0 {
		panic("PartSetHeader.Total is negative; canonical encoding requires non-negative")
	}
	return CanonicalPartSetHeader{
		Total: uint32(psh.Total),
		Hash:  psh.Hash,
	}
}

func CanonicalizeProposal(chainID string, proposal *Proposal) CanonicalProposal {
	return CanonicalProposal{
		Type:      ProposalType,
		Height:    proposal.Height,
		Round:     int64(proposal.Round), // cast int->int64 to make amino encode it fixed64 (does not work for int)
		POLRound:  int64(proposal.POLRound),
		BlockID:   CanonicalizeBlockID(proposal.BlockID),
		Timestamp: proposal.Timestamp,
		ChainID:   chainID,
	}
}

func CanonicalizeVote(chainID string, vote *Vote) CanonicalVote {
	return CanonicalVote{
		Type:      vote.Type,
		Height:    vote.Height,
		Round:     int64(vote.Round), // cast int->int64 to make amino encode it fixed64 (does not work for int)
		BlockID:   CanonicalizeBlockID(vote.BlockID),
		Timestamp: vote.Timestamp,
		ChainID:   chainID,
	}
}

// CanonicalTime can be used to stringify time in a canonical way.
func CanonicalTime(t time.Time) string {
	// Note that sending time over amino resets it to
	// local time, we need to force UTC here, so the
	// signatures match
	return tmtime.Canonical(t).Format(TimeFormat)
}
