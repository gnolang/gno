package remote

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

// Register Amino package with remote signer message types.
var Package = amino.RegisterPackage(
	amino.NewPackage(
		"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote",
		"tm.remotesigner",
		amino.GetCallersDirname(),
	).
		WithDependencies().
		WithTypes(
			&PubKeyRequest{},
			&PubKeyResponse{},
			&SignRequest{},
			&SignResponse{},
		))
