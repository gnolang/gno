package conn

import (
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/p2p/conn",
	"p2p", // keep short, do not change.
	amino.GetCallersDirname(),
).
	WithDependencies(
	// NA
	).
	WithTypes(

		// NOTE: Keep the names short.
		pkg.Type{
			Type: reflect.TypeFor[PacketPing](),
			Name: "Ping",
		},
		pkg.Type{
			Type: reflect.TypeFor[PacketPong](),
			Name: "Pong",
		},
		pkg.Type{
			Type: reflect.TypeFor[PacketMsg](),
			Name: "Msg",
		},
	))
