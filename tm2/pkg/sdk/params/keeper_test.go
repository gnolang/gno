package params

import (
	"reflect"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

func TestKeeper(t *testing.T) {
	env := setupTestEnv()
	ctx, store, keeper := env.ctx, env.store, env.keeper
	_ = store // XXX: add store tests?

	require.False(t, keeper.Has(ctx, "param1.string"))
	require.False(t, keeper.Has(ctx, "param2.bool"))
	require.False(t, keeper.Has(ctx, "param3.uint64"))
	require.False(t, keeper.Has(ctx, "param4.int64"))
	require.False(t, keeper.Has(ctx, "param5.bytes"))

	// initial set
	require.NotPanics(t, func() { keeper.SetString(ctx, "param1.string", "foo") })
	require.NotPanics(t, func() { keeper.SetBool(ctx, "param2.bool", true) })
	require.NotPanics(t, func() { keeper.SetUint64(ctx, "param3.uint64", 42) })
	require.NotPanics(t, func() { keeper.SetInt64(ctx, "param4.int64", -1337) })
	require.NotPanics(t, func() { keeper.SetBytes(ctx, "param5.bytes", []byte("hello world!")) })

	require.True(t, keeper.Has(ctx, "param1.string"))
	require.True(t, keeper.Has(ctx, "param2.bool"))
	require.True(t, keeper.Has(ctx, "param3.uint64"))
	require.True(t, keeper.Has(ctx, "param4.int64"))
	require.True(t, keeper.Has(ctx, "param5.bytes"))

	var (
		param1 string
		param2 bool
		param3 uint64
		param4 int64
		param5 []byte
	)

	require.NotPanics(t, func() { keeper.GetString(ctx, "param1.string", &param1) })
	require.NotPanics(t, func() { keeper.GetBool(ctx, "param2.bool", &param2) })
	require.NotPanics(t, func() { keeper.GetUint64(ctx, "param3.uint64", &param3) })
	require.NotPanics(t, func() { keeper.GetInt64(ctx, "param4.int64", &param4) })
	require.NotPanics(t, func() { keeper.GetBytes(ctx, "param5.bytes", &param5) })

	require.Equal(t, param1, "foo")
	require.Equal(t, param2, true)
	require.Equal(t, param3, uint64(42))
	require.Equal(t, param4, int64(-1337))
	require.Equal(t, param5, []byte("hello world!"))

	// reset
	require.NotPanics(t, func() { keeper.SetString(ctx, "param1.string", "bar") })
	require.NotPanics(t, func() { keeper.SetBool(ctx, "param2.bool", false) })
	require.NotPanics(t, func() { keeper.SetUint64(ctx, "param3.uint64", 12345) })
	require.NotPanics(t, func() { keeper.SetInt64(ctx, "param4.int64", 1000) })
	require.NotPanics(t, func() { keeper.SetBytes(ctx, "param5.bytes", []byte("bye")) })

	require.True(t, keeper.Has(ctx, "param1.string"))
	require.True(t, keeper.Has(ctx, "param2.bool"))
	require.True(t, keeper.Has(ctx, "param3.uint64"))
	require.True(t, keeper.Has(ctx, "param4.int64"))
	require.True(t, keeper.Has(ctx, "param5.bytes"))

	require.NotPanics(t, func() { keeper.GetString(ctx, "param1.string", &param1) })
	require.NotPanics(t, func() { keeper.GetBool(ctx, "param2.bool", &param2) })
	require.NotPanics(t, func() { keeper.GetUint64(ctx, "param3.uint64", &param3) })
	require.NotPanics(t, func() { keeper.GetInt64(ctx, "param4.int64", &param4) })
	require.NotPanics(t, func() { keeper.GetBytes(ctx, "param5.bytes", &param5) })

	require.Equal(t, param1, "bar")
	require.Equal(t, param2, false)
	require.Equal(t, param3, uint64(12345))
	require.Equal(t, param4, int64(1000))
	require.Equal(t, param5, []byte("bye"))

	// invalid sets
	require.PanicsWithValue(t, `key should be like "<name>.string"`, func() { keeper.SetString(ctx, "invalid.int64", "hello") })
	require.PanicsWithValue(t, `key should be like "<name>.int64"`, func() { keeper.SetInt64(ctx, "invalid.string", int64(42)) })
	require.PanicsWithValue(t, `key should be like "<name>.uint64"`, func() { keeper.SetUint64(ctx, "invalid.int64", uint64(42)) })
	require.PanicsWithValue(t, `key should be like "<name>.bool"`, func() { keeper.SetBool(ctx, "invalid.int64", true) })
	require.PanicsWithValue(t, `key should be like "<name>.bytes"`, func() { keeper.SetBytes(ctx, "invalid.int64", []byte("hello")) })
}

// adapted from TestKeeperSubspace from Cosmos SDK, but adapted to a subspace-less Keeper.
func TestKeeper_internal(t *testing.T) {
	env := setupTestEnv()
	ctx, store, keeper := env.ctx, env.store, env.keeper

	kvs := []struct {
		key   string
		param interface{}
		zero  interface{}
		ptr   interface{}
	}{
		{"string", "test", "", new(string)},
		{"bool", true, false, new(bool)},
		{"int16", int16(1), int16(0), new(int16)},
		{"int32", int32(1), int32(0), new(int32)},
		{"int64", int64(1), int64(0), new(int64)},
		{"uint16", uint16(1), uint16(0), new(uint16)},
		{"uint32", uint32(1), uint32(0), new(uint32)},
		{"uint64", uint64(1), uint64(0), new(uint64)},
		{"struct", s{1}, s{0}, new(s)},
	}

	for i, kv := range kvs {
		require.NotPanics(t, func() { keeper.set(ctx, kv.key, kv.param) }, "keeper.Set panics, tc #%d", i)
	}

	for i, kv := range kvs {
		require.NotPanics(t, func() { keeper.getIfExists(ctx, "invalid", kv.ptr) }, "keeper.GetIfExists panics when no value exists, tc #%d", i)
		require.Equal(t, kv.zero, indirect(kv.ptr), "keeper.GetIfExists unmarshalls when no value exists, tc #%d", i)
		require.Panics(t, func() { keeper.get(ctx, "invalid", kv.ptr) }, "invalid keeper.Get not panics when no value exists, tc #%d", i)
		require.Equal(t, kv.zero, indirect(kv.ptr), "invalid keeper.Get unmarshalls when no value exists, tc #%d", i)

		require.NotPanics(t, func() { keeper.getIfExists(ctx, kv.key, kv.ptr) }, "keeper.GetIfExists panics, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)
		require.NotPanics(t, func() { keeper.get(ctx, kv.key, kv.ptr) }, "keeper.Get panics, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)

		require.Panics(t, func() { keeper.get(ctx, "invalid", kv.ptr) }, "invalid keeper.Get not panics when no value exists, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "invalid keeper.Get unmarshalls when no value existt, tc #%d", i)

		require.Panics(t, func() { keeper.get(ctx, kv.key, nil) }, "invalid keeper.Get not panics when the pointer is nil, tc #%d", i)
	}

	for i, kv := range kvs {
		bz := store.Get([]byte(kv.key))
		require.NotNil(t, bz, "store.Get() returns nil, tc #%d", i)
		err := amino.UnmarshalJSON(bz, kv.ptr)
		require.NoError(t, err, "cdc.UnmarshalJSON() returns error, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)
	}
}

type (
	invalid struct{}
	s       struct{ I int }
)

func indirect(ptr interface{}) interface{} { return reflect.ValueOf(ptr).Elem().Interface() }
