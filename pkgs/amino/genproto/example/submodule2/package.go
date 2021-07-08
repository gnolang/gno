package submodule2

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(
	amino.NewPackage(
		"github.com/gnolang/gno/pkgs/amino/genproto/example/submodule2",
		"submodule2",
		amino.GetCallersDirname(),
	).WithTypes(
		StructSM2{},
	),
)
