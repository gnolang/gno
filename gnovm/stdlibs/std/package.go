package std

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gnovm/stdlibs/std",
	"tm",
	amino.GetCallersDirname(),
).
	WithDependencies(
		abci.Package,
	).
	WithTypes(
		gnoEventAttribute{},
		gnoEvent{},
	))
