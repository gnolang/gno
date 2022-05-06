package hd

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/crypto/hd",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	BIP44Params{}, "Bip44Params",
))
