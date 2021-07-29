package pex

import (
	"github.com/tendermint/go-amino-x"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/tendermint/classic/p2p/pex",
	"p2p", // keep short, do not change.
	amino.GetCallersDirname(),
).
	WithDependencies(
	// NA
	).
	WithTypes(
		// NOTE: Keep the names short.
		&PexRequestMessage{}, "PexRequest",
		&PexAddrsMessage{}, "PexAddrs",
	))
