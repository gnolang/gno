package types

// (Global) Store options are used to construct new stores.
type StoreOptions struct {
	PruningOptions
	LazyLoad  bool
	Immutable bool
}

// PruningOptions specifies how old states will be deleted over time where
// KeepRecent can be used with KeepEvery to create a pruning "strategy".
type PruningOptions struct {
	// How many old versions we hold onto.
	// A value of 0 means keep no recent states.
	KeepRecent int64
	// This is the distance between state-sync waypoint states to be stored.
	// See https://github.com/tendermint/classic/issues/828
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
	// PruneSyncable means only those states not needed for state syncing will be deleted (keeps last 100 + every 10000th)
	PruneSyncable = NewPruningOptions(100, 10000)
)
