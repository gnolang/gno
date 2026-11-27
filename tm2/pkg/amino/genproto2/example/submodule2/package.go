package submodule2

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(
	amino.NewPackage(
		"github.com/gnolang/gno/tm2/pkg/amino/genproto2/example/submodule2",
		"submodule2",
		amino.GetCallersDirname(),
	).WithTypes(
		StructSM2{},
	),
)
