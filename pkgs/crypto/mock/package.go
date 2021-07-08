package mock

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/crypto/mock",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	PubKeyMock{}, "PubKeyMock",
	PrivKeyMock{}, "PrivKeyMock",
))
