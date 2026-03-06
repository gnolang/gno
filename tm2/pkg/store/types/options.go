package types

// (Global) Store options are used to construct new stores.
type StoreOptions struct {
	PruningOptions
	Immutable bool
}

// PruningOptions specifies how old states will be deleted over time where
// KeepRecent can be used with KeepEvery to create a pruning "strategy".
type PruningOptions struct {
	// How many old versions we hold onto.
	// A value of 0 means keep no recent states.
	KeepRecent int64
	// This is the distance between state-sync waypoint states to be stored.
	// See https://github.com/tendermint/tendermint/issues/828
	// A value of 1 means store every state.
	// A value of 0 means store no waypoints. (node cannot assist in state-sync)
	// By default this value should be set the same across all nodes,
	// so that nodes can know the waypoints their peers store.
	KeepEvery int64
}

func NewPruningOptions(keepRecent, keepEvery int64) PruningOptions {
	return PruningOptions{
		KeepRecent: keepRecent,
		KeepEvery:  keepEvery,
	}
}

// default pruning strategies
var (
	// PruneEverything means all saved states will be deleted, storing only the current state
	PruneEverything = NewPruningOptions(0, 0)
	// PruneNothing means all historic states will be saved, nothing will be deleted
	PruneNothing = NewPruningOptions(0, 1)
	// PruneSyncable means only those states not needed for state syncing will be deleted.
	// Assuming 3s block times, and a span of 3.5w, ~705600 blocks should be kept.
	PruneSyncable = NewPruningOptions(705600, 10)
)

type PruneStrategy string

const (
	PruneEverythingStrategy PruneStrategy = "everything"
	PruneNothingStrategy    PruneStrategy = "nothing"
	PruneSyncableStrategy   PruneStrategy = "syncable"
)

// Options returns the corresponding prune options.
// If the pruning strategy is invalid, defaults to no pruning.
func (s PruneStrategy) Options() PruningOptions {
	switch s {
	case PruneEverythingStrategy:
		return PruneEverything
	case PruneNothingStrategy:
		return PruneNothing
	default:
		return PruneSyncable
	}
}
