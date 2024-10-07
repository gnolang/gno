package params

import (
	"reflect"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

// XXX: transient test

func TestKeeper(t *testing.T) {
	kvs := []struct {
		key   string
		param int64
	}{
		{"key1", 10},
		{"key2", 55},
		{"key3", 182},
		{"key4", 17582},
		{"key5", 2768554},
		{"key6", 1157279},
		{"key7", 9058701},
	}

	table := NewKeyTable(
		NewParamSetPair("key1", int64(0), validateNoOp),
		NewParamSetPair("key2", int64(0), validateNoOp),
		NewParamSetPair("key3", int64(0), validateNoOp),
		NewParamSetPair("key4", int64(0), validateNoOp),
		NewParamSetPair("key5", int64(0), validateNoOp),
		NewParamSetPair("key6", int64(0), validateNoOp),
		NewParamSetPair("key7", int64(0), validateNoOp),
		NewParamSetPair("extra1", bool(false), validateNoOp),
		NewParamSetPair("extra2", string(""), validateNoOp),
	)

	env := setupTestEnv()
	ctx, store, keeper := env.ctx, env.store, env.keeper

	require.False(t, keeper.HasKeyTable())
	keeper = keeper.WithKeyTable(table)
	require.True(t, keeper.HasKeyTable())

	// Set params
	for i, kv := range kvs {
		require.NotPanics(t, func() { keeper.Set(ctx, kv.key, kv.param) }, "keeper.Set panics, tc #%d", i)
	}

	// Test keeper.Get
	for i, kv := range kvs {
		var param int64
		require.NotPanics(t, func() { keeper.Get(ctx, kv.key, &param) }, "keeper.Get panics, tc #%d", i)
		require.Equal(t, kv.param, param, "stored param not equal, tc #%d", i)
	}

	// Test keeper.GetRaw
	for i, kv := range kvs {
		var param int64
		bz := keeper.GetRaw(ctx, kv.key)
		err := amino.UnmarshalJSON(bz, &param)
		require.Nil(t, err, "err is not nil, tc #%d", i)
		require.Equal(t, kv.param, param, "stored param not equal, tc #%d", i)
	}

	// Test store.Get equals keeper.Get
	for i, kv := range kvs {
		var param int64
		bz := store.Get([]byte(kv.key))
		require.NotNil(t, bz, "KVStore.Get returns nil, tc #%d", i)
		err := amino.UnmarshalJSON(bz, &param)
		require.NoError(t, err, "UnmarshalJSON returns error, tc #%d", i)
		require.Equal(t, kv.param, param, "stored param not equal, tc #%d", i)
	}

	// Test invalid keeper.Get
	for i, kv := range kvs {
		var param bool
		require.Panics(t, func() { keeper.Get(ctx, kv.key, &param) }, "invalid keeper.Get not panics, tc #%d", i)
	}

	// Test invalid keeper.Set
	for i, kv := range kvs {
		require.Panics(t, func() { keeper.Set(ctx, kv.key, true) }, "invalid keeper.Set not panics, tc #%d", i)
	}
}

// adapted from TestKeeperSubspace from Cosmos SDK, but adapted to a subspace-less Keeper.
func TestKeeper_Subspace(t *testing.T) {
	env := setupTestEnv()
	ctx, store, keeper := env.ctx, env.store, env.keeper
	// XXX: keeper = keeper.WithKeyTable(table)

	// cdc, ctx, key, _, keeper := testComponents()

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
		// XXX: {"int", math.NewInt(1), math.Int{}, new(math.Int)},
		// XXX: {"uint", math.NewUint(1), math.Uint{}, new(math.Uint)},
		// XXX: {"dec", math.LegacyNewDec(1), math.LegacyDec{}, new(math.LegacyDec)},
		{"struct", s{1}, s{0}, new(s)},
	}

	table := NewKeyTable(
		NewParamSetPair("string", "", validateNoOp),
		NewParamSetPair("bool", false, validateNoOp),
		NewParamSetPair("int16", int16(0), validateNoOp),
		NewParamSetPair("int32", int32(0), validateNoOp),
		NewParamSetPair("int64", int64(0), validateNoOp),
		NewParamSetPair("uint16", uint16(0), validateNoOp),
		NewParamSetPair("uint32", uint32(0), validateNoOp),
		NewParamSetPair("uint64", uint64(0), validateNoOp),
		// XXX: NewParamSetPair("int", math.Int{}, validateNoOp),
		// XXX: NewParamSetPair("uint", math.Uint{}, validateNoOp),
		// XXX: NewParamSetPair("dec", math.LegacyDec{}, validateNoOp),
		NewParamSetPair("struct", s{}, validateNoOp),
	)
	keeper = keeper.WithKeyTable(table)

	// Test keeper.Set, keeper.Modified
	for i, kv := range kvs {
		// require.False(t, keeper.Modified(ctx, kv.key), "keeper.Modified returns true before setting, tc #%d", i)
		require.NotPanics(t, func() { keeper.Set(ctx, kv.key, kv.param) }, "keeper.Set panics, tc #%d", i)
		// require.True(t, keeper.Modified(ctx, kv.key), "keeper.Modified returns false after setting, tc #%d", i)
	}

	// Test keeper.Get, keeper.GetIfExists
	for i, kv := range kvs {
		require.NotPanics(t, func() { keeper.GetIfExists(ctx, "invalid", kv.ptr) }, "keeper.GetIfExists panics when no value exists, tc #%d", i)
		require.Equal(t, kv.zero, indirect(kv.ptr), "keeper.GetIfExists unmarshalls when no value exists, tc #%d", i)
		require.Panics(t, func() { keeper.Get(ctx, "invalid", kv.ptr) }, "invalid keeper.Get not panics when no value exists, tc #%d", i)
		require.Equal(t, kv.zero, indirect(kv.ptr), "invalid keeper.Get unmarshalls when no value exists, tc #%d", i)

		require.NotPanics(t, func() { keeper.GetIfExists(ctx, kv.key, kv.ptr) }, "keeper.GetIfExists panics, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)
		require.NotPanics(t, func() { keeper.Get(ctx, kv.key, kv.ptr) }, "keeper.Get panics, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)

		require.Panics(t, func() { keeper.Get(ctx, "invalid", kv.ptr) }, "invalid keeper.Get not panics when no value exists, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "invalid keeper.Get unmarshalls when no value existt, tc #%d", i)

		require.Panics(t, func() { keeper.Get(ctx, kv.key, nil) }, "invalid keeper.Get not panics when the pointer is nil, tc #%d", i)
		require.Panics(t, func() { keeper.Get(ctx, kv.key, new(invalid)) }, "invalid keeper.Get not panics when the pointer is different type, tc #%d", i)
	}

	// Test store.Get equals keeper.Get
	for i, kv := range kvs {
		bz := store.Get([]byte(kv.key))
		require.NotNil(t, bz, "store.Get() returns nil, tc #%d", i)
		err := amino.UnmarshalJSON(bz, kv.ptr)
		require.NoError(t, err, "cdc.UnmarshalJSON() returns error, tc #%d", i)
		require.Equal(t, kv.param, indirect(kv.ptr), "stored param not equal, tc #%d", i)
	}
}

func TestJSONUpdate(t *testing.T) {
	env := setupTestEnv()
	ctx, keeper := env.ctx, env.keeper
	key := "key"

	space := keeper.WithKeyTable(NewKeyTable(NewParamSetPair(key, paramJSON{}, validateNoOp)))

	var param paramJSON

	err := space.Update(ctx, key, []byte(`{"param1": "10241024"}`))
	require.NoError(t, err)
	space.Get(ctx, key, &param)
	require.Equal(t, paramJSON{10241024, ""}, param)

	err = space.Update(ctx, key, []byte(`{"param2": "helloworld"}`))
	require.NoError(t, err)
	space.Get(ctx, key, &param)
	require.Equal(t, paramJSON{10241024, "helloworld"}, param)

	err = space.Update(ctx, key, []byte(`{"param1": "20482048"}`))
	require.NoError(t, err)
	space.Get(ctx, key, &param)
	require.Equal(t, paramJSON{20482048, "helloworld"}, param)

	err = space.Update(ctx, key, []byte(`{"param1": "40964096", "param2": "goodbyeworld"}`))
	require.NoError(t, err)
	space.Get(ctx, key, &param)
	require.Equal(t, paramJSON{40964096, "goodbyeworld"}, param)
}

type (
	invalid   struct{}
	s         struct{ I int }
	paramJSON struct {
		Param1 int64  `json:"param1,omitempty" yaml:"param1,omitempty"`
		Param2 string `json:"param2,omitempty" yaml:"param2,omitempty"`
	}
)

func validateNoOp(_ interface{}) error     { return nil }
func indirect(ptr interface{}) interface{} { return reflect.ValueOf(ptr).Elem().Interface() }
