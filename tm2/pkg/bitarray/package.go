package bitarray

import "github.com/gnolang/gno/tm2/pkg/amino"

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/bitarray",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	BitArray{}, "BitArray",
))
