package gnolang

import (
	"fmt"
	"io"
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

// TestGetPackageLazyFileBlocks locks in the lazy file-block behavior:
// loading a multi-file package from the store must NOT eagerly materialize
// its file blocks; a block is materialized only when a function in that
// file is first accessed. Guards against re-introducing eager loading in
// fillPackage.
func TestGetPackageLazyFileBlocks(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)

	// Create and persist a three-file package.
	wrapped := tm2Store.CacheWrap()
	txSt := st.BeginTransaction(wrapped, wrapped, nil, nil)
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "gno.vm/t/multi",
		Store:   txSt,
		Output:  io.Discard,
	})
	m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd,
		Name: "multi",
		Path: "gno.vm/t/multi",
		Files: []*std.MemFile{
			{Name: "a.gno", Body: "package multi\nfunc FA() int { return 1 }"},
			{Name: "b.gno", Body: "package multi\nfunc FB() int { return 2 }"},
			{Name: "c.gno", Body: "package multi\nfunc FC() int { return 3 }"},
		},
	}, true)
	txSt.Write()
	wrapped.Write()

	// Load the package fresh in a new transaction: its file blocks come
	// back from the store as RefValues.
	txSt2 := st.BeginTransaction(tm2Store.CacheWrap(), tm2Store.CacheWrap(), nil, nil)
	pv := txSt2.GetPackage("gno.vm/t/multi", false)
	require.NotNil(t, pv)
	require.Len(t, pv.FNames, 3)

	// fillPackage must leave fBlocksMap empty — no file block loaded yet.
	assert.Empty(t, pv.fBlocksMap, "fillPackage should not eagerly materialize file blocks")

	// Touching one file materializes only that file's block.
	assert.NotNil(t, pv.GetFileBlock(txSt2, "a.gno"))
	assert.Len(t, pv.fBlocksMap, 1, "only the touched file's block should be materialized")
	_, ok := pv.fBlocksMap["a.gno"]
	assert.True(t, ok, "the materialized block should be a.gno")
}

// TestGetPackageSingleFileEagerHydration locks in the single-file guard in
// fillPackage: a package with <= 1 file has no unused file to skip, so it is
// hydrated eagerly on load (preserving master's gas), not lazily. Complements
// TestGetPackageLazyFileBlocks, which covers the multi-file lazy path.
func TestGetPackageSingleFileEagerHydration(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)

	// Create and persist a single-file package.
	wrapped := tm2Store.CacheWrap()
	txSt := st.BeginTransaction(wrapped, wrapped, nil, nil)
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "gno.vm/t/single",
		Store:   txSt,
		Output:  io.Discard,
	})
	m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd,
		Name: "single",
		Path: "gno.vm/t/single",
		Files: []*std.MemFile{
			{Name: "a.gno", Body: "package single\nfunc FA() int { return 1 }"},
		},
	}, true)
	txSt.Write()
	wrapped.Write()

	// Load the package fresh in a new transaction.
	txSt2 := st.BeginTransaction(tm2Store.CacheWrap(), tm2Store.CacheWrap(), nil, nil)
	pv := txSt2.GetPackage("gno.vm/t/single", false)
	require.NotNil(t, pv)
	require.Len(t, pv.FNames, 1)

	// fillPackage must hydrate the one block eagerly (no unused file to skip).
	assert.Len(t, pv.fBlocksMap, 1, "single-file package should be eagerly hydrated in fillPackage")
	_, ok := pv.fBlocksMap["a.gno"]
	assert.True(t, ok, "the hydrated block should be a.gno")
}

// readRecordingStore wraps a tm2 store and counts reads per key, so tests
// can assert which objects were (not) fetched from the underlying store.
type readRecordingStore struct {
	storetypes.Store
	reads map[string]int
}

func (rs *readRecordingStore) Get(gctx *storetypes.GasContext, key []byte) []byte {
	rs.reads[string(key)]++
	return rs.Store.Get(gctx, key)
}

