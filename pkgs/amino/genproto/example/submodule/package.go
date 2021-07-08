package submodule

import (
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/amino/genproto/example/submodule2"
)

var Package = amino.RegisterPackage(
	amino.NewPackage(
		"github.com/gnolang/gno/pkgs/amino/genproto/example/submodule",
		"submodule",
		amino.GetCallersDirname(),
	).WithDependencies(
		submodule2.Package,
	).WithTypes(
		StructSM{},
	),
)
