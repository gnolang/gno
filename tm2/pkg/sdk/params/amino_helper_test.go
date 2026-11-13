package params

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

// tiParamsA and tiParamsB have the same Go shape but different JSON names, so
// a memo that confused them would read and write the wrong store keys.
type tiParamsA struct {
	Max    int64    `json:"max"`
	Name   string   `json:"name"`
	Denoms []string `json:"denoms"`
}

type tiParamsB struct {
	Max    int64    `json:"max_b"`
	Name   string   `json:"name_b"`
	Denoms []string `json:"denoms_b"`
}

func jsonNames(ti *amino.TypeInfo) []string {
	names := make([]string, len(ti.Fields))
	for i, f := range ti.Fields {
		names[i] = f.JSONName
	}
	return names
}

func TestStructTypeInfoMatchesAmino(t *testing.T) {
	typeInfos.Clear()
	for _, rt := range []reflect.Type{
		reflect.TypeFor[tiParamsA](),
		reflect.TypeFor[tiParamsB](),
	} {
		fresh, err := amino.GetTypeInfo(rt)
		require.NoError(t, err)
		got, err := structTypeInfo(rt)
		require.NoError(t, err)
		again, err := structTypeInfo(rt)
		require.NoError(t, err)

		assert.Same(t, got, again, "second call must hit the memo")
		assert.Equal(t, rt, got.Type)
		assert.Equal(t, jsonNames(fresh), jsonNames(got))
	}
	a, _ := structTypeInfo(reflect.TypeFor[tiParamsA]())
	b, _ := structTypeInfo(reflect.TypeFor[tiParamsB]())
	assert.Equal(t, []string{"max", "name", "denoms"}, jsonNames(a))
	assert.Equal(t, []string{"max_b", "name_b", "denoms_b"}, jsonNames(b))
}

// The memoized *amino.TypeInfo is shared by every caller, so a struct
// round trip must leave it exactly as amino built it.
func TestStructTypeInfoNotMutatedByRoundTrip(t *testing.T) {
	typeInfos.Clear()
	env := newTypeInfoTestEnv(t)
	rt := reflect.TypeFor[tiParamsA]()
	before, err := structTypeInfo(rt)
	require.NoError(t, err)
	snapshot := jsonNames(before)

	in := tiParamsA{Max: 7, Name: "x", Denoms: []string{"ugnot"}}
	env.keeper.SetStruct(env.ctx, "ti_a:p", in)
	var out tiParamsA
	env.keeper.GetStruct(env.ctx, "ti_a:p", &out)
	require.Equal(t, in, out)

	after, err := structTypeInfo(rt)
	require.NoError(t, err)
	assert.Same(t, before, after)
	assert.Equal(t, snapshot, jsonNames(after))
}

// Parallel GetStruct/SetStruct on identical and distinct types, each
// goroutine on its own store, all sharing the memo. Run with -race.
func TestStructTypeInfoConcurrent(t *testing.T) {
	typeInfos.Clear()
	const workers = 32
	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			env := newTypeInfoTestEnv(t)
			for i := range 50 {
				name := fmt.Sprintf("w%d-%d", w, i)
				if (w+i)%2 == 0 {
					in := tiParamsA{Max: int64(i), Name: name, Denoms: []string{name}}
					env.keeper.SetStruct(env.ctx, "ti_a:p", in)
					var out tiParamsA
					env.keeper.GetStruct(env.ctx, "ti_a:p", &out)
					assert.Equal(t, in, out)
				} else {
					in := tiParamsB{Max: int64(i), Name: name, Denoms: []string{name}}
					env.keeper.SetStruct(env.ctx, "ti_b:p", in)
					var out tiParamsB
					env.keeper.GetStruct(env.ctx, "ti_b:p", &out)
					assert.Equal(t, in, out)
				}
			}
		})
	}
	wg.Wait()

	n := 0
	typeInfos.Range(func(_, _ any) bool { n++; return true })
	assert.Equal(t, 2, n, "one entry per distinct param struct type")
}

type typeInfoTestEnv struct {
	ctx    sdk.Context
	keeper ParamsKeeper
}

func newTypeInfoTestEnv(tb testing.TB) typeInfoTestEnv {
	tb.Helper()
	db := memdb.NewMemDB()
	key := store.NewStoreKey("paramsCapKey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(key, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	pk := NewParamsKeeper(key)
	for _, m := range []string{"ti_a", "ti_b"} {
		pk.Register(m, NewDummyKeeper(pk.ForModule(m)))
	}
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{Height: 1, ChainID: "test-chain-id"}, log.NewNoopLogger())
	return typeInfoTestEnv{ctx: ctx, keeper: pk}
}

// BenchmarkGetStruct is the per-transaction cost of reading a module's param
// struct, which the auth and vm ante handlers do several times per tx.
func BenchmarkGetStruct(b *testing.B) {
	env := newTypeInfoTestEnv(b)
	in := tiParamsA{Max: 7, Name: "x", Denoms: []string{"ugnot"}}
	env.keeper.SetStruct(env.ctx, "ti_a:p", in)
	b.ReportAllocs()
	for b.Loop() {
		var out tiParamsA
		env.keeper.GetStruct(env.ctx, "ti_a:p", &out)
	}
}
