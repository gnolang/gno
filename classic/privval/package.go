package privval

import (
	"github.com/tendermint/go-amino-x"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/tendermint/classic/privval",
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
