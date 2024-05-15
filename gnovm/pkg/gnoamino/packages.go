package gnoamino

import "github.com/gnolang/gno/tm2/pkg/amino"

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gnovm/pkg/gnoamino",
	"gno",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	&TypedValueWrapper{},
))
