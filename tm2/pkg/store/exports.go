package store

import (
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// Import cosmos-sdk/types/store.go for convenience.
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
)

var (
	PruneNothing          = types.PruneNothing
	PruneEverything       = types.PruneEverything
	PruneSyncable         = types.PruneSyncable
	PrefixIterator        = types.PrefixIterator
	ReversePrefixIterator = types.ReversePrefixIterator
	NewStoreKey           = types.NewStoreKey
)
