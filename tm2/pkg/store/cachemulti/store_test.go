package cachemulti

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// panicOnWriteStore wraps a real store but panics on Write(), simulating
// a sub-store failure (e.g., LMDB write error) during MultiWrite().
// Checkpoint methods delegate to the underlying store.
type panicOnWriteStore struct {
	types.Store
}

func (ps panicOnWriteStore) Write() {
	panic("simulated write failure")
}

func (ps panicOnWriteStore) CacheWrap() types.Store {
	return panicOnWriteStore{Store: ps.Store.CacheWrap()}
}

// Delegate checkpoint methods to the underlying cache store.
func (ps panicOnWriteStore) Checkpoint() {
	ps.Store.(interface{ Checkpoint() }).Checkpoint()
}

func (ps panicOnWriteStore) HasCheckpoint() bool {
	return ps.Store.(interface{ HasCheckpoint() bool }).HasCheckpoint()
}

func (ps panicOnWriteStore) WriteCheckpoint() {
	ps.Store.(interface{ WriteCheckpoint() }).WriteCheckpoint()
}

// TestMultiWritePartialFailureDoublePanic demonstrates bug C5:
// If MultiWrite() panics mid-way through sub-stores, the defer
// that calls WriteCheckpoint() double-panics because stores that
// already committed had their checkpoints cleared by clear().
//
// Sequence:
//  1. Checkpoint() snapshots all sub-stores
//  2. MultiWrite() iterates sub-stores: store A commits (clears checkpoint),
//     store B panics
//  3. Defer sees HasCheckpoint()==true (B still has one), calls WriteCheckpoint()
//  4. WriteCheckpoint() on store A panics: "WriteCheckpoint called without Checkpoint"
//
// This hides the original error and leaves multi-store in torn state.
func TestMultiWritePartialFailureDoublePanic(t *testing.T) {
	// Go map iteration order is non-deterministic. The bug only triggers
	// when the normal store commits before the panicking store is reached.
	// Run enough iterations to hit both orderings.
	bugTriggered := false

	for range 50 {
		memA := dbadapter.Store{DB: memdb.NewMemDB()}
		memA.Set(nil, []byte("keyA"), []byte("origA"))

		memB := dbadapter.Store{DB: memdb.NewMemDB()}
		memB.Set(nil, []byte("keyB"), []byte("origB"))

		keyA := types.NewStoreKey("storeA")
		keyB := types.NewStoreKey("storeB")

		stores := map[types.StoreKey]types.Store{
			keyA: memA,
			keyB: panicOnWriteStore{Store: memB},
		}
		keys := map[string]types.StoreKey{
			"storeA": keyA,
			"storeB": keyB,
		}
		cms := New(stores, keys)

		// Ante handler writes.
		cms.GetStore(keyA).Set(nil, []byte("keyA"), []byte("anteA"))
		cms.GetStore(keyB).Set(nil, []byte("keyB"), []byte("anteB"))

		// Snapshot ante state.
		cms.Checkpoint()

		// Msg handler writes.
		cms.GetStore(keyA).Set(nil, []byte("keyA"), []byte("msgA"))
		cms.GetStore(keyB).Set(nil, []byte("keyB"), []byte("msgB"))

		// MultiWrite panics because store B's Write() fails.
		func() {
			defer func() { recover() }()
			cms.MultiWrite()
		}()

		// Check which ordering we got by inspecting whether store A committed.
		aCommitted := string(memA.Get(nil, []byte("keyA"))) == "msgA"

		if !aCommitted {
			// Store B panicked first, A never wrote. Both checkpoints
			// are intact. WriteCheckpoint works fine — no bug in this ordering.
			continue
		}

		// Store A committed before B panicked.
		// This is the torn state: A flushed msg writes, B did not.
		assert.Equal(t, []byte("msgA"), memA.Get(nil, []byte("keyA")),
			"store A flushed msg writes to parent (torn state)")
		assert.Equal(t, []byte("origB"), memB.Get(nil, []byte("keyB")),
			"store B never flushed (panic prevented it)")

		// HasCheckpoint returns true because B still has its checkpoint.
		require.True(t, cms.HasCheckpoint(),
			"HasCheckpoint should be true — store B's checkpoint survived the panic")

		// BUG: WriteCheckpoint() double-panics because store A's checkpoint
		// was cleared by clear() during its successful Write().
		assert.Panics(t, func() {
			cms.WriteCheckpoint()
		}, "WriteCheckpoint double-panics: store A lost its checkpoint during partial MultiWrite")

		bugTriggered = true
		break
	}

	require.True(t, bugTriggered,
		"failed to trigger the bug after 50 attempts — map ordering never put the normal store first")
}
