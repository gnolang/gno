package bptree

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// The malformed-stream matrix: every structurally invalid stream must be
// rejected with a named error, the importer must be unusable afterward, and
// Close must leave the tree session clean (version, initialVersion, staged
// values all restored). The honest invariants being enforced (derived from
// split.go/remove.go): leaf keys strictly ascend across the stream; every
// leaf marker drains the buffer exactly; separators sit in
// max(left) < sep <= min(right); inner heights equal child height + 1.

func importLeaf(t *testing.T, imp *Importer, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if err := imp.Add(&ExportNode{Height: 0, Key: []byte(k), Value: []byte("v-" + k)}); err != nil {
			t.Fatalf("Add leaf entry %q: %v", k, err)
		}
	}
	if err := imp.Add(&ExportNode{Height: -1, NumKeys: int16(len(keys))}); err != nil {
		t.Fatalf("Add leaf marker (%d keys): %v", len(keys), err)
	}
}

func TestImportValidation_RejectsMalformedStreams(t *testing.T) {
	cases := []struct {
		name    string
		errPart string
		feed    func(t *testing.T, imp *Importer) error
	}{
		{
			// The migration guard for the M24 empty-key divergence: bptree
			// rejects empty keys at Set, so the importer must reject them too —
			// an IAVL→bptree migration carrying an empty-key entry fails loud
			// at import time, never as a runtime panic after cutover.
			"empty leaf key", "empty", func(t *testing.T, imp *Importer) error {
				t.Helper()
				return imp.Add(&ExportNode{Height: 0, Key: []byte{}, Value: []byte("v")})
			},
		},
		{
			"unsorted leaf keys", "sorted", func(t *testing.T, imp *Importer) error {
				t.Helper()
				if err := imp.Add(&ExportNode{Height: 0, Key: []byte("b"), Value: []byte("v")}); err != nil {
					return err
				}
				return imp.Add(&ExportNode{Height: 0, Key: []byte("a"), Value: []byte("v")})
			},
		},
		{
			"duplicate leaf keys", "sorted", func(t *testing.T, imp *Importer) error {
				t.Helper()
				if err := imp.Add(&ExportNode{Height: 0, Key: []byte("a"), Value: []byte("v1")}); err != nil {
					return err
				}
				return imp.Add(&ExportNode{Height: 0, Key: []byte("a"), Value: []byte("v2")})
			},
		},
		{
			"unsorted across leaf boundary", "sorted", func(t *testing.T, imp *Importer) error {
				t.Helper()
				importLeaf(t, imp, "m")
				return imp.Add(&ExportNode{Height: 0, Key: []byte("a"), Value: []byte("v")})
			},
		},
		{
			"cross-boundary regrouping", "exactly", func(t *testing.T, imp *Importer) error {
				t.Helper()
				// e1,e2, marker(1), would leave e1 buffered — the exporter
				// always drains exactly.
				if err := imp.Add(&ExportNode{Height: 0, Key: []byte("a"), Value: []byte("v")}); err != nil {
					return err
				}
				if err := imp.Add(&ExportNode{Height: 0, Key: []byte("b"), Value: []byte("v")}); err != nil {
					return err
				}
				return imp.Add(&ExportNode{Height: -1, NumKeys: 1})
			},
		},
		{
			"separator below left max", "window", func(t *testing.T, imp *Importer) error {
				t.Helper()
				importLeaf(t, imp, "a", "b")
				importLeaf(t, imp, "m", "n")
				// sep must satisfy max(left)="b" < sep <= min(right)="m";
				// "b" violates the strict left bound.
				return imp.Add(&ExportNode{Height: 1, NumKeys: 1, SeparatorKeys: [][]byte{[]byte("b")}})
			},
		},
		{
			"separator above right min", "window", func(t *testing.T, imp *Importer) error {
				t.Helper()
				importLeaf(t, imp, "a", "b")
				importLeaf(t, imp, "m", "n")
				// "z" > min(right)="m".
				return imp.Add(&ExportNode{Height: 1, NumKeys: 1, SeparatorKeys: [][]byte{[]byte("z")}})
			},
		},
		{
			"wrong inner height", "height", func(t *testing.T, imp *Importer) error {
				t.Helper()
				importLeaf(t, imp, "a")
				importLeaf(t, imp, "m")
				// Children are leaves (height 0) → derived inner height 1.
				return imp.Add(&ExportNode{Height: 5, NumKeys: 1, SeparatorKeys: [][]byte{[]byte("m")}})
			},
		},
		{
			"non-uniform child heights", "height", func(t *testing.T, imp *Importer) error {
				t.Helper()
				// Build an inner over two leaves (height 1), then a third
				// leaf, then an inner claiming both as children: heights 1
				// and 0 are non-uniform.
				importLeaf(t, imp, "a")
				importLeaf(t, imp, "f")
				if err := imp.Add(&ExportNode{Height: 1, NumKeys: 1, SeparatorKeys: [][]byte{[]byte("f")}}); err != nil {
					return err
				}
				importLeaf(t, imp, "m")
				return imp.Add(&ExportNode{Height: 2, NumKeys: 1, SeparatorKeys: [][]byte{[]byte("m")}})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree := newMemTree()
			imp, err := tree.Import(1)
			if err != nil {
				t.Fatal(err)
			}
			defer imp.Close()
			err = tc.feed(t, imp)
			if err == nil {
				t.Fatal("malformed stream accepted")
			}
			if !strings.Contains(err.Error(), tc.errPart) {
				t.Fatalf("error %q does not name %q", err, tc.errPart)
			}
		})
	}
}

