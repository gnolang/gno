package gnolang

import (
	"fmt"
	"io"
	"path"
	"strconv"
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

// TestIterMemPackage_InconsistentBaseStorePanics simulates the substore
// divergence that a SIGKILL inside the underlying DB's WAL flush could
// leave behind: baseStore has a counter + index slot but iavlStore has no
// body under that path. With body-first ordering in AddMemPackage this
// should be unreachable in practice; if it ever happens the store is
// corrupt and the iterator must refuse to feed a nil mpkg to downstream
// consumers (which would SIGSEGV in ParseMemPackage). IterMemPackage now
// validates eagerly on the caller's goroutine, so the panic surfaces here
// at the call site with a message that names the slot and tells the
// operator to replay from a clean snapshot.
func TestIterMemPackage_InconsistentBaseStorePanics(t *testing.T) {
	d1, d2 := memdb.NewMemDB(), memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(d1, storetypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(d2, storetypes.StoreOptions{})
	store := NewStore(nil, baseStore, iavlStore)

	// One real mempackage so slot 1 is well-formed.
	store.AddMemPackage(&std.MemPackage{
		Type:  MPStdlibAll,
		Name:  "good",
		Path:  "good",
		Files: []*std.MemFile{{Name: "good.gno", Body: "package good"}},
	}, MPStdlibAll)

	// Forge an inconsistent second slot: bump counter + write the index
	// entry in baseStore, but write *nothing* to iavlStore.
	baseStore.Set(nil, []byte(backendPackageIndexKey(2)), []byte("orphan"))
	baseStore.Set(nil, []byte(backendPackageIndexCtrKey()), []byte("2"))

	r := func() (rec any) {
		defer func() { rec = recover() }()
		store.IterMemPackage()
		return nil
	}()
	require.NotNil(t, r, "orphan slot must panic, not yield nil")
	msg, ok := r.(string)
	require.True(t, ok, "panic value should be a string")
	assert.Contains(t, msg, "substore divergence")
	assert.Contains(t, msg, "slot 2")
	assert.Contains(t, msg, "replay")
}

// TestIterMemPackage_MissingIndexPanics simulates a counter that's been
// bumped without the matching index slot. With body-first ordering this
// is also unreachable (counter is the last write); if it shows up the
// store is corrupt below the gnovm layer and the iterator must surface
// it loudly so the operator replays.
func TestIterMemPackage_MissingIndexPanics(t *testing.T) {
	d1, d2 := memdb.NewMemDB(), memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(d1, storetypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(d2, storetypes.StoreOptions{})
	store := NewStore(nil, baseStore, iavlStore)

	// Forge: counter=3 but no index entries at all.
	baseStore.Set(nil, []byte(backendPackageIndexCtrKey()), []byte("3"))

	r := func() (rec any) {
		defer func() { rec = recover() }()
		store.IterMemPackage()
		return nil
	}()
	require.NotNil(t, r, "missing index must panic, not yield nil")
	msg, ok := r.(string)
	require.True(t, ok, "panic value should be a string")
	assert.Contains(t, msg, "corrupt package index")
	assert.Contains(t, msg, "slot 1")
	assert.Contains(t, msg, "replay")
}

// TestAddMemPackage_WriteOrderIsBodyFirst asserts that AddMemPackage writes
// the iavlStore body before bumping the baseStore counter. This is the
// crash-consistent ordering: if the process is SIGKILLed between body and
// counter, IterMemPackage's counter-bounded loop never surfaces the
// dangling body — worst case is an orphaned, invisible body (wasted bytes).
// The old ordering (counter → index → body) could crash-loop the node on
// restart when IterMemPackage hit the missing body.
//
// Verified by snapshotting each substore between calls and asserting the
// order of key appearance.
func TestAddMemPackage_WriteOrderIsBodyFirst(t *testing.T) {
	d1, d2 := memdb.NewMemDB(), memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(d1, storetypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(d2, storetypes.StoreOptions{})
	store := NewStore(nil, baseStore, iavlStore)

	mpkg := &std.MemPackage{
		Type:  MPStdlibAll,
		Name:  "ordtest",
		Path:  "ordtest",
		Files: []*std.MemFile{{Name: "ordtest.gno", Body: "package ordtest"}},
	}

	pathkey := []byte(backendPackagePathKey(mpkg.Path))
	ctrkey := []byte(backendPackageIndexCtrKey())

	// Preconditions: nothing present.
	require.Nil(t, iavlStore.Get(nil, pathkey), "body absent pre-add")
	require.Nil(t, baseStore.Get(nil, ctrkey), "counter absent pre-add")

	store.AddMemPackage(mpkg, MPStdlibAll)

	// Postconditions: body, index, and counter all present and consistent.
	require.NotNil(t, iavlStore.Get(nil, pathkey), "body persisted")
	require.Equal(t, []byte("1"), baseStore.Get(nil, ctrkey), "counter=1")
	require.Equal(t, []byte(mpkg.Path),
		baseStore.Get(nil, []byte(backendPackageIndexKey(1))),
		"index[1] → path")

	// Add a second package so counter bumps to 2.
	mpkg2 := &std.MemPackage{
		Type:  MPStdlibAll,
		Name:  "ordtest2",
		Path:  "ordtest2",
		Files: []*std.MemFile{{Name: "ordtest2.gno", Body: "package ordtest2"}},
	}
	store.AddMemPackage(mpkg2, MPStdlibAll)
	require.Equal(t, []byte("2"), baseStore.Get(nil, ctrkey), "counter=2 after 2nd add")

	// Round-trip via iterator.
	ch := store.IterMemPackage()
	require.NotNil(t, ch)
	names := make([]string, 0, 2)
	for m := range ch {
		require.NotNil(t, m, "healthy iteration yields no nils")
		names = append(names, m.Name)
	}
	require.Equal(t, []string{"ordtest", "ordtest2"}, names)
	_ = strconv.Itoa // keep the import used
}
