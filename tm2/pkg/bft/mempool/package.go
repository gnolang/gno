package mempool

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/bft/mempool",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	&TxMessage{},
))
