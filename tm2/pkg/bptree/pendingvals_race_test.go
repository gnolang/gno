package bptree

import (
	"errors"
	"sync"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestPendingVals_ConcurrentValueResolve_NoRace guards the committed-snapshot
// half of the pendingVals split (75c946820): committed-snapshot value reads
// resolve DB-only (getCommittedValue) and never touch ndb.pendingVals, so they
// cannot race the single writer's SaveValue map writes.
//
// The reader hits the committed-snapshot value-resolution paths:
//   - GetImmutable(v).Get(key)          → resolveValue → getCommittedValue
//   - GetMembershipProof(key)           → valueResolver → getCommittedValue
//   - snapshot Iterator.Value()         → getCommittedValue
//
// Regressing any of those resolvers back to GetValue (the map-touching,
// writer-only path) must make this fail under -race.
func TestPendingVals_ConcurrentValueResolve_NoRace(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
	const n = 2_000
	for i := 0; i < n; i++ {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	_, version, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	// Committed snapshot taken before concurrency starts.
	imm, err := tree.GetImmutable(version)
	if err != nil {
		t.Fatal(err)
	}

	var writerWg sync.WaitGroup
	var readerWg sync.WaitGroup
	stop := make(chan struct{})

	// Writer: continuously Set new keys (each Set → SaveValue writes pendingVals)
	// WITHOUT committing, so pendingVals is being mutated for the whole window.
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		k := n
		for {
			select {
			case <-stop:
				return
			default:
			}
			if _, err := tree.Set(i2b(k), i2b(k)); err != nil {
				t.Error(err)
				return
			}
			k++
		}
	}()

	// Reader 1: committed-snapshot Get (value resolution).
	readerWg.Add(1)
	go func() {
		defer readerWg.Done()
		for round := 0; round < 200; round++ {
			for i := 0; i < n; i++ {
				if _, err := imm.Get(i2b(i)); err != nil {
					t.Error(err)
					return
				}
			}
		}
	}()

	// Reader 2: committed membership proof (value resolution via valueResolver).
	readerWg.Add(1)
	go func() {
		defer readerWg.Done()
		for round := 0; round < 400; round++ {
			for i := 0; i < n; i += 50 {
				if _, err := tree.GetMembershipProof(i2b(i)); err != nil {
					t.Error(err)
					return
				}
			}
		}
	}()

	// Reader 3: committed-snapshot iterator Value().
	readerWg.Add(1)
	go func() {
		defer readerWg.Done()
		for round := 0; round < 400; round++ {
			itr, err := imm.Iterator(nil, nil, true)
			if err != nil {
				t.Error(err)
				return
			}
			for itr.Valid() {
				_ = itr.Key()
				_ = itr.Value()
				itr.Next()
			}
			itr.Close()
		}
	}()

	// Reader 4: store-style snapshot iterator via NewIteratorWithNDB (the path
	// the store wrapper uses for an immutable store).
	readerWg.Add(1)
	go func() {
		defer readerWg.Done()
		for round := 0; round < 400; round++ {
			itr := NewIteratorWithNDB(imm, nil, nil, true, tree)
			for itr.Valid() {
				_ = itr.Key()
				_ = itr.Value()
				itr.Next()
			}
			itr.Close()
		}
	}()

	readerWg.Wait()
	close(stop)
	writerWg.Wait()
}

// TestPendingVals_ReadYourWrites verifies the invariant the fix must NOT break:
// on the single writer goroutine, a Get / Iterate AFTER a Set (same session,
// before SaveVersion) returns the staged value.
func TestPendingVals_ReadYourWrites(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())

	// --- Round 1: pure working session (no prior committed version). ---
	staged := map[string][]byte{}
	for i := 0; i < 500; i++ {
		k := i2b(i)
		v := i2b(i * 7)
		if _, err := tree.Set(k, v); err != nil {
			t.Fatal(err)
		}
		staged[string(k)] = v
	}

	// Get-after-Set must see staged values BEFORE SaveVersion.
	for i := 0; i < 500; i++ {
		k := i2b(i)
		got, err := tree.Get(k)
		if err != nil {
			t.Fatalf("Get(%d): %v", i, err)
		}
		if string(got) != string(staged[string(k)]) {
			t.Fatalf("read-your-writes Get(%d): got %x want %x", i, got, staged[string(k)])
		}
	}

	// Iterate-after-Set must see staged values BEFORE SaveVersion.
	seen := 0
	_, err := tree.Iterate(func(key, value []byte) bool {
		want := staged[string(key)]
		if string(value) != string(want) {
			t.Errorf("read-your-writes Iterate key=%x: got %x want %x", key, value, want)
		}
		seen++
		return false
	})
	if err != nil {
		t.Fatal(err)
	}
	if seen != 500 {
		t.Fatalf("Iterate saw %d keys, want 500", seen)
	}

	// MutableTree.Iterator (working-tree iterator) must also see staged values.
	itr, err := tree.Iterator(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	cnt := 0
	for itr.Valid() {
		k := itr.Key()
		v := itr.Value()
		if string(v) != string(staged[string(k)]) {
			t.Errorf("working-tree Iterator key=%x: got %x want %x", k, v, staged[string(k)])
		}
		cnt++
		itr.Next()
	}
	itr.Close()
	if cnt != 500 {
		t.Fatalf("working-tree Iterator saw %d, want 500", cnt)
	}

	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}

	// --- Round 2: UPDATE staged values over a committed base, then read. ---
	// This is the critical case: a Set that overwrites a committed key must be
	// visible to a same-session Get (the new value lives only in pendingVals).
	for i := 0; i < 500; i++ {
		k := i2b(i)
		v := i2b(i*7 + 1) // new value
		if _, err := tree.Set(k, v); err != nil {
			t.Fatal(err)
		}
		staged[string(k)] = v
	}
	for i := 0; i < 500; i++ {
		k := i2b(i)
		got, err := tree.Get(k)
		if err != nil {
			t.Fatalf("Get(%d): %v", i, err)
		}
		if string(got) != string(staged[string(k)]) {
			t.Fatalf("read-your-writes (update) Get(%d): got %x want %x", i, got, staged[string(k)])
		}
	}
	itr2, err := tree.Iterator(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	for itr2.Valid() {
		k := itr2.Key()
		v := itr2.Value()
		if string(v) != string(staged[string(k)]) {
			t.Errorf("working-tree Iterator (update) key=%x: got %x want %x", k, v, staged[string(k)])
		}
		itr2.Next()
	}
	itr2.Close()
}

// TestExport_ConcurrentWithWriter_NoRace guards the Export half of the
// pendingVals split: Export is concurrent BY CONSTRUCTION (it spawns its own
// streaming goroutine), so its value resolution must be DB-only
// (getCommittedValue in export.go) even when a writer is staging values
// (SaveValue → pendingVals map writes) at the same time. Regressing the
// exporter's resolution to ndb.GetValue must make this fail under -race —
// none of the other guard tests covers the exporter goroutine.
func TestExport_ConcurrentWithWriter_NoRace(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
	const n = 2_000
	for i := 0; i < n; i++ {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	_, version, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	imm, err := tree.GetImmutable(version)
	if err != nil {
		t.Fatal(err)
	}
	defer imm.Close()

	// Writer: continuously stage values WITHOUT committing, so pendingVals is
	// being mutated for the whole window.
	stop := make(chan struct{})
	var writerWg sync.WaitGroup
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		k := n
		for {
			select {
			case <-stop:
				return
			default:
			}
			if _, err := tree.Set(i2b(k), i2b(k)); err != nil {
				t.Error(err)
				return
			}
			k++
		}
	}()

	// Drain full exports of the committed snapshot while the writer runs. Each
	// Export spawns the exporter goroutine, which resolves every leaf value.
	for round := 0; round < 50; round++ {
		exp, err := imm.Export(imm.ndb)
		if err != nil {
			t.Fatal(err)
		}
		leaves := 0
		for {
			node, nerr := exp.Next()
			if errors.Is(nerr, ErrExportDone) {
				break
			}
			if nerr != nil {
				exp.Close()
				t.Fatalf("export round %d: %v", round, nerr)
			}
			if node.Height == 0 {
				leaves++
			}
		}
		exp.Close()
		if leaves != n {
			t.Fatalf("export round %d: drained %d leaves, want %d", round, leaves, n)
		}
	}

	close(stop)
	writerWg.Wait()
}
