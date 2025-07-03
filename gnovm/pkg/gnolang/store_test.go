package gnolang

import (
	"fmt"
	"io"
	"path"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionStore(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})

	st := NewStore(nil, tm2Store, tm2Store)
	wrappedTm2Store := tm2Store.CacheWrap()
	txSt := st.BeginTransaction(wrappedTm2Store, wrappedTm2Store, nil)
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "gno.vm/t/hello",
		Store:   txSt,
		Output:  io.Discard,
	})
	_, pv := m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd,
		Name: "hello",
		Path: "gno.vm/t/hello",
		Files: []*std.MemFile{
			{Name: "hello.gno", Body: "package hello; func main() { println(A(11)); }; type A int"},
		},
	}, true)
	m.SetActivePackage(pv)
	m.RunMain()

	// mem package should only exist in txSt
	// (check both memPackage and types - one is stored directly in the db,
	// the other uses txlog)
	assert.Nil(t, st.GetMemPackage("gno.vm/t/hello"))
	assert.NotNil(t, txSt.GetMemPackage("gno.vm/t/hello"))
	assert.PanicsWithValue(t, "unexpected type with id gno.vm/t/hello.A", func() { st.GetType("gno.vm/t/hello.A") })
	assert.NotNil(t, txSt.GetType("gno.vm/t/hello.A"))

	// use write on the stores
	txSt.Write()
	wrappedTm2Store.Write()

	// mem package should exist and be ==.
	res := st.GetMemPackage("gno.vm/t/hello")
	assert.NotNil(t, res)
	assert.Equal(t, txSt.GetMemPackage("gno.vm/t/hello"), res)
	helloA := st.GetType("gno.vm/t/hello.A")
	assert.NotNil(t, helloA)
	assert.Equal(t, txSt.GetType("gno.vm/t/hello.A"), helloA)
}

func TestTransactionStore_blockedMethods(t *testing.T) {
	// These methods should panic as they modify store settings, which should
	// only be changed in the root store.
	assert.Panics(t, func() { transactionStore{}.SetPackageGetter(nil) })
	assert.Panics(t, func() { transactionStore{}.SetNativeResolver(nil) })
}

func TestCopyFromCachedStore(t *testing.T) {
	// Create cached store, with a type and a mempackage.
	c1 := memdb.NewMemDB()
	c1s := dbadapter.StoreConstructor(c1, storetypes.StoreOptions{})
	c2 := memdb.NewMemDB()
	c2s := dbadapter.StoreConstructor(c2, storetypes.StoreOptions{})
	cachedStore := NewStore(nil, c1s, c2s)
	cachedStore.SetType(&DeclaredType{
		PkgPath: "io",
		Name:    "Reader",
		Base:    BoolType,
	})
	cachedStore.AddMemPackage(&std.MemPackage{
		Type: MPStdlibAll,
		Name: "math",
		Path: "math",
		Files: []*std.MemFile{
			{Name: "math.gno", Body: "package math"},
		},
	}, MPAnyAll)

	// Create dest store and copy.
	d1, d2 := memdb.NewMemDB(), memdb.NewMemDB()
	d1s := dbadapter.StoreConstructor(d1, storetypes.StoreOptions{})
	d2s := dbadapter.StoreConstructor(d2, storetypes.StoreOptions{})
	destStore := NewStore(nil, d1s, d2s)
	destStoreTx := destStore.BeginTransaction(nil, nil, nil) // CopyFromCachedStore requires a tx store.
	CopyFromCachedStore(destStoreTx, cachedStore, c1s, c2s)
	destStoreTx.Write()

	assert.Equal(t, c1, d1, "cached baseStore and dest baseStore should match")
	assert.Equal(t, c2, d2, "cached iavlStore and dest iavlStore should match")
	assert.Equal(t, cachedStore.cacheTypes, destStore.cacheTypes, "cacheTypes should match")
}

func TestFindByPrefix(t *testing.T) {
	stdlibs := []string{"abricot", "balloon", "call", "dingdong", "gnocchi"}
	pkgs := []string{
		"fruits.org/t/abricot",
		"fruits.org/t/abricot/fraise",
		"fruits.org/t/fraise",
	}

	cases := []struct {
		Prefix   string
		Limit    int
		Expected []string
	}{
		{"", 100, append(stdlibs, pkgs...)}, // no prefix == everything
		{"fruits.org", 100, pkgs},
		{"fruits.org/t/abricot", 100, []string{
			"fruits.org/t/abricot", "fruits.org/t/abricot/fraise",
		}},
		{"fruits.org/t/abricot/", 100, []string{
			"fruits.org/t/abricot/fraise",
		}},
		{"fruits", 100, pkgs}, // no stdlibs (prefixed by "_" keys)
		{"_", 100, stdlibs},
		{"_/a", 100, []string{"abricot"}},
		// special case
		{string([]byte{255}), 100, []string{}}, // using 255 as prefix, should not panic
		{string([]byte{0}), 100, []string{}},   // using 0 as prefix, should not panic
		// testing iter seq
		{"_", 0, []string{}},
		{"_", 2, stdlibs[:2]},
	}

	// Create cached store, with a type and a mempackage.
	d1, d2 := memdb.NewMemDB(), memdb.NewMemDB()
	d1s := dbadapter.StoreConstructor(d1, storetypes.StoreOptions{})
	d2s := dbadapter.StoreConstructor(d2, storetypes.StoreOptions{})
	store := NewStore(nil, d1s, d2s)

	// Add stdlibs
	for _, lib := range stdlibs {
		store.AddMemPackage(&std.MemPackage{
			Type: MPStdlibAll,
			Name: lib,
			Path: lib,
			Files: []*std.MemFile{
				{Name: lib + ".gno", Body: "package " + lib},
			},
		}, MPStdlibAll)
	}

	// Add pkgs
	for _, pkg := range pkgs {
		name := path.Base(pkg)
		store.AddMemPackage(&std.MemPackage{
			Type: MPUserProd,
			Name: name,
			Path: pkg,
			Files: []*std.MemFile{
				{Name: name + ".gno", Body: "package " + name},
			},
		}, MPUserProd)
	}

	for _, tc := range cases {
		name := fmt.Sprintf("%s:limit(%d)", tc.Prefix, tc.Limit)
		t.Run(name, func(t *testing.T) {
			t.Logf("lookup prefix: %q, limit: %d", tc.Prefix, tc.Limit)

			paths := []string{}

			var count int
			yield := func(path string) bool {
				if count >= tc.Limit {
					return false
				}

				paths = append(paths, path)
				count++
				return true // continue
			}

			store.FindPathsByPrefix(tc.Prefix)(yield) // find stdlibs
			require.Equal(t, tc.Expected, paths)
		})
	}
}
