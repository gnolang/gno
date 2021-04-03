package ed25519

import (
	"github.com/tendermint/go-amino-x"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/tendermint/classic/crypto/ed25519",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	PubKeyEd25519{}, "PubKeyEd25519",
	PrivKeyEd25519{}, "PrivKeyEd25519",
))
