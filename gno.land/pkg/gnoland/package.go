package gnoland

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gno.land/pkg/gnoland",
	"gno",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	&GnoAccount{}, "Account",
	GnoGenesisState{}, "GenesisState",
))
