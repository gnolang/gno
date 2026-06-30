package bptree

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func newTestStore() *Store {
	db := memdb.NewMemDB()
	return StoreConstructor(db, types.StoreOptions{}).(*Store)
}

func TestStore_SetGetHasDelete(t *testing.T) {
	st := newTestStore()

	st.Set(nil, []byte("hello"), []byte("world"))
	v := st.Get(nil, []byte("hello"))
	if !bytes.Equal(v, []byte("world")) {
		t.Fatalf("Get = %q, want 'world'", v)
	}
	if !st.Has(nil, []byte("hello")) {
		t.Fatalf("Has = false")
	}

	st.Delete(nil, []byte("hello"))
	v = st.Get(nil, []byte("hello"))
	if v != nil {
		t.Fatalf("Get after delete = %q, want nil", v)
	}
}

func TestStore_Commit(t *testing.T) {
	st := newTestStore()
	st.Set(nil, []byte("a"), []byte("1"))
	st.Set(nil, []byte("b"), []byte("2"))

	cid := st.Commit()
	if cid.Version != 1 {
		t.Fatalf("version = %d, want 1", cid.Version)
	}
	if cid.Hash == nil {
		t.Fatalf("hash is nil")
	}
}

func TestStore_LastCommitID(t *testing.T) {
	st := newTestStore()
	st.Set(nil, []byte("x"), []byte("y"))
	cid := st.Commit()

	last := st.LastCommitID()
	if last.Version != cid.Version {
		t.Fatalf("version mismatch")
	}
	if !bytes.Equal(last.Hash, cid.Hash) {
		t.Fatalf("hash mismatch")
	}
}

func TestStore_LoadVersion(t *testing.T) {
	db := memdb.NewMemDB()
	// Use KeepRecent > 0 to prevent auto-pruning of old versions
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("k1"), []byte("v1"))
	st.Commit() // v1

	st.Set(nil, []byte("k2"), []byte("v2"))
	st.Commit() // v2

	// New store from same DB
	st2 := StoreConstructor(db, opts).(*Store)
	if err := st2.LoadLatestVersion(); err != nil {
		t.Fatalf("LoadLatestVersion: %v", err)
	}
	err := st2.LoadVersion(1)
	if err != nil {
		t.Fatalf("LoadVersion(1): %v", err)
	}

	v := st2.Get(nil, []byte("k1"))
	if !bytes.Equal(v, []byte("v1")) {
		t.Fatalf("v1 k1 = %q", v)
	}
	// k2 should not exist in v1
	v = st2.Get(nil, []byte("k2"))
	if v != nil {
		t.Fatalf("v1 k2 should be nil")
	}
}

func TestStore_LoadLatestVersion(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("a"), []byte("b"))
	st.Commit()
	st.Set(nil, []byte("c"), []byte("d"))
	st.Commit()

	st2 := StoreConstructor(db, opts).(*Store)
	err := st2.LoadLatestVersion()
	if err != nil {
		t.Fatalf("LoadLatestVersion: %v", err)
	}
	last := st2.LastCommitID()
	if last.Version != 2 {
		t.Fatalf("latest version = %d", last.Version)
	}
}

func TestStore_Iterator(t *testing.T) {
	st := newTestStore()
	st.Set(nil, []byte("a"), []byte("1"))
	st.Set(nil, []byte("b"), []byte("2"))
	st.Set(nil, []byte("c"), []byte("3"))

	itr := st.Iterator(nil, nil, nil)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != 3 || keys[0] != "a" || keys[2] != "c" {
		t.Fatalf("Iterator keys = %v", keys)
	}
}

func TestStore_ReverseIterator(t *testing.T) {
	st := newTestStore()
	st.Set(nil, []byte("a"), []byte("1"))
	st.Set(nil, []byte("b"), []byte("2"))
	st.Set(nil, []byte("c"), []byte("3"))

	itr := st.ReverseIterator(nil, nil, nil)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != 3 || keys[0] != "c" || keys[2] != "a" {
		t.Fatalf("ReverseIterator keys = %v", keys)
	}
}

func TestStore_CacheWrap(t *testing.T) {
	st := newTestStore()
	st.Set(nil, []byte("a"), []byte("1"))

	cw := st.CacheWrap()
	// Should be able to read through cache
	v := cw.Get(nil, []byte("a"))
	if !bytes.Equal(v, []byte("1")) {
		t.Fatalf("CacheWrap Get = %q", v)
	}

	// Write to cache, verify not in base
	cw.Set(nil, []byte("b"), []byte("2"))
	v = st.Get(nil, []byte("b"))
	if v != nil {
		t.Fatalf("base should not see cache write")
	}

	// Flush cache
	cw.Write()
	v = st.Get(nil, []byte("b"))
	if !bytes.Equal(v, []byte("2")) {
		t.Fatalf("base after Write = %q", v)
	}
}

func TestStore_VersionExists(t *testing.T) {
	st := newTestStore()
	st.Set(nil, []byte("a"), []byte("1"))
	st.Commit()

	if !st.VersionExists(1) {
		t.Fatalf("v1 should exist")
	}
	if st.VersionExists(2) {
		t.Fatalf("v2 should not exist")
	}
}

