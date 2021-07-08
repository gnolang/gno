package merkle

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/crypto/merkle",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	ProofOp{},
	Proof{},
))
