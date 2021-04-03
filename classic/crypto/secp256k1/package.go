package secp256k1

import (
	"github.com/tendermint/go-amino-x"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/tendermint/classic/crypto/secp256k1",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	PubKeySecp256k1{}, "PubKeySecp256k1",
	PrivKeySecp256k1{}, "PrivKeySecp256k1",
))
