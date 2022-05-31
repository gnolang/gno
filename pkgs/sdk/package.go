package sdk

import (
	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/sdk",
	"tm",
	amino.GetCallersDirname(),
).
	WithDependencies(
		abci.Package,
	).
	WithTypes(
		Result{},
	))
