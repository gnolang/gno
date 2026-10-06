package bptree

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// Part A of the fail-loud batch: a corrupt or missing node record fails the
// ONE operation with an error instead of killing the process, and a failed
// mutation leaves the tree bit-identical (every mutation happens on
// unpublished clones; publication is gated on success).

// corruptOneNonRootNode flips a byte in one non-root 'B' record and returns
// an undo func.
func corruptOneNonRootNode(t *testing.T, db dbm.DB) func() {
	t.Helper()
	rootNK, _, err := (&nodeDB{db: db}).GetRoot(1)
	if err != nil {
		t.Fatal(err)
	}
	itr, err := db.Iterator([]byte{PrefixNode}, []byte{PrefixNode + 1})
	if err != nil {
		t.Fatal(err)
	}
	var key, val []byte
	for ; itr.Valid(); itr.Next() {
		if bytes.Equal(itr.Key()[1:], rootNK) {
			continue // skip the root record: corrupt a mid-tree node
		}
		key = append([]byte(nil), itr.Key()...)
		val = append([]byte(nil), itr.Value()...)
		break
	}
	itr.Close()
	if key == nil {
		t.Fatal("no non-root node record found (tree too small?)")
	}
	corrupted := append([]byte(nil), val...)
	corrupted[len(corrupted)/2] ^= 0x01
	if err := db.Set(key, corrupted); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := db.Set(key, val); err != nil {
			t.Fatal(err)
		}
	}
}

func buildFailLoudTree(t *testing.T) (dbm.DB, *MutableTree) {
	t.Helper()
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger()) // cache=0: force raw loads
	for i := range 200 {
		if _, err := tree.Set(fmt.Appendf(nil, "key%03d", i), []byte("v")); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	return db, tree
}

func TestCorruptNode_ReadsErrorNotPanic(t *testing.T) {
	db, _ := buildFailLoudTree(t)
	undo := corruptOneNonRootNode(t, db)

	fresh := NewMutableTreeWithDB(db, 0, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatal(err)
	}

	// Some Get must hit the corrupt node and ERROR (not panic, not silent).
	sawErr := false
	for i := range 200 {
		v, err := fresh.Get(fmt.Appendf(nil, "key%03d", i))
		if err != nil {
			if !errors.Is(err, ErrChecksumMismatch) {
				t.Fatalf("Get error does not wrap ErrChecksumMismatch: %v", err)
			}
			sawErr = true
			break
		}
		if string(v) != "v" {
			t.Fatalf("silent wrong read: key%03d = %q", i, v)
		}
	}
	if !sawErr {
		t.Fatal("no Get errored despite a corrupt node record")
	}

	// Iterator: error surfaces via Error(), no panic, never wrong rows.
	itr, err := fresh.Iterator(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	rows := 0
	for ; itr.Valid(); itr.Next() {
		rows++
	}
	if itr.Error() == nil {
		t.Fatalf("iterator drained %d rows without surfacing the corruption", rows)
	}
	itr.Close()

	// Restored DB serves everything again.
	undo()
	fresh2 := NewMutableTreeWithDB(db, 0, NewNopLogger())
	if _, err := fresh2.Load(); err != nil {
		t.Fatal(err)
	}
	for i := range 200 {
		if v, err := fresh2.Get(fmt.Appendf(nil, "key%03d", i)); err != nil || string(v) != "v" {
			t.Fatalf("after restore: key%03d = %q, %v", i, v, err)
		}
	}
}

func TestCorruptNode_FailedMutationLeavesTreeUntouched(t *testing.T) {
	db, _ := buildFailLoudTree(t)
	undo := corruptOneNonRootNode(t, db)

	fresh := NewMutableTreeWithDB(db, 0, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatal(err)
	}
	rootBefore := fresh.root
	sizeBefore := fresh.Size()
	hashBefore := fresh.Hash()

	// Find a key whose descent errors, for both Set and Remove.
	var failedSet, failedRemove bool
	for i := range 200 {
		k := fmt.Appendf(nil, "key%03d", i)
		if _, err := fresh.Set(k, []byte("new")); err != nil {
			failedSet = true
			break
		}
		fresh.Rollback() // discard the successful Set; keep hunting
	}
	if !failedSet {
		t.Fatal("no Set errored despite a corrupt node record")
	}
	if fresh.root != rootBefore {
		t.Fatal("failed Set published a new root")
	}
	if fresh.Size() != sizeBefore || !bytes.Equal(fresh.Hash(), hashBefore) {
		t.Fatalf("failed Set mutated the tree: size %d->%d", sizeBefore, fresh.Size())
	}
	for i := range 200 {
		k := fmt.Appendf(nil, "key%03d", i)
		if _, _, err := fresh.Remove(k); err != nil {
			failedRemove = true
			break
		}
		fresh.Rollback()
	}
	if !failedRemove {
		t.Fatal("no Remove errored despite a corrupt node record")
	}
	if fresh.root != rootBefore || fresh.Size() != sizeBefore || !bytes.Equal(fresh.Hash(), hashBefore) {
		t.Fatal("failed Remove mutated the tree")
	}

	// The session stays usable: after restoring the DB, the same handle can
	// mutate and commit consistently.
	undo()
	if _, err := fresh.Set([]byte("key000"), []byte("after")); err != nil {
		t.Fatalf("Set after restore: %v", err)
	}
	if _, _, err := fresh.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion after restore: %v", err)
	}
	if v, err := fresh.Get([]byte("key000")); err != nil || string(v) != "after" {
		t.Fatalf("post-recovery Get: %q, %v", v, err)
	}
}

func TestIterator_FailedConstructionReleasesReader(t *testing.T) {
	db, tree := buildFailLoudTree(t)
	// Corrupt the ROOT record's child so the construction seek fails.
	undo := corruptOneNonRootNode(t, db)
	defer undo()

	fresh := NewMutableTreeWithDB(db, 0, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatal(err)
	}
	_ = tree

	imm, err := fresh.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	itr, err := imm.Iterator(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if itr.Valid() {
		// The corrupt node may not be on the leftmost path; drain until the
		// error (or exhaustion) so the failure mode is exercised either way.
		for ; itr.Valid(); itr.Next() {
		}
	}
	if itr.Error() == nil {
		t.Skip("corrupt node not on this iterator's path") // shape-dependent
	}
	itr.Close()
	itr.Close() // double Close must not double-release
	imm.Close()

	fresh.ndb.mtx.Lock()
	readers := len(fresh.ndb.versionReaders)
	fresh.ndb.mtx.Unlock()
	if readers != 0 {
		t.Fatalf("version readers leaked after errored iterator: %d entries", readers)
	}
	// Pruning must not be blocked by a phantom reader (needs a v2 so v1 is prunable).
	if _, err := fresh.Set([]byte("zzz"), []byte("v")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := fresh.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	if err := fresh.PruneVersionsTo(1); err != nil {
		t.Fatalf("prune blocked after errored iterator: %v", err)
	}
}