// TestLazyFileBlocksSkipUnusedStoreReads showcases the lazy win end-to-end at
// the store boundary (where I/O gas is charged): loading a multi-file package
// and calling into one file must never read the other files' block objects
// from the underlying store. Under eager fillPackage (pre-lazy), GetPackage
// read all three. The method call also exercises the #4527 panic site:
// DeclaredType.GetValueAt fills the method's nil Parent via
// FuncValue.GetParent, which now lazily loads the file block.
func TestLazyFileBlocksSkipUnusedStoreReads(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)

	// Create and persist a three-file package.
	wrapped := tm2Store.CacheWrap()
	txSt := st.BeginTransaction(wrapped, wrapped, nil, nil)
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "gno.land/r/lazyread",
		Store:   txSt,
		Output:  io.Discard,
	})
	m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd,
		Name: "lazyread",
		Path: "gno.land/r/lazyread",
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: GenGnoModLatest("gno.land/r/lazyread")},
			{Name: "a.gno", Body: "package lazyread\ntype T struct{}\nfunc (T) MA() int { return 7 }\nfunc FA() int { return T{}.MA() }"},
			{Name: "b.gno", Body: "package lazyread\nfunc FB() int { return 2 }"},
			{Name: "c.gno", Body: "package lazyread\nfunc FC() int { return 3 }"},
		},
	}, true)
	txSt.Write()
	wrapped.Write()

	// Reload in a fresh transaction through a read-recording base store.
	spy := &readRecordingStore{Store: tm2Store.CacheWrap(), reads: map[string]int{}}
	txSt2 := st.BeginTransaction(spy, spy, nil, nil)
	pv := txSt2.GetPackage("gno.land/r/lazyread", false)
	require.NotNil(t, pv)
	require.Len(t, pv.FNames, 3)

	// The file-block store keys, from the package's RefValue slots.
	blockKey := map[string]string{}
	for i, fname := range pv.FNames {
		ref, ok := pv.FBlocks[i].(RefValue)
		require.True(t, ok, "file block %q should still be a RefValue after load", fname)
		blockKey[fname] = backendObjectKey(ref.ObjectID)
	}

	// Loading the package must not read any file block from the store.
	for fname, key := range blockKey {
		assert.Zero(t, spy.reads[key], "GetPackage should not read %s's file block", fname)
	}

	// Call into a.gno: FA calls the method T.MA, whose dispatch copies the
	// method FuncValue and fills its nil Parent via GetParent →
	// GetFileBlock, reading a.gno's block from the store on demand.
	m2 := NewMachineWithOptions(MachineOptions{
		PkgPath: "gno.land/r/lazyread",
		Store:   txSt2,
		Output:  io.Discard,
	})
	m2.SetActivePackage(pv)
	res := m2.Eval(m2.MustParseExpr("FA()"))
	require.Len(t, res, 1)
	assert.Equal(t, int64(7), res[0].GetInt()) // FA() → T{}.MA() → 7, per a.gno above

	// Only a.gno's block was read; the unused files' blocks never were.
	assert.Positive(t, spy.reads[blockKey["a.gno"]], "the called method's file block should be read")
	assert.Zero(t, spy.reads[blockKey["b.gno"]], "unused file block b.gno must not be read")
	assert.Zero(t, spy.reads[blockKey["c.gno"]], "unused file block c.gno must not be read")
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

