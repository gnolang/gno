package mock

import (
	"github.com/tendermint/go-amino-x"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/tendermint/classic/crypto/mock",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	PubKeyMock{}, "PubKeyMock",
	PrivKeyMock{}, "PrivKeyMock",
))
