package params

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// cacheParams has the kinds the real param structs are made of: numbers,
// strings, an address-like array, slices of both, a nested struct.
type cacheParams struct {
	N  int64      `json:"n"`
	S  string     `json:"s"`
	A  [4]byte    `json:"a"`
	L  []string   `json:"l"`
	AL [][4]byte  `json:"al"`
	G  cacheInner `json:"g"`
}

type cacheInner struct {
	X int64  `json:"x"`
	Y string `json:"y"`
}

const cacheKey = dummyModuleName + ":p"

func TestGetStructCachesDecode(t *testing.T) {
	env := setupTestEnv()
	ctx, keeper := env.ctx, env.keeper
	a := cacheParams{N: 7, S: "seven", A: [4]byte{1, 2, 3, 4}, L: []string{"x", "y"}, AL: [][4]byte{{9, 9, 9, 9}}, G: cacheInner{X: 1, Y: "one"}}
	keeper.SetStruct(ctx, cacheKey, a)

	var b cacheParams
	keeper.GetStruct(ctx, cacheKey, &b)
	require.Equal(t, a, b)
	e := keeper.cache.entries[cacheKey]
	require.NotNil(t, e)
	require.Len(t, e.vals, 6)

	// The same bytes in the store: the read is served from the cache, which
	// a value planted in it shows.
	e.vals[0] = reflect.ValueOf(int64(99))
	var c cacheParams
	keeper.GetStruct(ctx, cacheKey, &c)
	require.Equal(t, int64(99), c.N)
	require.Equal(t, a.L, c.L)

	// What a caller gets is theirs: changing it changes neither the cache
	// nor the next read.
	c.L[0] = "changed"
	c.AL[0][0] = 0
	var d cacheParams
	keeper.GetStruct(ctx, cacheKey, &d)
	require.Equal(t, "x", d.L[0])
	require.Equal(t, byte(9), d.AL[0][0])

	// New bytes: decoded afresh, the planted value gone.
	a2 := a
	a2.N = 8
	a2.L = []string{"z"}
	keeper.SetStruct(ctx, cacheKey, a2)
	var f cacheParams
	keeper.GetStruct(ctx, cacheKey, &f)
	require.Equal(t, a2, f)
	require.NotSame(t, e, keeper.cache.entries[cacheKey])
}

// A field with no value in the store is left as the caller had it, by a
// decode and by the cache alike.
func TestGetStructCacheLeavesAbsentFields(t *testing.T) {
	env := setupTestEnv()
	ctx, keeper := env.ctx, env.keeper
	keeper.SetStruct(ctx, cacheKey, cacheParams{N: 1, S: "one"})
	env.store.Delete(nil, storeKey(cacheKey+":s"))
	for _, pass := range []string{"decoded", "cached"} {
		p := cacheParams{S: "kept"}
		keeper.GetStruct(ctx, cacheKey, &p)
		require.Equal(t, int64(1), p.N, pass)
		require.Equal(t, "kept", p.S, pass)
	}
}

type pointerParams struct {
	N int64  `json:"n"`
	P *int64 `json:"p"`
}

// A struct with a field the cache cannot copy safely is decoded every time
// and never cached.
func TestGetStructCacheSkipsUncopyableStructs(t *testing.T) {
	env := setupTestEnv()
	ctx, keeper := env.ctx, env.keeper
	seven := int64(7)
	keeper.SetStruct(ctx, cacheKey, pointerParams{N: 1, P: &seven})
	for range 2 {
		var p pointerParams
		keeper.GetStruct(ctx, cacheKey, &p)
		require.Equal(t, int64(1), p.N)
		require.NotNil(t, p.P)
		require.Equal(t, int64(7), *p.P)
		require.NotContains(t, keeper.cache.entries, cacheKey)
	}
}

// The cache follows the bytes the context's store holds, not the writes: a
// write in a cache-wrapped store is seen through that store and not through
// its parent, as it would be for a transaction that is then reverted.
func TestGetStructCacheFollowsTheStore(t *testing.T) {
	env := setupTestEnv()
	ctx, keeper := env.ctx, env.keeper
	keeper.SetStruct(ctx, cacheKey, cacheParams{N: 1})
	var p cacheParams
	keeper.GetStruct(ctx, cacheKey, &p)
	require.Equal(t, int64(1), p.N)

	wrapped := ctx.WithMultiStore(ctx.MultiStore().MultiCacheWrap())
	keeper.SetStruct(wrapped, cacheKey, cacheParams{N: 2})
	keeper.GetStruct(wrapped, cacheKey, &p)
	require.Equal(t, int64(2), p.N)
	keeper.GetStruct(ctx, cacheKey, &p)
	require.Equal(t, int64(1), p.N)
	keeper.GetStruct(wrapped, cacheKey, &p)
	require.Equal(t, int64(2), p.N)
}

