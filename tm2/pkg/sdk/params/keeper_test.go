package params

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func TestKeeper(t *testing.T) {
	env := setupTestEnv()
	ctx, store, keeper := env.ctx, env.store, env.keeper
	_ = store // XXX: add store tests?

	require.False(t, keeper.Has(ctx, "param1"))
	require.False(t, keeper.Has(ctx, "param2"))
	require.False(t, keeper.Has(ctx, "param3"))
	require.False(t, keeper.Has(ctx, "param4"))
	require.False(t, keeper.Has(ctx, "param5"))

	// initial set
	require.NotPanics(t, func() { keeper.SetString(ctx, "param1", "foo") })
	require.NotPanics(t, func() { keeper.SetBool(ctx, "param2", true) })
	require.NotPanics(t, func() { keeper.SetUint64(ctx, "param3", 42) })
	require.NotPanics(t, func() { keeper.SetInt64(ctx, "param4", -1337) })
	require.NotPanics(t, func() { keeper.SetBytes(ctx, "param5", []byte("hello world!")) })

	require.True(t, keeper.Has(ctx, "param1"))
	require.True(t, keeper.Has(ctx, "param2"))
	require.True(t, keeper.Has(ctx, "param3"))
	require.True(t, keeper.Has(ctx, "param4"))
	require.True(t, keeper.Has(ctx, "param5"))

	var (
		param1 string
		param2 bool
		param3 uint64
		param4 int64
		param5 []byte
	)

	require.NotPanics(t, func() { keeper.GetString(ctx, "param1", &param1) })
	require.NotPanics(t, func() { keeper.GetBool(ctx, "param2", &param2) })
	require.NotPanics(t, func() { keeper.GetUint64(ctx, "param3", &param3) })
	require.NotPanics(t, func() { keeper.GetInt64(ctx, "param4", &param4) })
	require.NotPanics(t, func() { keeper.GetBytes(ctx, "param5", &param5) })

	require.Equal(t, param1, "foo")
	require.Equal(t, param2, true)
	require.Equal(t, param3, uint64(42))
	require.Equal(t, param4, int64(-1337))
	require.Equal(t, param5, []byte("hello world!"))

	// reset
	require.NotPanics(t, func() { keeper.SetString(ctx, "param1", "bar") })
	require.NotPanics(t, func() { keeper.SetBool(ctx, "param2", false) })
	require.NotPanics(t, func() { keeper.SetUint64(ctx, "param3", 12345) })
	require.NotPanics(t, func() { keeper.SetInt64(ctx, "param4", 1000) })
	require.NotPanics(t, func() { keeper.SetBytes(ctx, "param5", []byte("bye")) })

	require.True(t, keeper.Has(ctx, "param1"))
	require.True(t, keeper.Has(ctx, "param2"))
	require.True(t, keeper.Has(ctx, "param3"))
	require.True(t, keeper.Has(ctx, "param4"))
	require.True(t, keeper.Has(ctx, "param5"))

	require.NotPanics(t, func() { keeper.GetString(ctx, "param1", &param1) })
	require.NotPanics(t, func() { keeper.GetBool(ctx, "param2", &param2) })
	require.NotPanics(t, func() { keeper.GetUint64(ctx, "param3", &param3) })
	require.NotPanics(t, func() { keeper.GetInt64(ctx, "param4", &param4) })
	require.NotPanics(t, func() { keeper.GetBytes(ctx, "param5", &param5) })

	require.Equal(t, param1, "bar")
	require.Equal(t, param2, false)
	require.Equal(t, param3, uint64(12345))
	require.Equal(t, param4, int64(1000))
	require.Equal(t, param5, []byte("bye"))

	// Test SetBytes with nil deletes the key
	keeper.SetBytes(ctx, "param5", nil)
	require.False(t, keeper.Has(ctx, "param5"))
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
		require.NotPanics(t, func() { keeper.getIfExists(ctx, kv.key, kv.ptr) }, "keeper.GetIfExists panics, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)
		require.Panics(t, func() { keeper.getIfExists(ctx, kv.key, nil) }, "invalid keeper.Get not panics when the pointer is nil, tc #%d", i)
	}

	for i, kv := range kvs {
		vk := storeKey(kv.key)
		bz := store.Get(nil, vk)
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

// TestKeeperGetReturnsFound verifies the (value, found) contract on
// the typed Get methods: an absent key returns false with the ptr
// untouched; a key SET to a zero value returns true. Without the
// found bit, callers couldn't distinguish "never written" from
// "written as 0/false/empty".
func TestKeeperGetReturnsFound(t *testing.T) {
	env := setupTestEnv()
	ctx, keeper := env.ctx, env.keeper

	// Set each typed key to its zero value.
	keeper.SetString(ctx, "s", "")
	keeper.SetBool(ctx, "b", false)
	keeper.SetInt64(ctx, "i", 0)
	keeper.SetUint64(ctx, "u", 0)
	keeper.SetBytes(ctx, "y", []byte{})
	keeper.SetStrings(ctx, "ss", []string{})

	var (
		s  string
		b  bool
		i  int64
		u  uint64
		y  []byte
		ss []string
	)
	require.True(t, keeper.GetString(ctx, "s", &s), "set-to-empty string is found")
	require.True(t, keeper.GetBool(ctx, "b", &b), "set-to-false bool is found")
	require.True(t, keeper.GetInt64(ctx, "i", &i), "set-to-zero int64 is found")
	require.True(t, keeper.GetUint64(ctx, "u", &u), "set-to-zero uint64 is found")
	require.True(t, keeper.GetBytes(ctx, "y", &y), "set-to-empty bytes is found")
	require.True(t, keeper.GetStrings(ctx, "ss", &ss), "set-to-empty strings is found")

	// Pre-set ptrs to non-zero so we can detect mutation on absent
	// keys.
	s, b, i, u = "preset", true, 7, 7
	y, ss = []byte("preset"), []string{"preset"}
	require.False(t, keeper.GetString(ctx, "absent", &s))
	require.False(t, keeper.GetBool(ctx, "absent", &b))
	require.False(t, keeper.GetInt64(ctx, "absent", &i))
	require.False(t, keeper.GetUint64(ctx, "absent", &u))
	require.False(t, keeper.GetBytes(ctx, "absent", &y))
	require.False(t, keeper.GetStrings(ctx, "absent", &ss))
	// Ptrs left untouched on absent.
	require.Equal(t, "preset", s)
	require.True(t, b)
	require.Equal(t, int64(7), i)
	require.Equal(t, uint64(7), u)
	require.Equal(t, []byte("preset"), y)
	require.Equal(t, []string{"preset"}, ss)
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

// recordingKeeper captures every WillSetParam invocation. Used by the
// H1 regression test below.
type recordingKeeper struct {
	calls []recordedCall
}

type recordedCall struct {
	key   string
	value any
}

func (k *recordingKeeper) WillSetParam(_ sdk.Context, key string, value any) {
	k.calls = append(k.calls, recordedCall{key: key, value: value})
}

// TestSetBytesTriggersWillSetParam (H1 regression): pre-fix, SetBytes
// wrote raw bytes directly to the store, bypassing the registered
// module keeper's WillSetParam hook. This let governance proposals
// using NewSysParamBytesPropRequest write arbitrary bytes to any key
// without validation. After H1, SetBytes routes through validate so
// WillSetParam fires for byte-typed writes too — same as all other
// Set* paths. Storage stays raw (no JSON encoding), so GetBytes still
// reads the original bytes.
func TestSetBytesTriggersWillSetParam(t *testing.T) {
	env := setupTestEnv()
	rec := &recordingKeeper{}
	const moduleName = "h1_test"
	env.keeper.Register(moduleName, rec)

	// Set bytes through a module-prefixed key.
	payload := []byte{0xde, 0xad, 0xbe, 0xef}
	env.keeper.SetBytes(env.ctx, moduleName+":blob", payload)

	require.Len(t, rec.calls, 1, "WillSetParam must fire on SetBytes")
	require.Equal(t, "blob", rec.calls[0].key, "rawKey should strip module prefix")
	gotValue, ok := rec.calls[0].value.([]byte)
	require.True(t, ok, "value passed to WillSetParam must keep []byte type")
	require.Equal(t, payload, gotValue)

	// Storage stays raw (round-trip via GetBytes returns original bytes).
	var read []byte
	env.keeper.GetBytes(env.ctx, moduleName+":blob", &read)
	require.Equal(t, payload, read, "raw bytes must round-trip via GetBytes")
}
