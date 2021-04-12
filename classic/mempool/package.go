package mempool

import (
	"github.com/tendermint/go-amino-x"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/tendermint/classic/mempool",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	&TxMessage{},
))
