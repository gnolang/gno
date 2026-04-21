package gnoland

import (
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gno.land/pkg/gnoland",
	"gno",
	amino.GetCallersDirname(),
).WithDependencies(
	std.Package,
	auth.Package,
	bank.Package,
	vm.Package,
).WithTypes(
	&GnoAccount{}, "Account",
	&GnoSessionAccount{}, "SessionAccount",
	GnoGenesisState{}, "GenesisState",
	TxWithMetadata{}, "TxWithMetadata",
	GnoTxMetadata{}, "GnoTxMetadata",
))
