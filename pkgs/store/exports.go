package store

import (
	"github.com/gnolang/gno/pkgs/store/types"
	stypes "github.com/gnolang/gno/pkgs/store/types"
)

// Import cosmos-sdk/types/store.go for convenience.
// nolint
type (
	PruningOptions   = types.PruningOptions
	Store            = types.Store
	Committer        = types.Committer
	CommitStore      = types.CommitStore
	MultiStore       = types.MultiStore
	CommitMultiStore = types.CommitMultiStore
	KVPair           = types.KVPair
	Iterator         = types.Iterator
	CommitID         = types.CommitID
	StoreKey         = types.StoreKey
	Queryable        = types.Queryable
	Gas              = stypes.Gas
	GasMeter         = types.GasMeter
	GasConfig        = stypes.GasConfig
)

// nolint - reexport
var (
	PruneNothing    = types.PruneNothing
	PruneEverything = types.PruneEverything
	PruneSyncable   = types.PruneSyncable
)
