package cache_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// Gas charging at the cache.Store boundary. All assertions are relative to
// the GasConfig fields, never absolute calibrated numbers, and depths use a
// stub estimator (100x fixed-point) so tests are independent of tree state.

var gasTestConfig = types.GasConfig{
	ReadCostFlat:     100,
	ReadCostPerByte:  1,
	WriteCostFlat:    200,
	WriteCostPerByte: 2,
	DeleteCost:       150,
	IterNextCostFlat: 30,
}

// stubDepth implements types.DepthEstimator with fixed 100x depths.
type stubDepth struct{ g, s, w int64 }

func (d stubDepth) ExpectedGetReadDepth100() int64 { return d.g }
func (d stubDepth) ExpectedSetReadDepth100() int64 { return d.s }
func (d stubDepth) ExpectedWriteDepth100() int64   { return d.w }

// depthParent wraps a flat store with a stub estimator so cache.New
// auto-detects depth charging.
type depthParent struct {
	types.Store
	stubDepth
}

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
	require.Equal(t, cfg.ReadCostFlat+cfg.ReadCostPerByte*5,
		delta(m, func() { cs.Get(gctx, []byte("k")) }))
	// Hit (now cached): free.
	require.Zero(t, delta(m, func() { cs.Get(gctx, []byte("k")) }))
	// Absent-key miss charges the flat cost only; cached absence is free.
	require.Equal(t, cfg.ReadCostFlat, delta(m, func() { cs.Get(gctx, []byte("absent")) }))
	require.Zero(t, delta(m, func() { cs.Get(gctx, []byte("absent")) }))
}

func TestGas_GetDepth(t *testing.T) {
	t.Parallel()
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	parent.Set(nil, []byte("k"), []byte("hello"))
	// 7.5 read ops: fractional depths must not truncate to integers early.
	cs := cache.New(depthParent{Store: parent, stubDepth: stubDepth{g: 750, s: 200, w: 440}})
	gctx, m := newGasCtx(gasTestConfig)
	cfg := gasTestConfig

	miss := delta(m, func() { cs.Get(gctx, []byte("k")) })
	require.Equal(t, 750*cfg.ReadCostFlat/100+cfg.ReadCostPerByte*5, miss)
	require.Zero(t, delta(m, func() { cs.Get(gctx, []byte("k")) }))
}

func TestGas_DepthFloorsAndOverrides(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig

	// Min* floors lift a low tree estimate (200 -> 300).
	cfgMin := cfg
	cfgMin.MinGetReadDepth100 = 300
	cs := cache.New(depthParent{
		Store:     dbadapter.Store{DB: memdb.NewMemDB()},
		stubDepth: stubDepth{g: 200, s: 200, w: 200},
	})
	gctx, m := newGasCtx(cfgMin)
	require.Equal(t, 300*cfg.ReadCostFlat/100,
		delta(m, func() { cs.Get(gctx, []byte("absent")) }))

	// Fixed* overrides win over both the estimate and the floor.
	cfgFixed := cfgMin
	cfgFixed.FixedGetReadDepth100 = 500
	cs2 := cache.New(depthParent{
		Store:     dbadapter.Store{DB: memdb.NewMemDB()},
		stubDepth: stubDepth{g: 200, s: 200, w: 200},
	})
	gctx2, m2 := newGasCtx(cfgFixed)
	require.Equal(t, 500*cfg.ReadCostFlat/100,
		delta(m2, func() { cs2.Get(gctx2, []byte("absent")) }))

	// Floors do NOT reach flat stores: without an estimator the flat
	// path charges plain ReadCostFlat regardless of Min*Depth100.
	cs3 := cache.New(dbadapter.Store{DB: memdb.NewMemDB()})
	gctx3, m3 := newGasCtx(cfgMin)
	require.Equal(t, cfg.ReadCostFlat,
		delta(m3, func() { cs3.Get(gctx3, []byte("absent")) }))
}

