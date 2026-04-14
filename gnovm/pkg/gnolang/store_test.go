package gnolang

import (
	"fmt"
	"io"
	"path"
	"sort"
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
	txSt := st.BeginTransaction(wrappedTm2Store, wrappedTm2Store, nil, nil)
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

	// Check that hello.A is set in txSt.
	stA := txSt.GetType("gno.vm/t/hello.A")
	assert.NotNil(t, stA)
	assert.Empty(t, stA.(*DeclaredType).Methods)

	// use write on the stores
	txSt.Write()
	wrappedTm2Store.Write()

	// mem package should exist and be ==.
	res := st.GetMemPackage("gno.vm/t/hello")
	assert.NotNil(t, res)
	assert.Equal(t, txSt.GetMemPackage("gno.vm/t/hello"), res)
	helloA := st.GetType("gno.vm/t/hello.A")
	assert.NotNil(t, helloA)
	// Normalize nil vs empty slice: amino-unmarshal of an empty repeated field
	// returns nil, while the in-memory type retains nil from construction.
	// Both represent "no methods" and are semantically equivalent.
	if helloA.(*DeclaredType).Methods == nil {
		helloA.(*DeclaredType).Methods = []TypedValue{}
	}
	if stA.(*DeclaredType).Methods == nil {
		stA.(*DeclaredType).Methods = []TypedValue{}
	}
	assert.Equal(t, stA, helloA)
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
	destStoreTx := destStore.BeginTransaction(nil, nil, nil, nil) // CopyFromCachedStore requires a tx store.
	CopyFromCachedStore(destStoreTx, cachedStore, c1s, c2s)
	destStoreTx.Write()

	assert.Equal(t, c1, d1, "cached baseStore and dest baseStore should match")
	assert.Equal(t, c2, d2, "cached iavlStore and dest iavlStore should match")
	assert.Equal(t, cachedStore.cacheNodes, destStore.cacheNodes, "cacheNodes should match")
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

// TestStdlibCacheSharedWithoutMutex demonstrates bug H6:
// stdlibKeyBytes is a plain map[string][]byte shared by reference across
// transaction stores (BeginTransaction copies the pointer, not the map).
// There is no mutex protecting it.
//
// In practice, PopulateStdlibCache is only called at node startup/genesis
// (before transactions run), and after that the map is read-only. So this
// is a defensive-programming concern rather than a runtime race under the
// current ABCI model. However, the lack of any synchronization makes the
// code fragile: any future write path (e.g., governance stdlib upgrade,
// hot reload) would introduce a fatal "concurrent map read and map write".
func TestStdlibCacheSharedWithoutMutex(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})

	// Seed some data in the backing store so populateStdlibCache has
	// something to write into the cache.
	baseStore.Set(nil, []byte("tid:strings.Builder"), []byte("builder_bytes"))
	baseStore.Set(nil, []byte("tid:strings.Replacer"), []byte("replacer_bytes"))

	parentStore := NewStore(nil, baseStore, baseStore)
	parentStore.PopulateStdlibCache([]string{"strings"})

	// BeginTransaction shares the same stdlibKeyBytes map reference.
	tx1 := parentStore.BeginTransaction(nil, nil, nil, nil)
	tx2 := parentStore.BeginTransaction(nil, nil, nil, nil)
	ds0 := parentStore
	ds1 := tx1.(transactionStore).defaultStore
	ds2 := tx2.(transactionStore).defaultStore

	// All three point to the same underlying map — no copy-on-write.
	assert.Equal(t,
		fmt.Sprintf("%p", ds0.stdlibKeyBytes),
		fmt.Sprintf("%p", ds1.stdlibKeyBytes),
		"tx1 should share the same stdlibKeyBytes map as parent")
	assert.Equal(t,
		fmt.Sprintf("%p", ds0.stdlibKeyBytes),
		fmt.Sprintf("%p", ds2.stdlibKeyBytes),
		"tx2 should share the same stdlibKeyBytes map as parent")

	// Prove the sharing: parent, tx1, and tx2 all see the same entries.
	require.NotNil(t, ds0.stdlibKeyBytes["tid:strings.Builder"])
	require.NotNil(t, ds1.stdlibKeyBytes["tid:strings.Builder"])
	require.NotNil(t, ds2.stdlibKeyBytes["tid:strings.Builder"])

	// Prove writes through one reference are visible through the others —
	// this is the root cause of the race. A concurrent CheckTx reading the
	// map while DeliverTx's PopulateStdlibCache writes to it will crash
	// with "concurrent map read and map write" (fatal in Go 1.19+).
	ds1.stdlibKeyBytes["tid:strings.NewType"] = []byte("new_type_bytes")
	assert.NotNil(t, ds0.stdlibKeyBytes["tid:strings.NewType"],
		"write through tx1 visible in parent — same map, no copy-on-write")
	assert.NotNil(t, ds2.stdlibKeyBytes["tid:strings.NewType"],
		"write through tx1 visible in tx2 — same map, no copy-on-write")
}

