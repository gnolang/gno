package std

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gnovm/pkg/std",
	"std",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	// MemFile/MemPackage
	MemFile{}, "MemFile",
	MemPackage{}, "MemPackage",
))
