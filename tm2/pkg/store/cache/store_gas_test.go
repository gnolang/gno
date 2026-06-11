package cache_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// Gas charging at the cache.Store boundary (the only charging point since
// gas.Store was removed). All assertions are relative to the GasConfig
// fields, never absolute calibrated numbers.

var gasTestConfig = types.GasConfig{
	ReadCostFlat:     100,
	ReadCostPerByte:  1,
	WriteCostFlat:    200,
	WriteCostPerByte: 2,
	DeleteCost:       150,
}

type stubDepth struct{ d int64 }

func (s stubDepth) ExpectedDepth() int64 { return s.d }

func newGasCtx(cfg types.GasConfig) (*types.GasContext, types.GasMeter) {
	m := types.NewGasMeter(1 << 40)
	return &types.GasContext{Meter: m, Config: cfg}, m
}

func delta(m types.GasMeter, f func()) types.Gas {
	before := m.GasConsumed()
	f()
	return m.GasConsumed() - before
}

func TestGas_GetFlat(t *testing.T) {
	t.Parallel()
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	parent.Set(nil, []byte("k"), []byte("hello"))
	cs := cache.New(parent)
	gctx, m := newGasCtx(gasTestConfig)
	cfg := gasTestConfig

	// Miss: flat read + per-byte on the fetched value.
	miss := delta(m, func() { cs.Get(gctx, []byte("k")) })
	require.Equal(t, cfg.ReadCostFlat+cfg.ReadCostPerByte*5, miss)

	// Hit (now cached): free.
	require.Zero(t, delta(m, func() { cs.Get(gctx, []byte("k")) }))

	// Absent-key miss charges the flat cost only; the cached absence is
	// free on the second read.
	require.Equal(t, cfg.ReadCostFlat, delta(m, func() { cs.Get(gctx, []byte("absent")) }))
	require.Zero(t, delta(m, func() { cs.Get(gctx, []byte("absent")) }))
}

func TestGas_GetDepth(t *testing.T) {
	t.Parallel()
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	parent.Set(nil, []byte("k"), []byte("hello"))
	cs := cache.New(parent)
	cs.SetDepthEstimator(stubDepth{d: 7})
	gctx, m := newGasCtx(gasTestConfig)
	cfg := gasTestConfig

	miss := delta(m, func() { cs.Get(gctx, []byte("k")) })
	require.Equal(t, 7*cfg.ReadCostFlat+cfg.ReadCostPerByte*5, miss)
	require.Zero(t, delta(m, func() { cs.Get(gctx, []byte("k")) }))
}

func TestGas_MinDepthFloorsWithoutEstimator(t *testing.T) {
	t.Parallel()
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	cs := cache.New(parent)
	cfg := gasTestConfig
	cfg.MinDepth = 5
	gctx, m := newGasCtx(cfg)

	// No estimator, but MinDepth floors the depth: the depth-charging path
	// applies even on a flat store.
	miss := delta(m, func() { cs.Get(gctx, []byte("absent")) })
	require.Equal(t, 5*cfg.ReadCostFlat, miss)
}

func TestGas_SetFlatAndDepth(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig

	cs := cache.New(dbadapter.Store{DB: memdb.NewMemDB()})
	gctx, m := newGasCtx(cfg)
	flat := delta(m, func() { cs.Set(gctx, []byte("k"), []byte("vvvv")) })
	require.Equal(t, cfg.WriteCostFlat+cfg.WriteCostPerByte*4, flat)

	cs2 := cache.New(dbadapter.Store{DB: memdb.NewMemDB()})
	cs2.SetDepthEstimator(stubDepth{d: 7})
	gctx2, m2 := newGasCtx(cfg)
	deep := delta(m2, func() { cs2.Set(gctx2, []byte("k"), []byte("vvvv")) })
	// Key bytes are never charged; value bytes are.
	require.Equal(t, 7*(cfg.ReadCostFlat+cfg.WriteCostFlat)+cfg.WriteCostPerByte*4, deep)
}

func TestGas_WriteDedupLastOpWins(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig
	cs := cache.New(dbadapter.Store{DB: memdb.NewMemDB()})
	gctx, m := newGasCtx(cfg)

	// Set, re-Set, Delete, Set again on one key: every op refunds the
	// previous charge, so the net equals a single charge for the last op.
	cs.Set(gctx, []byte("k"), []byte("v1"))
	cs.Set(gctx, []byte("k"), []byte("v2-longer"))
	cs.Delete(gctx, []byte("k"))
	cs.Set(gctx, []byte("k"), []byte("v3"))
	wantLast := cfg.WriteCostFlat + cfg.WriteCostPerByte*2
	require.Equal(t, wantLast, m.GasConsumed())

	// A second key is charged independently.
	cs.Set(gctx, []byte("k2"), []byte("x"))
	require.Equal(t, wantLast+cfg.WriteCostFlat+cfg.WriteCostPerByte*1, m.GasConsumed())
}

