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

	require.False(t, keeper.Has(ctx, "params_test:param1"))
	require.False(t, keeper.Has(ctx, "params_test:param2"))
	require.False(t, keeper.Has(ctx, "params_test:param3"))
	require.False(t, keeper.Has(ctx, "params_test:param4"))
	require.False(t, keeper.Has(ctx, "params_test:param5"))

	// initial set
	require.NotPanics(t, func() { keeper.SetString(ctx, "params_test:param1", "foo") })
	require.NotPanics(t, func() { keeper.SetBool(ctx, "params_test:param2", true) })
	require.NotPanics(t, func() { keeper.SetUint64(ctx, "params_test:param3", 42) })
	require.NotPanics(t, func() { keeper.SetInt64(ctx, "params_test:param4", -1337) })
	require.NotPanics(t, func() { keeper.SetBytes(ctx, "params_test:param5", []byte("hello world!")) })

	require.True(t, keeper.Has(ctx, "params_test:param1"))
	require.True(t, keeper.Has(ctx, "params_test:param2"))
	require.True(t, keeper.Has(ctx, "params_test:param3"))
	require.True(t, keeper.Has(ctx, "params_test:param4"))
	require.True(t, keeper.Has(ctx, "params_test:param5"))

	var (
		param1 string
		param2 bool
		param3 uint64
		param4 int64
		param5 []byte
	)

	require.NotPanics(t, func() { keeper.GetString(ctx, "params_test:param1", &param1) })
	require.NotPanics(t, func() { keeper.GetBool(ctx, "params_test:param2", &param2) })
	require.NotPanics(t, func() { keeper.GetUint64(ctx, "params_test:param3", &param3) })
	require.NotPanics(t, func() { keeper.GetInt64(ctx, "params_test:param4", &param4) })
	require.NotPanics(t, func() { keeper.GetBytes(ctx, "params_test:param5", &param5) })

	require.Equal(t, param1, "foo")
	require.Equal(t, param2, true)
	require.Equal(t, param3, uint64(42))
	require.Equal(t, param4, int64(-1337))
	require.Equal(t, param5, []byte("hello world!"))

	// reset
	require.NotPanics(t, func() { keeper.SetString(ctx, "params_test:param1", "bar") })
	require.NotPanics(t, func() { keeper.SetBool(ctx, "params_test:param2", false) })
	require.NotPanics(t, func() { keeper.SetUint64(ctx, "params_test:param3", 12345) })
	require.NotPanics(t, func() { keeper.SetInt64(ctx, "params_test:param4", 1000) })
	require.NotPanics(t, func() { keeper.SetBytes(ctx, "params_test:param5", []byte("bye")) })

	require.True(t, keeper.Has(ctx, "params_test:param1"))
	require.True(t, keeper.Has(ctx, "params_test:param2"))
	require.True(t, keeper.Has(ctx, "params_test:param3"))
	require.True(t, keeper.Has(ctx, "params_test:param4"))
	require.True(t, keeper.Has(ctx, "params_test:param5"))

	require.NotPanics(t, func() { keeper.GetString(ctx, "params_test:param1", &param1) })
	require.NotPanics(t, func() { keeper.GetBool(ctx, "params_test:param2", &param2) })
	require.NotPanics(t, func() { keeper.GetUint64(ctx, "params_test:param3", &param3) })
	require.NotPanics(t, func() { keeper.GetInt64(ctx, "params_test:param4", &param4) })
	require.NotPanics(t, func() { keeper.GetBytes(ctx, "params_test:param5", &param5) })

	require.Equal(t, param1, "bar")
	require.Equal(t, param2, false)
	require.Equal(t, param3, uint64(12345))
	require.Equal(t, param4, int64(1000))
	require.Equal(t, param5, []byte("bye"))
}

// adapted from TestKeeperSubspace from Cosmos SDK, but adapted to a subspace-less Keeper.
func TestKeeper_internal(t *testing.T) {
	env := setupTestEnv()
	ctx, store, keeper := env.ctx, env.store, env.keeper

	kvs := []struct {
		key   string
		param any
		zero  any
		ptr   any
	}{
		{"params_test:string", "test", "", new(string)},
		{"params_test:bool", true, false, new(bool)},
		{"params_test:int16", int16(1), int16(0), new(int16)},
		{"params_test:int32", int32(1), int32(0), new(int32)},
		{"params_test:int64", int64(1), int64(0), new(int64)},
		{"params_test:uint16", uint16(1), uint16(0), new(uint16)},
		{"params_test:uint32", uint32(1), uint32(0), new(uint32)},
		{"params_test:uint64", uint64(1), uint64(0), new(uint64)},
		{"params_test:struct", s{1}, s{0}, new(s)},
	}

	for i, kv := range kvs {
		require.NotPanics(t, func() { keeper.set(ctx, kv.key, kv.param) }, "keeper.Set panics, tc #%d", i)
	}

	for i, kv := range kvs {
		require.NotPanics(t, func() { keeper.getIfExists(ctx, "invalid", kv.ptr) }, "keeper.GetIfExists panics when no value exists, tc #%d", i)
		require.Equal(t, kv.zero, indirect(kv.ptr), "keeper.GetIfExists unmarshalls when no value exists, tc #%d", i)
		require.NotPanics(t, func() { keeper.getIfExists(ctx, kv.key, kv.ptr) }, "keeper.GetIfExists panics, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)
		require.Panics(t, func() { keeper.getIfExists(ctx, kv.key, nil) }, "invalid keeper.Get not panics when the pointer is nil, tc #%d", i)
	}

	for i, kv := range kvs {
		vk := storeKey(kv.key)
		bz := store.Get(vk)
		require.NotNil(t, bz, "store.Get() returns nil, tc #%d", i)
		err := amino.UnmarshalJSON(bz, kv.ptr)
		require.NoError(t, err, "cdc.UnmarshalJSON() returns error, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)
	}
}

type s struct{ I int }

func indirect(ptr any) any { return reflect.ValueOf(ptr).Elem().Interface() }

type Params struct {
	p1 int
	p2 string
}

func TestGetAndSetStruct(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx
	keeper := env.keeper
	// SetStruct
	a := Params{p1: 1, p2: "a"}
	keeper.SetStruct(ctx, "params_test:p", a)

	// GetStruct
	a1 := Params{}
	keeper.GetStruct(ctx, "params_test:p", &a1)
	require.True(t, amino.DeepEqual(a, a1), "a and a1 should equal")
}
