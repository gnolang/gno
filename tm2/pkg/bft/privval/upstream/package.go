package upstream

import (
	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

// Package registers the upstream-shaped privval types with amino so they
// can be marshaled via the codec. Registration is required for the binary
// encoder to walk the struct fields with the right FieldOptions
// (binary:"varint" tags).
var Package = pkg.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream",
	"upstream",
	pkg.GetCallersDirname(),
).WithTypes(
	// Operational types
	Vote{},
	Proposal{},
	BlockID{},
	PartSetHeader{},

	// Privval socket-protocol messages
	PubKeyRequest{},
	PubKeyResponse{},
	SignVoteRequest{},
	SignedVoteResponse{},
	SignProposalRequest{},
	SignedProposalResponse{},
	PingRequest{},
	PingResponse{},
	RemoteSignerError{},
)
