package gnoverse

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/store"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gnovm/pkg/gnoverse",
	"gnoverse",
	amino.GetCallersDirname(),
).
	WithDependencies(
		bank.Package,
		db.Package,
		vm.Package,
		store.Package,
		auth.Package,
	).
	WithTypes(
		&Sandbox{}, "Sandbox",
	))
