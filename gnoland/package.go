package gnoland

import (
	"github.com/gnolang/gno/gnoland/types"
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gnoland/types",
	"gno",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	&types.GnoAccount{}, "Account",
	types.GnoGenesisState{}, "GenesisState",
))