func TestGas_SetFlatAndDepth(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig

	cs := cache.New(dbadapter.Store{DB: memdb.NewMemDB()})
	gctx, m := newGasCtx(cfg)
	require.Equal(t, cfg.WriteCostFlat+cfg.WriteCostPerByte*4,
		delta(m, func() { cs.Set(gctx, []byte("k"), []byte("vvvv")) }))

	// Depth path: setRead*ReadCostFlat/100 + write*WriteCostFlat/100 +
	// per-byte on the value. Key bytes are never charged.
	cs2 := cache.New(depthParent{
		Store:     dbadapter.Store{DB: memdb.NewMemDB()},
		stubDepth: stubDepth{g: 300, s: 200, w: 440},
	})
	gctx2, m2 := newGasCtx(cfg)
	want := 200*cfg.ReadCostFlat/100 + 440*cfg.WriteCostFlat/100 + cfg.WriteCostPerByte*4
	require.Equal(t, want, delta(m2, func() { cs2.Set(gctx2, []byte("k"), []byte("vvvv")) }))
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

	cs2 := cache.New(depthParent{
		Store:     dbadapter.Store{DB: memdb.NewMemDB()},
		stubDepth: stubDepth{g: 300, s: 200, w: 440},
	})
	gctx2, m2 := newGasCtx(cfg)
	want := 200*cfg.ReadCostFlat/100 + 440*cfg.WriteCostFlat/100
	require.Equal(t, want, delta(m2, func() { cs2.Delete(gctx2, []byte("k")) }))
}

func TestGas_HasDelegatesToGet(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	parent.Set(nil, []byte("k"), []byte("vv"))
	cs := cache.New(parent)
	gctx, m := newGasCtx(cfg)

	require.Equal(t, cfg.ReadCostFlat+cfg.ReadCostPerByte*2,
		delta(m, func() { cs.Has(gctx, []byte("k")) }))
	require.Zero(t, delta(m, func() { cs.Has(gctx, []byte("k")) }))
}

func TestGas_IteratorMetered(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	for _, k := range []string{"a", "b", "c"} {
		parent.Set(nil, []byte(k), []byte("v"))
	}
	cs := cache.New(parent)
	gctx, m := newGasCtx(cfg)

	// Creation charges one flat seek plus the first position; each valid
	// advance charges step + per-byte; the advance past the end is free.
	got := delta(m, func() {
		it := cs.Iterator(gctx, nil, nil)
		for ; it.Valid(); it.Next() {
			_ = it.Key()
		}
		it.Close()
	})
	want := cfg.ReadCostFlat + 3*(cfg.IterNextCostFlat+cfg.ReadCostPerByte*1)
	require.Equal(t, want, got)

	// An empty range still pays the seek (the walk happens regardless).
	got = delta(m, func() {
		it := cs.Iterator(gctx, []byte("x"), nil)
		require.False(t, it.Valid())
		it.Close()
	})
	require.Equal(t, cfg.ReadCostFlat, got)
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

	// The contract: nil gctx performs every op unmetered without panicking.
	cs.Get(nil, []byte("k"))
	cs.Set(nil, []byte("k2"), []byte("v"))
	cs.Delete(nil, []byte("k2"))
	cs.Has(nil, []byte("k"))
	it := cs.Iterator(nil, nil, nil)
	for ; it.Valid(); it.Next() {
	}
	it.Close()
}

func TestGas_CacheWrapPropagatesDepths(t *testing.T) {
	t.Parallel()
	cfg := gasTestConfig
	parent := dbadapter.Store{DB: memdb.NewMemDB()}
	parent.Set(nil, []byte("k"), []byte("vv"))
	cs := cache.New(depthParent{Store: parent, stubDepth: stubDepth{g: 750, s: 200, w: 440}})

	// CacheWrap copies the cached depths: the nested layer charges depth gas.
	cw := cs.CacheWrap()
	gctx, m := newGasCtx(cfg)
	require.Equal(t, 750*cfg.ReadCostFlat/100+cfg.ReadCostPerByte*2,
		delta(m, func() { cw.Get(gctx, []byte("k")) }))

	// cache.New over a cacheStore does NOT auto-detect depth (cacheStore is
	// not a DepthEstimator): the outer layer falls back to flat charging.
	outer := cache.New(cs)
	gctx2, m2 := newGasCtx(cfg)
	require.Equal(t, cfg.ReadCostFlat+cfg.ReadCostPerByte*2,
		delta(m2, func() { outer.Get(gctx2, []byte("k")) }))
}