// Equal-boundary separators are HONEST (splits set sep = right.keys[0]) and
// so are stale-low separators (a deletion raised the right subtree's min);
// both must import.
func TestImportValidation_AcceptsHonestSeparators(t *testing.T) {
	tree := newMemTree()
	imp, err := tree.Import(1)
	if err != nil {
		t.Fatal(err)
	}
	importLeaf(t, imp, "a", "b")
	importLeaf(t, imp, "m", "n")
	importLeaf(t, imp, "x", "y")
	// sep[0]="m" == min(child1) (equality); sep[1]="q" < min(child2)="x"
	// (stale-low, as after deleting old keys q..w).
	if err := imp.Add(&ExportNode{
		Height: 1, NumKeys: 2,
		SeparatorKeys: [][]byte{[]byte("m"), []byte("q")},
	}); err != nil {
		t.Fatalf("honest separators rejected: %v", err)
	}
	if err := imp.Commit(); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"a", "b", "m", "n", "x", "y"} {
		if got, err := tree.Get([]byte(k)); err != nil || string(got) != "v-"+k {
			t.Fatalf("Get(%s) = %q, %v", k, got, err)
		}
	}
}

// Export→import round-trip of a tree whose separators went stale-low through
// deletions — guards the window check's `<=` right bound against
// over-validation.
func TestImportValidation_RoundTripAfterDeletions(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
	for i := range 200 {
		if _, err := tree.Set(fmt.Appendf(nil, "key%03d", i), []byte("v")); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	// Delete many subtree minima to leave stale-low separators behind.
	for i := 0; i < 200; i += 3 {
		if _, _, err := tree.Remove(fmt.Appendf(nil, "key%03d", i)); err != nil {
			t.Fatal(err)
		}
	}
	hash, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	imm, err := tree.GetImmutable(2)
	if err != nil {
		t.Fatal(err)
	}
	defer imm.Close()
	exp, err := imm.Export(tree.ndb)
	if err != nil {
		t.Fatal(err)
	}
	defer exp.Close()

	tree2 := newMemTree()
	imp, err := tree2.Import(1)
	if err != nil {
		t.Fatal(err)
	}
	defer imp.Close()
	for {
		node, err := exp.Next()
		if err != nil {
			break
		}
		if err := imp.Add(node); err != nil {
			t.Fatalf("honest post-deletion stream rejected: %v", err)
		}
	}
	if err := imp.Commit(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tree2.Hash(), hash) {
		t.Fatalf("round-trip hash mismatch: %x != %x", tree2.Hash(), hash)
	}
}

// The importer lifecycle state machine: Add/Commit error after Close, after
// Commit, and after a FAILED Commit (whose retry would otherwise commit a
// tree with every value record missing under a root hash that matches the
// trusted app hash); Close is idempotent and restores the session.
func TestImportValidation_StateMachine(t *testing.T) {
	t.Run("add and commit after close error", func(t *testing.T) {
		tree := newMemTree()
		imp, err := tree.Import(1)
		if err != nil {
			t.Fatal(err)
		}
		importLeaf(t, imp, "a")
		if err := imp.Close(); err != nil {
			t.Fatal(err)
		}
		if err := imp.Add(&ExportNode{Height: 0, Key: []byte("b"), Value: []byte("v")}); err == nil {
			t.Fatal("Add after Close accepted")
		}
		if err := imp.Commit(); err == nil {
			t.Fatal("Commit after Close accepted")
		}
		if err := imp.Close(); err != nil {
			t.Fatalf("double Close: %v", err)
		}
	})

	t.Run("add after commit errors", func(t *testing.T) {
		tree := newMemTree()
		imp, err := tree.Import(1)
		if err != nil {
			t.Fatal(err)
		}
		importLeaf(t, imp, "a")
		if err := imp.Commit(); err != nil {
			t.Fatal(err)
		}
		if err := imp.Add(&ExportNode{Height: 0, Key: []byte("b"), Value: []byte("v")}); err == nil {
			t.Fatal("Add after Commit accepted")
		}
		if err := imp.Commit(); err == nil {
			t.Fatal("Commit after Commit accepted")
		}
		if err := imp.Close(); err != nil {
			t.Fatalf("Close after Commit: %v", err)
		}
		// The committed version survives Close.
		if got, err := tree.Get([]byte("a")); err != nil || string(got) != "v-a" {
			t.Fatalf("committed import lost: %q, %v", got, err)
		}
	})

	t.Run("commit after failed commit errors", func(t *testing.T) {
		tree := newMemTree()
		imp, err := tree.Import(5)
		if err != nil {
			t.Fatal(err)
		}
		// Two roots on the stack → Commit shape-check fails → importer poisoned.
		importLeaf(t, imp, "a")
		importLeaf(t, imp, "m")
		if err := imp.Commit(); err == nil {
			t.Fatal("two-root commit accepted")
		}
		if err := imp.Commit(); err == nil || !strings.Contains(err.Error(), "failed") {
			t.Fatalf("retry after failed Commit: want failed-importer error, got %v", err)
		}
		if err := imp.Add(&ExportNode{Height: 0, Key: []byte("z"), Value: []byte("v")}); err == nil {
			t.Fatal("Add after failed Commit accepted")
		}
	})

	t.Run("close restores session after failed commit", func(t *testing.T) {
		db := memdb.NewMemDB()
		tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
		if _, err := tree.Set([]byte("pre"), []byte("existing")); err != nil {
			t.Fatal(err)
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatal(err)
		}
		tree.SetInitialVersion(7) // must survive a failed import Commit

		imp, err := tree.Import(10)
		if err != nil {
			t.Fatal(err)
		}
		importLeaf(t, imp, "a")
		importLeaf(t, imp, "m") // two roots → Commit fails after staging values
		if err := imp.Commit(); err == nil {
			t.Fatal("two-root commit accepted")
		}
		if tree.version != 1 || tree.initialVersion != 7 {
			t.Fatalf("failed Commit left version=%d initialVersion=%d", tree.version, tree.initialVersion)
		}
		if err := imp.Close(); err != nil {
			t.Fatal(err)
		}
		// Staged import values must not ride into the next commit.
		if _, err := tree.Set([]byte("post"), []byte("v")); err != nil {
			t.Fatal(err)
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatal(err)
		}
		// No 'V' record in the import-target namespace (version 10).
		raw, err := db.Get(valueDBKey((&NodeKey{Version: 10, Nonce: 0}).GetKey()))
		if err != nil {
			t.Fatal(err)
		}
		if raw != nil {
			t.Fatalf("aborted import leaked a staged value into the next commit: %x", raw)
		}
		// And the committed state is sane.
		fresh := NewMutableTreeWithDB(db, 100, NewNopLogger())
		if _, err := fresh.Load(); err != nil {
			t.Fatal(err)
		}
		for k, want := range map[string]string{"pre": "existing", "post": "v"} {
			if got, err := fresh.Get([]byte(k)); err != nil || string(got) != want {
				t.Fatalf("Get(%s) = %q, %v", k, got, err)
			}
		}
	})

	t.Run("abandoned import leaks nothing", func(t *testing.T) {
		db := memdb.NewMemDB()
		tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
		imp, err := tree.Import(10)
		if err != nil {
			t.Fatal(err)
		}
		importLeaf(t, imp, "a")
		if err := imp.Close(); err != nil { // abandon
			t.Fatal(err)
		}
		if _, err := tree.Set([]byte("k"), []byte("v")); err != nil {
			t.Fatal(err)
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatal(err)
		}
		raw, err := db.Get(valueDBKey((&NodeKey{Version: 10, Nonce: 0}).GetKey()))
		if err != nil {
			t.Fatal(err)
		}
		if raw != nil {
			t.Fatalf("abandoned import leaked a staged value: %x", raw)
		}
	})
}

// M18: the import target must exceed the latest version — importing into the
// live namespace would overwrite records shared with retained versions.
func TestImport_RejectsNonFutureVersion(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
	for v := 1; v <= 3; v++ {
		if _, err := tree.Set(fmt.Appendf(nil, "k%d", v), []byte("v")); err != nil {
			t.Fatal(err)
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatal(err)
		}
	}
	// Prune v1 so it no longer "exists" — pre-M18 this was the corruption
	// hole: a pruned version passes the VersionExists check while its key
	// namespace is still shared into retained versions.
	if err := tree.PruneVersionsTo(1); err != nil {
		t.Fatal(err)
	}
	for _, v := range []int64{1, 2, 3, 0, -5} {
		if _, err := tree.Import(v); err == nil {
			t.Fatalf("Import(%d) accepted with latest=3", v)
		}
	}
	imp, err := tree.Import(4)
	if err != nil {
		t.Fatalf("Import(latest+1): %v", err)
	}
	imp.Close()
	// Gaps beyond latest are fine (fresh namespaces).
	imp, err = tree.Import(100)
	if err != nil {
		t.Fatalf("Import(latest+97): %v", err)
	}
	imp.Close()
}
