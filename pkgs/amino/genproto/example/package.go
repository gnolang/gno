package main

import (
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/amino/genproto/example/submodule"
)

var Package = amino.RegisterPackage(
	amino.NewPackage(
		"main",
		"main",
		amino.GetCallersDirname(),
	).WithP3GoPkgPath(
		"github.com/gnolang/gno/pkgs/amino/genproto/example/pb",
	).WithDependencies(
		submodule.Package,
	).WithTypes(
		StructA{},
		StructB{},
	),
)