func TestStore_GetImmutable(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)
	st.Set(nil, []byte("a"), []byte("1"))
	st.Commit()

	immSt, err := st.GetImmutable(1)
	if err != nil {
		t.Fatalf("GetImmutable: %v", err)
	}
	v := immSt.Get(nil, []byte("a"))
	if !bytes.Equal(v, []byte("1")) {
		t.Fatalf("immutable Get = %q", v)
	}
}

func TestStore_ExpectedDepth100(t *testing.T) {
	st := newTestStore()
	if d := st.ExpectedGetReadDepth100(); d != 100 {
		t.Fatalf("empty depth100 = %d, want 100 (one op floor)", d)
	}

	for i := 0; i < 100; i++ {
		st.Set(nil, []byte{byte(i)}, []byte("v"))
	}
	// 100 keys: bits.Len64(100)=7 -> 140 (1.4 node reads), and all three
	// depths agree for a B+ tree.
	g, s2, w := st.ExpectedGetReadDepth100(), st.ExpectedSetReadDepth100(), st.ExpectedWriteDepth100()
	if g != 140 || s2 != g || w != g {
		t.Fatalf("depth100 = %d/%d/%d, want 140 each", g, s2, w)
	}
}

func TestStore_ProofDecoder(t *testing.T) {
	// Register and verify the proof decoder works
	prt := merkle.NewProofRuntime()
	RegisterProofRuntime(prt)

	// The decoder should be registered (we can't easily test decoding
	// without a full proof, but we can verify registration doesn't panic)
}

func TestStore_ImmutableSetPanics(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)
	st.Set(nil, []byte("a"), []byte("1"))
	st.Commit()

	immSt, _ := st.GetImmutable(1)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("Set on immutable store should panic")
		}
	}()
	immSt.Set(nil, []byte("b"), []byte("2"))
}

func TestStore_ImmutableIterator(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("a"), []byte("1"))
	st.Set(nil, []byte("b"), []byte("2"))
	st.Commit() // v1

	st.Set(nil, []byte("c"), []byte("3")) // in working tree, not yet saved

	// Immutable at v1 should only see a, b — not c
	immSt, _ := st.GetImmutable(1)
	itr := immSt.Iterator(nil, nil, nil)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("immutable iterator keys = %v, want [a b]", keys)
	}
}

func TestStore_MultiCommitSnapshotIsolation(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("a"), []byte("v1"))
	st.Commit() // v1

	st.Set(nil, []byte("a"), []byte("v2"))
	st.Commit() // v2

	st.Set(nil, []byte("a"), []byte("v3"))
	st.Commit() // v3

	// Each version should return its own value
	for v, expected := range map[int64]string{1: "v1", 2: "v2", 3: "v3"} {
		immSt, err := st.GetImmutable(v)
		if err != nil {
			t.Fatalf("GetImmutable(%d): %v", v, err)
		}
		val := immSt.Get(nil, []byte("a"))
		if !bytes.Equal(val, []byte(expected)) {
			t.Fatalf("v%d: a = %q, want %q", v, val, expected)
		}
	}
}

func TestStore_StoreOptions(t *testing.T) {
	st := newTestStore()
	opts := st.GetStoreOptions()
	if opts.Immutable {
		t.Fatalf("should not be immutable by default")
	}
	opts.Immutable = true
	st.SetStoreOptions(opts)
	if !st.GetStoreOptions().Immutable {
		t.Fatalf("options not saved")
	}
}

// Hardfork chains start at InitialHeight > 1: the first Commit must land at
// the initial version, pruning must cross the boundary without touching
// sub-initial versions, and a cold reload must recover the same state.
func TestStore_SetInitialVersion(t *testing.T) {
	db := memdb.NewMemDB()
	st := StoreConstructor(db, types.StoreOptions{
		PruningOptions: types.NewPruningOptions(2, 0), // KeepRecent=2
	}).(*Store)

	st.SetInitialVersion(100)
	st.Set(nil, []byte("k"), []byte("v100"))
	id := st.Commit()
	if id.Version != 100 {
		t.Fatalf("first commit at v%d, want 100", id.Version)
	}

	// Commit through 106: prunes cross the initial-version boundary
	// (toRelease < 100 must no-op, then real prunes follow).
	for v := int64(101); v <= 106; v++ {
		st.Set(nil, []byte("k"), []byte(fmt.Sprintf("v%d", v)))
		if id = st.Commit(); id.Version != v {
			t.Fatalf("commit at v%d, want %d", id.Version, v)
		}
	}
	if st.tree.VersionExists(100) {
		t.Fatal("v100 still retained; want pruned (KeepRecent=2)")
	}
	for v := int64(104); v <= 106; v++ {
		if !st.tree.VersionExists(v) {
			t.Fatalf("v%d missing; want retained", v)
		}
	}

	// Cold reload from the same DB: discoverVersions over [104,106].
	st2 := StoreConstructor(db, types.StoreOptions{}).(*Store)
	if err := st2.LoadLatestVersion(); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := st2.Get(nil, []byte("k")); !bytes.Equal(got, []byte("v106")) {
		t.Fatalf("reloaded Get = %q, want v106", got)
	}
	st2.Set(nil, []byte("k"), []byte("v107"))
	if id = st2.Commit(); id.Version != 107 {
		t.Fatalf("post-reload commit at v%d, want 107", id.Version)
	}
}
