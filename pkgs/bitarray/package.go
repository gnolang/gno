package bitarray

import "github.com/gnolang/gno/pkgs/amino"

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/bitarray",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	BitArray{}, "BitArray",
))