func TestGas_DeleteFlatAndDepth(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig

	cs := cache.New(dbadapter.Store{DB: memdb.NewMemDB()})
	gctx, m := newGasCtx(cfg)
	require.Equal(t, cfg.DeleteCost, delta(m, func() { cs.Delete(gctx, []byte("k")) }))

	cs2 := cache.New(dbadapter.Store{DB: memdb.NewMemDB()})
	cs2.SetDepthEstimator(stubDepth{d: 3})
	gctx2, m2 := newGasCtx(cfg)
	require.Equal(t, 3*(cfg.ReadCostFlat+cfg.WriteCostFlat),
		delta(m2, func() { cs2.Delete(gctx2, []byte("k")) }))
}

func TestGas_HasDelegatesToGet(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	parent.Set(nil, []byte("k"), []byte("vv"))
	cs := cache.New(parent)
	gctx, m := newGasCtx(cfg)

	// Has is a Get: full read gas on miss, free once cached. (HasCost is
	// unused by design — there is no cheaper existence check.)
	require.Equal(t, cfg.ReadCostFlat+cfg.ReadCostPerByte*2,
		delta(m, func() { cs.Has(gctx, []byte("k")) }))
	require.Zero(t, delta(m, func() { cs.Has(gctx, []byte("k")) }))
}

func TestGas_IterationUnmetered(t *testing.T) {
	t.Parallel()
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	for _, k := range []string{"a", "b", "c"} {
		parent.Set(nil, []byte(k), []byte("v"))
	}
	cs := cache.New(parent)
	gctx, m := newGasCtx(gasTestConfig)

	// Current behavior: iterators charge nothing (WillIterator/WillIterNext
	// have no callers). The ADR specifies creation + per-Next charges — an
	// open gap; this pins today's behavior so a change is deliberate.
	it := cs.Iterator(gctx, nil, nil)
	for ; it.Valid(); it.Next() {
		_ = it.Key()
	}
	it.Close()
	require.Zero(t, m.GasConsumed())
}

func TestGas_WriteClearsDedup(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig
	cs := cache.New(dbadapter.Store{DB: memdb.NewMemDB()})
	gctx, m := newGasCtx(cfg)

	one := cfg.WriteCostFlat + cfg.WriteCostPerByte*2
	cs.Set(gctx, []byte("k"), []byte("v1"))
	require.Equal(t, one, m.GasConsumed())

	// Write() flushes and clears the dedup ledger: the same Set afterwards
	// is a fresh full charge, not a refund-and-recharge.
	cs.Write()
	cs.Set(gctx, []byte("k"), []byte("v1"))
	require.Equal(t, 2*one, m.GasConsumed())
}

func TestGas_NilGasContextUnmetered(t *testing.T) {
	t.Parallel()
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	parent.Set(nil, []byte("k"), []byte("v"))
	cs := cache.New(parent)

	cs.Get(nil, []byte("k"))
	cs.Set(nil, []byte("k2"), []byte("v"))
	cs.Delete(nil, []byte("k2"))
	cs.Has(nil, []byte("k"))
	// Nothing to assert on a meter — the contract is simply that nil gctx
	// performs the operation unmetered without panicking.
}

func TestGas_CacheWrapPropagatesEstimator(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	parent.Set(nil, []byte("k"), []byte("vv"))
	cs := cache.New(parent)
	cs.SetDepthEstimator(stubDepth{d: 7})

	// CacheWrap copies the estimator: the nested layer charges depth gas.
	cw := cs.CacheWrap()
	gctx, m := newGasCtx(cfg)
	require.Equal(t, 7*cfg.ReadCostFlat+cfg.ReadCostPerByte*2,
		delta(m, func() { cw.Get(gctx, []byte("k")) }))

	// cache.New over a cacheStore does NOT auto-detect depth (cacheStore is
	// not a DepthEstimator): the outer layer falls back to flat charging.
	outer := cache.New(cs)
	gctx2, m2 := newGasCtx(cfg)
	require.Equal(t, cfg.ReadCostFlat+cfg.ReadCostPerByte*2,
		delta(m2, func() { outer.Get(gctx2, []byte("k")) }))
}
