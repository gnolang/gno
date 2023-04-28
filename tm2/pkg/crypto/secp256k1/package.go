package secp256k1

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	PubKeySecp256k1{}, "PubKeySecp256k1",
	PrivKeySecp256k1{}, "PrivKeySecp256k1",
))
