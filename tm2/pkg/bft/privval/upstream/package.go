package upstream

import (
	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

// ProtocolVersion is the upstream Tendermint privval protocol dialect
// this package speaks. The wire types in upstreampb (Vote, Proposal,
// canonical sign-bytes layout) are hardcoded to v0.34 — the same value
// operators put in tmkms.toml's [[validator]].protocol_version. We
// expose it so config validation can refuse any other version (forward
// versions add fields that change canonical sign-bytes; we don't want
// to silently sign for a different dialect).
const ProtocolVersion = "v0.34"

// Package registers the upstream-shaped operational types with amino so
// they can be marshaled via the codec. Registration drives the binary
// encoder's per-field FieldOptions (binary:"varint" tags), which is how
// these types produce upstream-Tendermint-compatible wire bytes.
//
// NOTE: privval socket-protocol messages (Message, PubKeyRequest, etc.)
// are NOT amino-encoded. They live in the upstreampb sibling package as
// protoc-generated types, used directly via google.golang.org/protobuf
// for wire I/O. Mirrors how cometbft/privval uses cometbft/proto.
var Package = pkg.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream",
	"upstream",
	pkg.GetCallersDirname(),
).WithTypes(
	// Operational types — exposed as amino-encoded structs for byte-compat
	// verification and as Go-friendly bridges in places where amino is the
	// natural codec. Wire-layer code in privval/upstream/listener uses the
	// upstreampb protoc-generated equivalents, not these.
	Vote{},
	Proposal{},
	BlockID{},
	PartSetHeader{},
)
