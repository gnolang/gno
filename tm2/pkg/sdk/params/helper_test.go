package params

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/std"
)

type plainField struct{ A int } // does not implement json.Unmarshaler

func TestDecodeStructFields(t *testing.T) {
	t.Run("non-unmarshaler field: no fallback, no mutation, field name in panic", func(t *testing.T) {
		type params struct {
			Num   int64      `json:"num"`
			Plain plainField `json:"plain"`
		}
		prm := &params{Num: 7, Plain: plainField{A: 5}}
		kvz := []std.KVPair{{Key: []byte("plain"), Value: []byte(`"oops"`)}}

		defer func() {
			r := recover()
			require.NotNil(t, r, "expected panic on undecodable field")
			err, ok := r.(error)
			require.True(t, ok)
			assert.Contains(t, err.Error(), `"plain"`, "panic must name the field")
			// amino failed and plainField has no json.Unmarshaler, so the
			// field must be left untouched (no partial mutation).
			assert.Equal(t, plainField{A: 5}, prm.Plain)
		}()
		decodeStructFields(prm, kvz)
	})

	t.Run("json.Unmarshaler field: string value falls back and decodes", func(t *testing.T) {
		type params struct {
			Num   int64        `json:"num"`
			Price std.GasPrice `json:"price"`
		}
		prm := &params{Num: 7}
		kvz := []std.KVPair{{Key: []byte("price"), Value: []byte(`"1ugnot/1gas"`)}}

		require.NotPanics(t, func() { decodeStructFields(prm, kvz) })
		assert.Equal(t, std.GasPrice{Gas: 1, Price: std.Coin{Denom: "ugnot", Amount: 1}}, prm.Price)
		assert.Equal(t, int64(7), prm.Num)
	})
}

func TestMustParamString(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		got := MustParamString("foo", "bar")
		require.Equal(t, "bar", got)
	})

	t.Run("wrong type panics", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"invalid type for foo param: expected string, got int",
			func() { MustParamString("foo", 42) },
		)
	})
}

func TestMustParamInt64(t *testing.T) {
	t.Run("valid int64", func(t *testing.T) {
		got := MustParamInt64("num", int64(99))
		require.Equal(t, int64(99), got)
	})

	t.Run("wrong type panics", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"invalid type for num param: expected int64, got string",
			func() { MustParamInt64("num", "not a number") },
		)
	})
}

func TestMustParamStrings(t *testing.T) {
	t.Run("valid []string", func(t *testing.T) {
		got := MustParamStrings("tags", []string{"a", "b"})
		require.Equal(t, []string{"a", "b"}, got)
	})

	t.Run("wrong type panics", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"invalid type for tags param: expected []string, got string",
			func() { MustParamStrings("tags", "not a slice") },
		)
	})
}

type fieldNamesParams struct {
	Num   int64        `json:"num"`
	Price std.GasPrice `json:"price"`
	Names []string     `json:"names"`
	Plain int64
}

func TestFieldNames(t *testing.T) {
	rt := reflect.TypeFor[fieldNamesParams]()

	t.Run("computed once per type", func(t *testing.T) {
		a, b := fieldNames(rt), fieldNames(rt)
		require.NotEmpty(t, a)
		assert.Same(t, &a[0], &b[0], "second call must reuse the cached names")
	})

	t.Run("rejected type panics every time and is not cached", func(t *testing.T) {
		rt := reflect.TypeFor[struct{ F func() }]()
		assert.Panics(t, func() { fieldNames(rt) })
		assert.Panics(t, func() { fieldNames(rt) })
		_, cached := fieldNamesCache.Load(rt)
		assert.False(t, cached)
	})
}

func TestSetAndGetStructFieldNames(t *testing.T) {
	env := setupTestEnv()
	key := dummyModuleName + ":p"
	want := fieldNamesParams{
		Num:   7,
		Price: std.GasPrice{Gas: 1, Price: std.Coin{Denom: "ugnot", Amount: 1}},
		Names: []string{"x", "y"},
		Plain: 3,
	}
	env.keeper.SetStruct(env.ctx, key, want)

	for _, name := range []string{"num", "price", "names", "Plain"} {
		assert.True(t, env.store.Has(nil, storeKey(key+":"+name)), "field %q stored under its JSON name", name)
	}

	var got fieldNamesParams
	env.keeper.GetStruct(env.ctx, key, &got)
	assert.Equal(t, want, got)
}

// Run with -race: first lookups of a type race to fill the cache.
func TestFieldNamesConcurrent(t *testing.T) {
	type concurrentParams struct {
		A int64  `json:"a"`
		B string `json:"b"`
	}
	rt := reflect.TypeFor[concurrentParams]()

	var wg sync.WaitGroup
	results := make([][]string, 16)
	for i := range results {
		wg.Go(func() { results[i] = fieldNames(rt) })
	}
	wg.Wait()
	for _, names := range results {
		assert.Equal(t, []string{"a", "b"}, names)
	}
}

func BenchmarkGetStruct(b *testing.B) {
	env := setupTestEnv()
	prm := fieldNamesParams{Num: 7, Names: []string{"x", "y"}, Plain: 3}
	env.keeper.SetStruct(env.ctx, dummyModuleName+":p", prm)

	b.ReportAllocs()
	for b.Loop() {
		var got fieldNamesParams
		env.keeper.GetStruct(env.ctx, dummyModuleName+":p", &got)
	}
}
