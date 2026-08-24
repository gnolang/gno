package upstream

import (
	"fmt"
	"math"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// FromTM2Vote converts a chain-internal types.Vote to its upstream-compatible
// shape. ValidatorAddress is the raw 20 bytes of the bech32-encoded tm2
// Address; rounds and indices narrow to int32 (overflow panics — these are
// consensus-shaped values that cannot legitimately exceed int32 range).
func FromTM2Vote(v *types.Vote) *Vote {
	if v == nil {
		return nil
	}
	return &Vote{
		Type:             v.Type,
		Height:           v.Height,
		Round:            int32From(v.Round, "Vote.Round"),
		BlockID:          FromTM2BlockID(v.BlockID),
		Timestamp:        v.Timestamp,
		ValidatorAddress: append([]byte(nil), v.ValidatorAddress[:]...),
		ValidatorIndex:   int32From(v.ValidatorIndex, "Vote.ValidatorIndex"),
		Signature:        v.Signature,
	}
}

// ToTM2Vote converts an upstream-shape Vote back to a chain-internal
// types.Vote. ValidatorAddress bytes must be exactly crypto.AddressSize
// (20); otherwise the conversion panics — an upstream peer sending a
// malformed address is a protocol violation we surface immediately rather
// than silently truncate.
func ToTM2Vote(v *Vote) *types.Vote {
	if v == nil {
		return nil
	}
	return &types.Vote{
		Type:             v.Type,
		Height:           v.Height,
		Round:            int(v.Round),
		BlockID:          ToTM2BlockID(v.BlockID),
		Timestamp:        v.Timestamp,
		ValidatorAddress: addressFromBytes(v.ValidatorAddress),
		ValidatorIndex:   int(v.ValidatorIndex),
		Signature:        v.Signature,
	}
}

// FromTM2Proposal converts a chain-internal types.Proposal to upstream shape.
func FromTM2Proposal(p *types.Proposal) *Proposal {
	if p == nil {
		return nil
	}
	return &Proposal{
		Type:      p.Type,
		Height:    p.Height,
		Round:     int32From(p.Round, "Proposal.Round"),
		POLRound:  int32From(p.POLRound, "Proposal.POLRound"),
		BlockID:   FromTM2BlockID(p.BlockID),
		Timestamp: p.Timestamp,
		Signature: p.Signature,
	}
}

// ToTM2Proposal converts an upstream-shape Proposal back to chain-internal.
func ToTM2Proposal(p *Proposal) *types.Proposal {
	if p == nil {
		return nil
	}
	return &types.Proposal{
		Type:      p.Type,
		Height:    p.Height,
		Round:     int(p.Round),
		POLRound:  int(p.POLRound),
		BlockID:   ToTM2BlockID(p.BlockID),
		Timestamp: p.Timestamp,
		Signature: p.Signature,
	}
}

func FromTM2BlockID(b types.BlockID) BlockID {
	return BlockID{
		Hash:          b.Hash,
		PartSetHeader: FromTM2PartSetHeader(b.PartsHeader),
	}
}

func ToTM2BlockID(b BlockID) types.BlockID {
	return types.BlockID{
		Hash:        b.Hash,
		PartsHeader: ToTM2PartSetHeader(b.PartSetHeader),
	}
}

func FromTM2PartSetHeader(p types.PartSetHeader) PartSetHeader {
	if p.Total < 0 {
		panic(fmt.Sprintf("PartSetHeader.Total is negative (%d); upstream uint32 cannot represent it", p.Total))
	}
	if int64(p.Total) > math.MaxUint32 {
		panic(fmt.Sprintf("PartSetHeader.Total (%d) exceeds uint32 range", p.Total))
	}
	return PartSetHeader{
		Total: uint32(p.Total),
		Hash:  p.Hash,
	}
}

func ToTM2PartSetHeader(p PartSetHeader) types.PartSetHeader {
	return types.PartSetHeader{
		Total: int(p.Total),
		Hash:  p.Hash,
	}
}

// int32From narrows a Go int (consensus round/index value) to int32. Panics
// on out-of-range; consensus values cannot legitimately exceed int32 in
// any tm2 deployment — overflow here indicates either a buggy caller or a
// malicious peer, and silent truncation would corrupt sign-bytes.
func int32From(v int, name string) int32 {
	if v < math.MinInt32 || v > math.MaxInt32 {
		panic(fmt.Sprintf("%s value %d out of int32 range", name, v))
	}
	return int32(v)
}

// addressFromBytes constructs a tm2 crypto.Address from a raw byte slice.
// Length must match crypto.AddressSize; mismatch panics.
func addressFromBytes(b []byte) crypto.Address {
	if len(b) != crypto.AddressSize {
		panic(fmt.Sprintf("address has length %d, expected %d", len(b), crypto.AddressSize))
	}
	var addr crypto.Address
	copy(addr[:], b)
	return addr
}
