package upstream

import (
	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

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
