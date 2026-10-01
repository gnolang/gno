package params

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/sdk/params",
	"params",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	Param{}, "Param",
))
