package std

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gno/stdlibs/std",
	"tm",
	amino.GetCallersDirname(),
).
	WithDependencies(
		abci.Package,
	).
	WithTypes(
		GnoEventAttribute{},
		GnoEvent{},
	))
