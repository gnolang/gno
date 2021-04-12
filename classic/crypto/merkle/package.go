package merkle

import (
	"github.com/tendermint/go-amino-x"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/tendermint/classic/crypto/merkle",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	ProofOp{},
	Proof{},
))