func TestDeleteMemPackageClearsStaleBlobsOnReAdd(t *testing.T) {
	newStore := func() *defaultStore {
		d1, d2 := memdb.NewMemDB(), memdb.NewMemDB()
		d1s := dbadapter.StoreConstructor(d1, storetypes.StoreOptions{})
		d2s := dbadapter.StoreConstructor(d2, storetypes.StoreOptions{})
		return NewStore(nil, d1s, d2s)
	}
	const pkgPath = "gno.land/r/demo/foo"
	withTests := func() *std.MemPackage {
		return &std.MemPackage{
			Type: MPUserAll, Name: "foo", Path: pkgPath,
			Files: []*std.MemFile{
				{Name: "foo.gno", Body: "package foo\n"},
				{Name: "foo_test.gno", Body: "package foo\n"},
			},
		}
	}
	prodOnly := func() *std.MemPackage {
		return &std.MemPackage{
			Type: MPUserAll, Name: "foo", Path: pkgPath,
			Files: []*std.MemFile{
				{Name: "foo.gno", Body: "package foo\n"},
			},
		}
	}

	// DeleteMemPackage removes both the prod blob and the #allbutprod sibling.
	st := newStore()
	st.AddMemPackage(withTests(), MPUserAll)
	require.NotNil(t, st.GetMemFile(pkgPath, "foo_test.gno"))
	st.DeleteMemPackage(pkgPath)
	assert.Nil(t, st.GetMemPackage(pkgPath))
	assert.Nil(t, st.GetMemPackageAll(pkgPath))
	assert.Nil(t, st.GetMemFile(pkgPath, "foo_test.gno"))

	// Re-add idempotency (mirrors the keeper clearing a private package before
	// redeploy): dropping the test file must not leave a stale #allbutprod sibling.
	st = newStore()
	st.AddMemPackage(withTests(), MPUserAll)
	st.DeleteMemPackage(pkgPath)
	st.AddMemPackage(prodOnly(), MPUserAll)
	all := st.GetMemPackageAll(pkgPath)
	require.NotNil(t, all)
	require.Len(t, all.Files, 1)
	assert.Equal(t, "foo.gno", all.Files[0].Name)
	assert.Nil(t, st.GetMemFile(pkgPath, "foo_test.gno"), "stale test file must not survive re-add")
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
		name := LastPathElement(pkg)
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

func TestFindByPrefixDeDupesSplitPackages(t *testing.T) {
	d1, d2 := memdb.NewMemDB(), memdb.NewMemDB()
	d1s := dbadapter.StoreConstructor(d1, storetypes.StoreOptions{})
	d2s := dbadapter.StoreConstructor(d2, storetypes.StoreOptions{})
	store := NewStore(nil, d1s, d2s)

	add := func(name string, files ...*std.MemFile) {
		store.AddMemPackage(&std.MemPackage{
			Type:  MPUserAll,
			Name:  name,
			Path:  "gno.land/r/demo/" + name,
			Files: files,
		}, MPUserAll)
	}
	// alpha: prod + test -> prod blob + #allbutprod sibling (must list once, not twice).
	add("alpha", &std.MemFile{Name: "alpha.gno", Body: "package alpha\n"},
		&std.MemFile{Name: "alpha_test.gno", Body: "package alpha\n"})
	// beta: test-only -> #allbutprod sibling only (empty prod; must still list once).
	add("beta", &std.MemFile{Name: "beta_test.gno", Body: "package beta\n"})
	// gamma: prod only -> prod blob only.
	add("gamma", &std.MemFile{Name: "gamma.gno", Body: "package gamma\n"})

	var got []string
	store.FindPathsByPrefix("gno.land")(func(p string) bool {
		got = append(got, p)
		return true
	})
	require.Equal(t, []string{
		"gno.land/r/demo/alpha",
		"gno.land/r/demo/beta",
		"gno.land/r/demo/gamma",
	}, got)

	// A '#'-containing prefix (impossible in a valid package path, reachable
	// from raw query input) ranges over alpha's #allbutprod sibling key; the
	// trimmed path "gno.land/r/demo/alpha" does not carry the prefix and must
	// not be yielded.
	for _, prefix := range []string{"gno.land/r/demo/alpha#", "gno.land/r/demo/alpha#allbutprod"} {
		store.FindPathsByPrefix(prefix)(func(p string) bool {
			t.Fatalf("prefix %q must yield nothing, got %q", prefix, p)
			return true
		})
	}
}

// TestMemPackageTestBlobExcludedFromConsensusStore asserts that a package's
// test/filetest files (#allbutprod sibling) are written to the non-merkleized
// baseStore and NOT to the merkleized iavlStore, so they never enter the
// consensus AppHash — while the production blob stays in iavlStore and the full
// package (prod + test) remains reconstructable via GetMemPackageAll.
func TestMemPackageTestBlobExcludedFromConsensusStore(t *testing.T) {
	baseDB, iavlDB := memdb.NewMemDB(), memdb.NewMemDB()
	base := dbadapter.StoreConstructor(baseDB, storetypes.StoreOptions{})
	iavl := dbadapter.StoreConstructor(iavlDB, storetypes.StoreOptions{})
	store := NewStore(nil, base, iavl)

	path := "gno.land/r/demo/split"
	store.AddMemPackage(&std.MemPackage{
		Type: MPUserAll,
		Name: "split",
		Path: path,
		Files: []*std.MemFile{
			{Name: "split.gno", Body: "package split\n\nfunc Prod() int { return 1 }\n"},
			{Name: "split_test.gno", Body: "package split\n\nfunc TestX() {}\n"},
		},
	}, MPUserAll)

	prodKey := []byte(backendPackagePathKey(path))
	testKey := []byte(backendPackageAllButProdKey(path))

	// Prod blob: iavlStore only (consensus state).
	require.True(t, iavl.Has(nil, prodKey), "prod blob must be in the merkleized iavlStore")
	require.False(t, base.Has(nil, prodKey), "prod blob must not leak into baseStore")
	// Test/filetest blob: baseStore only (excluded from consensus AppHash).
	require.True(t, base.Has(nil, testKey), "test blob must be in the non-merkleized baseStore")
	require.False(t, iavl.Has(nil, testKey), "test blob must NOT enter the merkleized iavlStore")

	// Prod-only view excludes test files; full view includes them.
	prod := store.GetMemPackage(path)
	require.NotNil(t, prod)
	require.Equal(t, []string{"split.gno"}, memFileNames(prod), "GetMemPackage must return prod files only")

	all := store.GetMemPackageAll(path)
	require.NotNil(t, all)
	require.Equal(t, []string{"split.gno", "split_test.gno"}, memFileNames(all), "GetMemPackageAll must reconstruct prod + test files")
}

func memFileNames(mpkg *std.MemPackage) []string {
	names := make([]string, len(mpkg.Files))
	for i, f := range mpkg.Files {
		names[i] = f.Name
	}
	return names
}
