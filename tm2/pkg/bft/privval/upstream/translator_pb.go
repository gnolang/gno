package upstream

// translator_pb.go: bridge between upstreampb (protoc-generated) types
// and tm2's chain-internal types. Used at the privval-wire boundary —
// the listener decodes upstreampb messages from the wire and converts
// to tm2 types for the application layer (and vice versa).
//
// Mirrors how cometbft/privval converts between cometbft/proto types
// and cometbft/types domain types in signer_requestHandler.go.

import (
	"fmt"
	"math"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ---- PublicKey ↔ crypto.PubKey -------------------------------------------

// PubKeyToProto converts a tm2 crypto.PubKey to the upstreampb.PublicKey
// oneof. Returns an error for unsupported key types — only ed25519 and
// secp256k1 are mapped, matching upstream Tendermint v0.34's PublicKey
// oneof.
func PubKeyToProto(pk crypto.PubKey) (*upstreampb.PublicKey, error) {
	switch k := pk.(type) {
	case ed25519.PubKeyEd25519:
		return &upstreampb.PublicKey{Sum: &upstreampb.PublicKey_Ed25519{Ed25519: k[:]}}, nil
	case secp256k1.PubKeySecp256k1:
		return &upstreampb.PublicKey{Sum: &upstreampb.PublicKey_Secp256K1{Secp256K1: k[:]}}, nil
	default:
		return nil, fmt.Errorf("upstream.PubKeyToProto: unsupported pubkey type %T", pk)
	}
}

// PubKeyFromProto returns the tm2 crypto.PubKey for an upstreampb.PublicKey
// oneof, dispatching on the populated branch.
func PubKeyFromProto(p *upstreampb.PublicKey) (crypto.PubKey, error) {
	if p == nil {
		return nil, fmt.Errorf("upstream.PubKeyFromProto: nil PublicKey")
	}
	switch sum := p.Sum.(type) {
	case *upstreampb.PublicKey_Ed25519:
		if len(sum.Ed25519) != ed25519.PubKeyEd25519Size {
			return nil, fmt.Errorf("upstream.PubKeyFromProto: ed25519 length %d, expected %d", len(sum.Ed25519), ed25519.PubKeyEd25519Size)
		}
		var pk ed25519.PubKeyEd25519
		copy(pk[:], sum.Ed25519)
		return pk, nil
	case *upstreampb.PublicKey_Secp256K1:
		if len(sum.Secp256K1) != secp256k1.PubKeySecp256k1Size {
			return nil, fmt.Errorf("upstream.PubKeyFromProto: secp256k1 length %d, expected %d", len(sum.Secp256K1), secp256k1.PubKeySecp256k1Size)
		}
		var pk secp256k1.PubKeySecp256k1
		copy(pk[:], sum.Secp256K1)
		return pk, nil
	default:
		return nil, fmt.Errorf("upstream.PubKeyFromProto: empty or unknown PublicKey sum")
	}
}

// ---- Vote ↔ upstreampb.Vote ----------------------------------------------

// VoteToProto converts a tm2 types.Vote to upstreampb.Vote.
func VoteToProto(v *types.Vote) (*upstreampb.Vote, error) {
	if v == nil {
		return nil, nil
	}
	bid, err := blockIDToProto(v.BlockID)
	if err != nil {
		return nil, err
	}
	round, err := narrowInt32(v.Round, "Vote.Round")
	if err != nil {
		return nil, err
	}
	idx, err := narrowInt32(v.ValidatorIndex, "Vote.ValidatorIndex")
	if err != nil {
		return nil, err
	}
	return &upstreampb.Vote{
		Type:             upstreampb.SignedMsgType(v.Type),
		Height:           v.Height,
		Round:            round,
		BlockId:          bid,
		Timestamp:        timestamppb.New(v.Timestamp),
		ValidatorAddress: append([]byte(nil), v.ValidatorAddress[:]...),
		ValidatorIndex:   idx,
		Signature:        append([]byte(nil), v.Signature...),
	}, nil
}

// VoteFromProto converts upstreampb.Vote to tm2 types.Vote.
func VoteFromProto(v *upstreampb.Vote) (*types.Vote, error) {
	if v == nil {
		return nil, nil
	}
	bid, err := blockIDFromProto(v.BlockId)
	if err != nil {
		return nil, err
	}
	addr, err := addressFromProtoBytes(v.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	var ts time.Time
	if v.Timestamp != nil {
		ts = v.Timestamp.AsTime()
	}
	return &types.Vote{
		Type:             types.SignedMsgType(v.Type),
		Height:           v.Height,
		Round:            int(v.Round),
		BlockID:          bid,
		Timestamp:        ts,
		ValidatorAddress: addr,
		ValidatorIndex:   int(v.ValidatorIndex),
		Signature:        append([]byte(nil), v.Signature...),
	}, nil
}

// ---- Proposal ↔ upstreampb.Proposal --------------------------------------

func ProposalToProto(p *types.Proposal) (*upstreampb.Proposal, error) {
	if p == nil {
		return nil, nil
	}
	bid, err := blockIDToProto(p.BlockID)
	if err != nil {
		return nil, err
	}
	round, err := narrowInt32(p.Round, "Proposal.Round")
	if err != nil {
		return nil, err
	}
	pol, err := narrowInt32(p.POLRound, "Proposal.POLRound")
	if err != nil {
		return nil, err
	}
	return &upstreampb.Proposal{
		Type:      upstreampb.SignedMsgType(p.Type),
		Height:    p.Height,
		Round:     round,
		PolRound:  pol,
		BlockId:   bid,
		Timestamp: timestamppb.New(p.Timestamp),
		Signature: append([]byte(nil), p.Signature...),
	}, nil
}

func ProposalFromProto(p *upstreampb.Proposal) (*types.Proposal, error) {
	if p == nil {
		return nil, nil
	}
	bid, err := blockIDFromProto(p.BlockId)
	if err != nil {
		return nil, err
	}
	var ts time.Time
	if p.Timestamp != nil {
		ts = p.Timestamp.AsTime()
	}
	return &types.Proposal{
		Type:      types.SignedMsgType(p.Type),
		Height:    p.Height,
		Round:     int(p.Round),
		POLRound:  int(p.PolRound),
		BlockID:   bid,
		Timestamp: ts,
		Signature: append([]byte(nil), p.Signature...),
	}, nil
}

// ---- Internal helpers ----------------------------------------------------

func blockIDToProto(b types.BlockID) (*upstreampb.BlockID, error) {
	psh, err := partSetHeaderToProto(b.PartsHeader)
	if err != nil {
		return nil, err
	}
	return &upstreampb.BlockID{
		Hash:          append([]byte(nil), b.Hash...),
		PartSetHeader: psh,
	}, nil
}

func blockIDFromProto(b *upstreampb.BlockID) (types.BlockID, error) {
	if b == nil {
		return types.BlockID{}, nil
	}
	psh := partSetHeaderFromProto(b.PartSetHeader)
	return types.BlockID{
		Hash:        append([]byte(nil), b.Hash...),
		PartsHeader: psh,
	}, nil
}

func partSetHeaderToProto(p types.PartSetHeader) (*upstreampb.PartSetHeader, error) {
	if p.Total < 0 || int64(p.Total) > math.MaxUint32 {
		return nil, fmt.Errorf("PartSetHeader.Total %d out of canonical uint32 range", p.Total)
	}
	return &upstreampb.PartSetHeader{
		Total: uint32(p.Total),
		Hash:  append([]byte(nil), p.Hash...),
	}, nil
}

func partSetHeaderFromProto(p *upstreampb.PartSetHeader) types.PartSetHeader {
	if p == nil {
		return types.PartSetHeader{}
	}
	return types.PartSetHeader{
		Total: int(p.Total),
		Hash:  append([]byte(nil), p.Hash...),
	}
}

func addressFromProtoBytes(b []byte) (crypto.Address, error) {
	if len(b) != crypto.AddressSize {
		return crypto.Address{}, fmt.Errorf("validator address length %d, expected %d", len(b), crypto.AddressSize)
	}
	var a crypto.Address
	copy(a[:], b)
	return a, nil
}

func narrowInt32(v int, name string) (int32, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, fmt.Errorf("%s value %d out of int32 range", name, v)
	}
	return int32(v), nil
}
