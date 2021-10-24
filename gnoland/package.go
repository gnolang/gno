package gnoland

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gnoland",
	"gno",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	&GnoAccount{}, "Account",
	GnoGenesisState{}, "GenesisState",
))