// fill and remember from many goroutines at once (run with -race).
func TestStructCacheConcurrent(t *testing.T) {
	a := cacheParams{N: 7, S: "seven", L: []string{"x"}, AL: [][4]byte{{1, 2, 3, 4}}, G: cacheInner{X: 1}}
	kvz := encodeStructFields(a)
	c := newStructCache()
	var wg sync.WaitGroup
	var wrong atomic.Int32
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 200 {
				var p cacheParams
				if !c.fill(cacheKey, &p, kvz) {
					decodeStructFields(&p, kvz)
					c.remember(cacheKey, &p, kvz)
				}
				if !reflect.DeepEqual(p, a) {
					wrong.Add(1)
				}
				if i%50 == 0 {
					c.remember(cacheKey, &a, kvz)
				}
			}
		}()
	}
	wg.Wait()
	require.Zero(t, wrong.Load())
}

func TestCopyValue(t *testing.T) {
	seven := 7
	for _, tc := range []struct {
		name string
		v    any
		ok   bool
	}{
		{"int", int64(7), true},
		{"string", "s", true},
		{"array", [4]byte{1, 2, 3, 4}, true},
		{"slice", []string{"a", "b"}, true},
		{"nil slice", []string(nil), true},
		{"struct", cacheInner{X: 1, Y: "y"}, true},
		{"nested", cacheParams{N: 1, L: []string{"a"}, AL: [][4]byte{{1}}, G: cacheInner{X: 2}}, true},
		{"pointer", &seven, false},
		{"map", map[string]int{"a": 1}, false},
		{"interface field", struct{ I any }{I: 1}, false},
		{"unexported field", cacheUnexported{n: 1}, false},
	} {
		cp, ok := copyValue(reflect.ValueOf(tc.v))
		require.Equal(t, tc.ok, ok, tc.name)
		if ok {
			require.Equal(t, tc.v, cp.Interface(), tc.name)
		}
	}

	// A copy shares nothing with its source.
	src := cacheParams{L: []string{"a"}, AL: [][4]byte{{1}}}
	cp, ok := copyValue(reflect.ValueOf(src))
	require.True(t, ok)
	got := cp.Interface().(cacheParams)
	got.L[0] = "b"
	got.AL[0][0] = 2
	require.Equal(t, "a", src.L[0])
	require.Equal(t, byte(1), src.AL[0][0])
}

type cacheUnexported struct{ n int }

// benchParams is the shape of auth.Params: seven numbers, a nested struct, a
// slice of addresses and an address.
type benchParams struct {
	A1 int64         `json:"a1"`
	A2 int64         `json:"a2"`
	A3 int64         `json:"a3"`
	A4 int64         `json:"a4"`
	A5 int64         `json:"a5"`
	A6 int64         `json:"a6"`
	A7 int64         `json:"a7"`
	G  benchGasPrice `json:"g"`
	U  [][20]byte    `json:"u"`
	F  [20]byte      `json:"f"`
}

type benchGasPrice struct {
	Gas    int64  `json:"gas"`
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

// BenchmarkGetStruct reads a struct the size of auth.Params with the cache,
// as GetStruct runs, and without it, the decode on every call as before. The
// store reads are the same in both.
func BenchmarkGetStruct(b *testing.B) {
	env := setupTestEnv()
	ctx := env.ctx
	env.keeper.SetStruct(ctx, cacheKey, benchParams{A1: 1, A2: 2, A3: 3, A4: 4, A5: 5, A6: 6, A7: 7, G: benchGasPrice{Gas: 1000, Denom: "ugnot", Amount: 1}, U: [][20]byte{{1}, {2}}, F: [20]byte{3}})
	withoutCache := env.keeper
	withoutCache.cache = nil
	for _, bc := range []struct {
		name   string
		keeper ParamsKeeper
	}{{"cached", env.keeper}, {"decoded", withoutCache}} {
		b.Run(bc.name, func(b *testing.B) {
			var p benchParams
			for range b.N {
				bc.keeper.GetStruct(ctx, cacheKey, &p)
			}
		})
	}
}
