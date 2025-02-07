package gnovm

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gnovm",
	"gnovm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	// MemFile/MemPackage
	MemFile{}, "MemFile",
	MemPackage{}, "MemPackage",
))