// TestPopulateStdlibCacheMissesLocalTypes demonstrates bug H8:
// populateStdlibCache uses the iterator range ["tid:<path>.", "tid:<path>/")
// to capture stdlib type keys. This misses locally-declared types whose
// TypeID has the form "<path>[<loc>].<name>" (e.g., "strings[1:2].myType"),
// because '[' (0x5B) > '/' (0x2F) puts them outside the range.
//
// In practice, current stdlib packages only declare types at package level
// (ParentLoc is zero), producing "tid:path.Name" which IS captured. So this
// is currently a latent bug — it would only manifest if a stdlib package
// defined a type inside a function body. The range logic is still incorrect
// and should use PrefixEndBytes for correctness.
func TestPopulateStdlibCacheMissesLocalTypes(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})

	// Seed the backing store with type keys in both formats:
	//   Package-level type:  tid:<path>.<name>       (e.g., "tid:strings.Builder")
	//   Local/block type:    tid:<path>[<loc>].<name> (e.g., "tid:strings[1:2].localType")
	//
	// DeclaredTypeID produces the second form when ParentLoc is non-zero.
	typeKeys := map[string][]byte{
		// Package-level types — should be cached.
		"tid:strings.Builder":  []byte("builder_bytes"),
		"tid:strings.Replacer": []byte("replacer_bytes"),
		"tid:math/big.Int":     []byte("bigint_bytes"),
		// Local types (non-zero Location) — MISSED by current range.
		"tid:strings[1:2].localType":    []byte("local_type_bytes"),
		"tid:strings[3:15].anotherType": []byte("another_local_bytes"),
		"tid:math/big[7:9].helper":      []byte("helper_bytes"),
	}

	for k, v := range typeKeys {
		baseStore.Set(nil, []byte(k), v)
	}

	// Create a defaultStore and populate the stdlib cache.
	ds := &defaultStore{
		baseStore:      baseStore,
		iavlStore:      baseStore,
		stdlibKeyBytes: make(map[string][]byte),
	}
	ds.populateStdlibCache([]string{"strings", "math/big"}, baseStore)

	// Collect what was cached.
	var cached []string
	for k := range ds.stdlibKeyBytes {
		cached = append(cached, k)
	}
	sort.Strings(cached)

	// Package-level types are cached.
	assert.Contains(t, ds.stdlibKeyBytes, "tid:strings.Builder", "package-level type should be cached")
	assert.Contains(t, ds.stdlibKeyBytes, "tid:strings.Replacer", "package-level type should be cached")
	assert.Contains(t, ds.stdlibKeyBytes, "tid:math/big.Int", "package-level type should be cached")

	// BUG: Local types are NOT cached because '[' (0x5B) > '/' (0x2F)
	// puts "tid:strings[..." outside the range ["tid:strings.", "tid:strings/").
	assert.NotContains(t, ds.stdlibKeyBytes, "tid:strings[1:2].localType",
		"local type MISSED: '[' (0x5B) is outside the iterator range ending at '/' (0x2F)")
	assert.NotContains(t, ds.stdlibKeyBytes, "tid:strings[3:15].anotherType",
		"local type MISSED: '[' (0x5B) is outside the iterator range ending at '/' (0x2F)")
	assert.NotContains(t, ds.stdlibKeyBytes, "tid:math/big[7:9].helper",
		"local type MISSED: '[' (0x5B) is outside the iterator range ending at '/' (0x2F)")

	t.Logf("cached keys: %v", cached)
	t.Logf("expected 6 keys cached, got %d — local types are missing", len(cached))
}
