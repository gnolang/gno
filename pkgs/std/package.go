package std

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/std",
	"std",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	&BaseAccount{}, "BaseAccount",
))
