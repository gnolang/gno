package privval

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/privval",
	"tm.remotesigner",
	amino.GetCallersDirname(),
).
	WithDependencies().
	WithTypes(

		// Remote Signer
		&PubKeyRequest{},
		&PubKeyResponse{},
		&SignVoteRequest{},
		&SignedVoteResponse{},
		&SignProposalRequest{},
		&SignedProposalResponse{},
		&PingRequest{},
		&PingResponse{},
	))
