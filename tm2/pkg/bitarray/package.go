package bitarray

import "github.com/gnolang/gno/tm2/pkg/amino"

var dirname = amino.GetCallersDirname()

func init() {
	panic(dirname)
}

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/bitarray",
	"tm",
	dirname,
).WithDependencies().WithTypes(
	BitArray{}, "BitArray",
))
