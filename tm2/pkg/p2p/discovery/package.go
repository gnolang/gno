package discovery

import (
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/p2p/discovery",
	"p2p",
	amino.GetCallersDirname(),
).
	WithDependencies(
	// NA
	).
	WithTypes(
		// NOTE: Keep the names short.
		pkg.Type{
			Type: reflect.TypeOf(Request{}),
			Name: "Request",
		},
		pkg.Type{
			Type: reflect.TypeOf(Response{}),
			Name: "Response",
		},
	),
)
