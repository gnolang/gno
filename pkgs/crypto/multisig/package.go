package multisig

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/crypto/multisig",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	PubKeyMultisigThreshold{}, "PubKeyMultisig",
))
