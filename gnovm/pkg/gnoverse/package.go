package gnoverse

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/vm"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gnovm/pkg/gnoverse",
	"gnoverse",
	amino.GetCallersDirname(),
).
	WithDependencies(
		bank.Package,
		vm.Package,
	).
	WithTypes(
		&Sandbox{}, "Sandbox",
	))
