package bptree

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/boltdb"
	"github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// These tests pin the copy-on-return contract: every buffer the package hands
// out (Get results, iterator keys/values, Iterate callback keys, export
// nodes, GetByIndex keys) is the CALLER's to mutate — doing so must never
// change tree content, committed state, the shared node cache, or the app
// hash. Before the copies, mutating a Get result pre-commit committed
// DIFFERENT state on memdb vs goleveldb under one app hash, and mutating an
// Iterator.Key() corrupted the live tree persistently.

func buildAliasTree(t *testing.T, db dbm.DB) *MutableTree {
	t.Helper()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
	for i := range 100 {
		if _, err := tree.Set(fmt.Appendf(nil, "key%03d", i), fmt.Appendf(nil, "val%03d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	return tree
}

func assertAliasTreeIntact(t *testing.T, db dbm.DB, wantHash []byte) {
	t.Helper()
	fresh := NewMutableTreeWithDB(db, 100, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fresh.Hash(), wantHash) {
		t.Fatalf("committed hash changed: %x != %x", fresh.Hash(), wantHash)
	}
	for i := range 100 {
		got, err := fresh.Get(fmt.Appendf(nil, "key%03d", i))
		if err != nil || string(got) != fmt.Sprintf("val%03d", i) {
			t.Fatalf("key%03d: %q, %v", i, got, err)
		}
	}
}

func TestAlias_GetMutationDoesNotCorrupt(t *testing.T) {
	db := memdb.NewMemDB()
	tree := buildAliasTree(t, db)
	hash := tree.Hash()

	// Committed read.
	g, err := tree.Get([]byte("key007"))
	if err != nil {
		t.Fatal(err)
	}
	for i := range g {
		g[i] = 'X'
	}
	// Staged read (pendingVals path).
	if _, err := tree.Set([]byte("key007"), []byte("staged")); err != nil {
		t.Fatal(err)
	}
	s, err := tree.Get([]byte("key007"))
	if err != nil {
		t.Fatal(err)
	}
	for i := range s {
		s[i] = 'Y'
	}
	if again, _ := tree.Get([]byte("key007")); string(again) != "staged" {
		t.Fatalf("staged value corrupted by read mutation: %q", again)
	}
	tree.Rollback()
	assertAliasTreeIntact(t, db, hash)
}

func TestAlias_IteratorKeyValueMutation(t *testing.T) {
	db := memdb.NewMemDB()
	tree := buildAliasTree(t, db)
	hash := tree.Hash()

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	defer imm.Close()
	itr, err := imm.Iterator(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	rows := 0
	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(), itr.Value()
		for i := range k {
			k[i] = 'X'
		}
		for i := range v {
			v[i] = 'X'
		}
		rows++
	}
	itr.Close()
	if itr.Error() != nil {
		t.Fatal(itr.Error())
	}
	if rows != 100 {
		t.Fatalf("iterated %d rows, want 100", rows)
	}
	assertAliasTreeIntact(t, db, hash)
}

func TestAlias_IterateCallbackKeyMutation(t *testing.T) {
	db := memdb.NewMemDB()
	tree := buildAliasTree(t, db)
	hash := tree.Hash()

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	defer imm.Close()
	if _, err := imm.Iterate(func(key, value []byte) bool {
		for i := range key {
			key[i] = 'X'
		}
		return false
	}); err != nil {
		t.Fatal(err)
	}
	// MutableTree.Iterate too.
	if _, err := tree.Iterate(func(key, value []byte) bool {
		for i := range key {
			key[i] = 'X'
		}
		return false
	}); err != nil {
		t.Fatal(err)
	}
	// GetByIndex returned key.
	key, _, err := imm.GetByIndex(3)
	if err != nil {
		t.Fatal(err)
	}
	for i := range key {
		key[i] = 'X'
	}
	assertAliasTreeIntact(t, db, hash)
}

func TestAlias_ExportNodeMutation(t *testing.T) {
	db := memdb.NewMemDB()
	tree := buildAliasTree(t, db)
	hash := tree.Hash()

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	defer imm.Close()
	exp, err := imm.Export(tree.ndb)
	if err != nil {
		t.Fatal(err)
	}
	for {
		node, err := exp.Next()
		if err != nil {
			break
		}
		for i := range node.Key {
			node.Key[i] = 'X'
		}
		for _, sep := range node.SeparatorKeys {
			for i := range sep {
				sep[i] = 'X'
			}
		}
	}
	exp.Close()
	assertAliasTreeIntact(t, db, hash)
}

func TestAlias_IteratorBoundsMutation(t *testing.T) {
	db := memdb.NewMemDB()
	tree := buildAliasTree(t, db)

	start := []byte("key010")
	end := []byte("key020")
	itr, err := tree.Iterator(start, end, true)
	if err != nil {
		t.Fatal(err)
	}
	// Hostile caller shifts its own slices mid-iteration.
	copy(start, "key000")
	copy(end, "zzzzzz")
	rows := 0
	for ; itr.Valid(); itr.Next() {
		rows++
	}
	itr.Close()
	if rows != 10 {
		t.Fatalf("bounds shifted with caller mutation: %d rows, want 10", rows)
	}

	// Nil bounds stay nil (unbounded) — full-range iteration must see all keys.
	full, err := tree.Iterator(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	rows = 0
	for ; full.Valid(); full.Next() {
		rows++
	}
	full.Close()
	if rows != 100 {
		t.Fatalf("nil-bound iterator saw %d rows, want 100", rows)
	}
}

// TestIterate_NilResolverErrors (site 9): a resolver-less ImmutableTree must
// refuse Iterate instead of silently passing value HASHES (slices aliasing
// the leaf's live hash arrays) where the callback expects values.
func TestIterate_NilResolverErrors(t *testing.T) {
	tree := newMemTree()
	if _, err := tree.Set([]byte("a"), []byte("v")); err != nil {
		t.Fatal(err)
	}
	bare := NewImmutableTree(tree.root, 1)
	if _, err := bare.Iterate(func(key, value []byte) bool { return false }); err == nil {
		t.Fatal("Iterate on a resolver-less tree should error")
	} else if !strings.Contains(err.Error(), "resolver") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestN47_CrossBackendCommitIdentical: the live-divergence repro. The same op
// stream — including a hostile mutation of a pre-commit Get result and of
// iterator Key/Value buffers — must commit byte-identical state and an
// identical app hash on every backend (memdb retains staged slices by
// reference; goleveldb copies at batch.Set; boltdb retains like memdb).
func TestN47_CrossBackendCommitIdentical(t *testing.T) {
	type backend struct {
		name string
		db   dbm.DB
	}
	dir := t.TempDir()
	gldb, err := goleveldb.NewGoLevelDB("n47gl", dir)
	if err != nil {
		t.Fatal(err)
	}
	defer gldb.Close()
	bdb, err := boltdb.New("n47bolt", dir)
	if err != nil {
		t.Fatal(err)
	}
	defer bdb.Close()
	backends := []backend{
		{"memdb", memdb.NewMemDB()},
		{"goleveldb", gldb},
		{"boltdb", bdb},
	}

	hashes := make(map[string][]byte)
	for _, be := range backends {
		tree := NewMutableTreeWithDB(be.db, 100, NewNopLogger())
		if _, err := tree.Set([]byte("k"), []byte("AAAA")); err != nil {
			t.Fatal(err)
		}
		g, err := tree.Get([]byte("k"))
		if err != nil {
			t.Fatal(err)
		}
		g[0] = 'Z' // hostile pre-commit mutation
		itr, err := tree.Iterator(nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		for ; itr.Valid(); itr.Next() {
			k, v := itr.Key(), itr.Value()
			if len(k) > 0 {
				k[0] = 'Z'
			}
			if len(v) > 0 {
				v[0] = 'Z'
			}
		}
		itr.Close()
		hash, _, err := tree.SaveVersion()
		if err != nil {
			t.Fatal(err)
		}
		hashes[be.name] = hash

		fresh := NewMutableTreeWithDB(be.db, 100, NewNopLogger())
		if _, err := fresh.Load(); err != nil {
			t.Fatal(err)
		}
		got, err := fresh.Get([]byte("k"))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "AAAA" {
			t.Fatalf("%s: committed value corrupted by hostile mutation: %q", be.name, got)
		}
	}
	for _, be := range backends[1:] {
		if !bytes.Equal(hashes[be.name], hashes["memdb"]) {
			t.Fatalf("app hash diverges across backends: memdb=%x %s=%x",
				hashes["memdb"], be.name, hashes[be.name])
		}
	}
}
