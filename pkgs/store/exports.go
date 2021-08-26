package store

import (
	"github.com/gnolang/gno/pkgs/store/types"
)

// Import cosmos-sdk/types/store.go for convenience.
// nolint
type (
	PruningOptions         = types.PruningOptions
	Store                  = types.Store
	Committer              = types.Committer
	CommitStore            = types.CommitStore
	MultiStore             = types.MultiStore
	CommitMultiStore       = types.CommitMultiStore
	CommitStoreConstructor = types.CommitStoreConstructor
	KVPair                 = types.KVPair
	Iterator               = types.Iterator
	CommitID               = types.CommitID
	StoreKey               = types.StoreKey
	StoreOptions           = types.StoreOptions
	Queryable              = types.Queryable
	Gas                    = types.Gas
	GasMeter               = types.GasMeter
	GasConfig              = types.GasConfig
)

// nolint - reexport
var (
	PruneNothing        = types.PruneNothing
	PruneEverything     = types.PruneEverything
	PruneSyncable       = types.PruneSyncable
	NewInfiniteGasMeter = types.NewInfiniteGasMeter
	NewGasMeter         = types.NewGasMeter
	DefaultGasConfig    = types.DefaultGasConfig
)
