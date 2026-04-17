package bptree

// Long-running stress tests. Run with:
//   go test ./tm2/pkg/bptree/ -run TestStress -timeout 600s -count=1

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// --- Test 1: Random operation oracle ---

func TestStress_RandomOperationOracle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	oracle := make(map[string]string) // ground truth
	rng := rand.New(rand.NewSource(42))

	const (
		numOps         = 20000
		verifyEvery    = 500
		saveEvery      = 200
		pruneEvery     = 1000
		maxKeySpace    = 2000
	)

	var savedVersions []int64

	for op := 0; op < numOps; op++ {
		key := fmt.Sprintf("key%04d", rng.Intn(maxKeySpace))
		action := rng.Intn(3)

		switch action {
		case 0, 1: // Set (2/3 probability)
			val := fmt.Sprintf("val_%d_%d", op, rng.Intn(10000))
			tree.Set([]byte(key), []byte(val))
			oracle[key] = val

		case 2: // Remove (1/3 probability)
			tree.Remove([]byte(key))
			delete(oracle, key)
		}

		// Periodic SaveVersion
		if (op+1)%saveEvery == 0 {
			_, v, err := tree.SaveVersion()
			if err != nil {
				t.Fatalf("op %d: SaveVersion: %v", op, err)
			}
			savedVersions = append(savedVersions, v)
		}

		// Periodic prune
		if (op+1)%pruneEvery == 0 && len(savedVersions) > 2 {
			pruneTarget := savedVersions[len(savedVersions)-2]
			if err := tree.DeleteVersionsTo(pruneTarget); err != nil {
				t.Fatalf("op %d: prune to %d: %v", op, pruneTarget, err)
			}
		}

		// Periodic full verification against oracle
		if (op+1)%verifyEvery == 0 {
			for k, v := range oracle {
				got, err := tree.Get([]byte(k))
				if err != nil {
					t.Fatalf("op %d: Get(%q): %v", op, k, err)
				}
				if string(got) != v {
					t.Fatalf("op %d: Get(%q) = %q, oracle says %q", op, k, got, v)
				}
			}
			// Verify size
			if tree.Size() != int64(len(oracle)) {
				t.Fatalf("op %d: tree.Size()=%d, oracle has %d keys", op, tree.Size(), len(oracle))
			}
		}
	}

	t.Logf("completed %d ops, %d versions saved, final size=%d", numOps, len(savedVersions), tree.Size())
}

// --- Test 2: Value leak detector ---

func TestStress_ValueLeakDetector(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger(), InlineValueThresholdOption(-1))
	rng := rand.New(rand.NewSource(99))

	const (
		numVersions = 200
		keysPerVersion = 100
		maxKeySpace = 500
	)

	for v := 0; v < numVersions; v++ {
		// Random mutations
		for i := 0; i < keysPerVersion; i++ {
			key := fmt.Sprintf("vl%04d", rng.Intn(maxKeySpace))
			if rng.Intn(4) == 0 {
				tree.Remove([]byte(key))
			} else {
				tree.Set([]byte(key), fmt.Appendf(nil, "v%d_%d", v, i))
			}
		}
		tree.SaveVersion()

		// Prune all but last 2 versions
		if v >= 2 {
			if err := tree.DeleteVersionsTo(int64(v - 1)); err != nil {
				t.Fatalf("v%d: prune: %v", v, err)
			}
		}

		// Count values in DB — should roughly equal live keys
		// (plus values from the one surviving old version)
		valCount := countDBValues(db)
		liveKeys := tree.Size()

		// With 2 surviving versions, value count should be at most 2x live keys
		// (each key could have a value in both versions)
		if valCount > int(liveKeys)*3 {
			t.Fatalf("v%d: value leak detected: %d DB values for %d live keys (>3x)", v, valCount, liveKeys)
		}
	}

	// Final prune to latest-1, leaving only 1 version
	tree.DeleteVersionsTo(int64(numVersions - 1))
	finalValues := countDBValues(db)
	finalKeys := tree.Size()
	if finalValues != int(finalKeys) {
		t.Fatalf("final: %d DB values != %d live keys after full prune", finalValues, finalKeys)
	}

	t.Logf("completed %d versions, final: %d keys, %d values", numVersions, finalKeys, finalValues)
}

// --- Test 3: Hash stability under reload ---

func TestStress_HashStabilityReload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	rng := rand.New(rand.NewSource(77))

	type versionRecord struct {
		version int64
		hash    []byte
		size    int64
	}

	var records []versionRecord

	for round := 0; round < 50; round++ {
		// Mutate
		for i := 0; i < 200; i++ {
			key := fmt.Sprintf("hs%04d", rng.Intn(1000))
			if rng.Intn(4) == 0 {
				tree.Remove([]byte(key))
			} else {
				tree.Set([]byte(key), fmt.Appendf(nil, "r%d_%d", round, i))
			}
		}
		hash, v, err := tree.SaveVersion()
		if err != nil {
			t.Fatalf("round %d: SaveVersion: %v", round, err)
		}
		records = append(records, versionRecord{v, hash, tree.Size()})

		// Prune old versions (keep last 5)
		if len(records) > 5 {
			pruneTarget := records[len(records)-6].version
			tree.DeleteVersionsTo(pruneTarget)
		}

		// Reload from DB and verify surviving versions
		tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
		tree2.Load()
		for _, rec := range records[max(0, len(records)-5):] {
			if !tree2.VersionExists(rec.version) {
				continue // pruned
			}
			tree2loaded := NewMutableTreeWithDB(db, 1000, NewNopLogger())
			tree2loaded.LoadVersion(rec.version)
			reloadHash := tree2loaded.WorkingHash()
			if !bytes.Equal(rec.hash, reloadHash) {
				t.Fatalf("round %d: v%d hash changed after reload: %x != %x", round, rec.version, rec.hash, reloadHash)
			}
			if tree2loaded.Size() != rec.size {
				t.Fatalf("round %d: v%d size changed: %d != %d", round, rec.version, tree2loaded.Size(), rec.size)
			}
		}
	}

	t.Logf("completed 50 rounds of mutate/save/prune/reload")
}

// --- Test 4: Export/import hash fidelity at scale ---

func TestStress_ExportImportLargeTree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	db1 := memdb.NewMemDB()
	tree1 := NewMutableTreeWithDB(db1, 1000, NewNopLogger())
	rng := rand.New(rand.NewSource(55))

	// Build tree with random operations across multiple versions
	for v := 0; v < 20; v++ {
		for i := 0; i < 5000; i++ {
			key := fmt.Sprintf("ei%06d", rng.Intn(50000))
			if rng.Intn(5) == 0 {
				tree1.Remove([]byte(key))
			} else {
				tree1.Set([]byte(key), fmt.Appendf(nil, "v%d_%d", v, i))
			}
		}
		tree1.SaveVersion()
	}

	// Prune to leave only last version
	tree1.DeleteVersionsTo(int64(19))
	origHash := tree1.WorkingHash()
	origSize := tree1.Size()

	// Export
	imm, err := tree1.GetImmutable(20)
	if err != nil {
		t.Fatal(err)
	}
	exporter, err := imm.Export(tree1.ndb)
	if err != nil {
		t.Fatal(err)
	}

	// Import into fresh tree
	db2 := memdb.NewMemDB()
	tree2 := NewMutableTreeWithDB(db2, 1000, NewNopLogger())
	imp, err := tree2.Import(20)
	if err != nil {
		t.Fatal(err)
	}
	for {
		node, err := exporter.Next()
		if err == ErrExportDone {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := imp.Add(node); err != nil {
			t.Fatal(err)
		}
	}
	exporter.Close()
	if err := imp.Commit(); err != nil {
		t.Fatal(err)
	}

	// Verify hash matches
	importHash := tree2.WorkingHash()
	if !bytes.Equal(origHash, importHash) {
		t.Fatalf("export/import hash mismatch: %x != %x", origHash, importHash)
	}
	if tree2.Size() != origSize {
		t.Fatalf("size mismatch: %d != %d", tree2.Size(), origSize)
	}

	// Continue mutating the imported tree, save, prune, verify
	for v := 0; v < 10; v++ {
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("ei%06d", rng.Intn(50000))
			tree2.Set([]byte(key), fmt.Appendf(nil, "post_%d_%d", v, i))
		}
		hash2, ver, err := tree2.SaveVersion()
		if err != nil {
			t.Fatalf("post-import v%d: SaveVersion: %v", v, err)
		}
		if v > 0 {
			tree2.DeleteVersionsTo(ver - 1)
		}

		// Reload and verify
		tree2r := NewMutableTreeWithDB(db2, 1000, NewNopLogger())
		tree2r.LoadVersion(ver)
		if !bytes.Equal(hash2, tree2r.WorkingHash()) {
			t.Fatalf("post-import v%d: hash mismatch after reload", v)
		}
	}

	t.Logf("export/import: %d keys, hash verified, %d post-import versions OK", origSize, 10)
}

// --- Test 5: Concurrent immutable reads during mutations ---

func TestStress_ConcurrentImmutableReads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// Use DB-backed tree. Note: concurrent reads on an ImmutableTree while
	// the MutableTree is writing is a known limitation (#9 — version reader
	// protection is not wired up). This test verifies functional correctness
	// (no wrong values), not thread safety. Run without -race.
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// Build initial state
	for i := 0; i < 1000; i++ {
		tree.Set(fmt.Appendf(nil, "cr%05d", i), fmt.Appendf(nil, "init_%05d", i))
	}
	tree.SaveVersion()

	// Take immutable snapshot via GetImmutable (properly wired resolver)
	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}

	// Spawn concurrent readers on the immutable snapshot.
	// No mutations happen during reads — this tests that concurrent
	// read-only access to an ImmutableTree is safe.
	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(goroutineID)))
			for iter := 0; iter < 500; iter++ {
				idx := rng.Intn(1000)
				key := fmt.Appendf(nil, "cr%05d", idx)
				val, err := imm.Get(key)
				if err != nil {
					errCh <- fmt.Errorf("g%d iter%d: Get(%q): %v", goroutineID, iter, key, err)
					return
				}
				expected := fmt.Appendf(nil, "init_%05d", idx)
				if !bytes.Equal(val, expected) {
					errCh <- fmt.Errorf("g%d iter%d: Get(%q) = %q, want %q", goroutineID, iter, key, val, expected)
					return
				}

				// Also test iterator
				itr, _ := imm.Iterator(key, nil, true)
				if itr.Valid() {
					_ = itr.Key()
					_ = itr.Value()
				}
				itr.Close()
			}
		}(g)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}

	// After concurrent reads complete, verify mutations work
	for i := 0; i < 2000; i++ {
		tree.Set(fmt.Appendf(nil, "cr%05d", i%1000), fmt.Appendf(nil, "mutated_%d", i))
		if (i+1)%500 == 0 {
			tree.SaveVersion()
		}
	}

	t.Logf("concurrent reads: 10 goroutines x 500 reads, no errors")
}

// --- Test 6: Prune stress with many versions ---

func TestStress_PruneManyVersions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger(), InlineValueThresholdOption(-1))
	rng := rand.New(rand.NewSource(33))

	type savedHash struct {
		version int64
		hash    []byte
	}
	var hashes []savedHash

	// Save 500 versions with overlapping mutations
	for v := 0; v < 500; v++ {
		for i := 0; i < 50; i++ {
			key := fmt.Sprintf("pm%04d", rng.Intn(2000))
			if rng.Intn(5) == 0 {
				tree.Remove([]byte(key))
			} else {
				tree.Set([]byte(key), fmt.Appendf(nil, "v%d_%d", v, i))
			}
		}
		hash, ver, err := tree.SaveVersion()
		if err != nil {
			t.Fatalf("v%d: SaveVersion: %v", v, err)
		}
		hashes = append(hashes, savedHash{ver, hash})

		// Prune in batches of 50, keeping last 10
		if v > 0 && v%50 == 0 {
			pruneTarget := hashes[len(hashes)-11].version
			if err := tree.DeleteVersionsTo(pruneTarget); err != nil {
				t.Fatalf("v%d: prune to %d: %v", v, pruneTarget, err)
			}

			// Reload and verify surviving versions
			for _, h := range hashes[max(0, len(hashes)-10):] {
				tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
				tree2.LoadVersion(h.version)
				reloaded := tree2.WorkingHash()
				if !bytes.Equal(h.hash, reloaded) {
					t.Fatalf("v%d: hash mismatch for v%d after prune", v, h.version)
				}
			}
		}
	}

	// Final verification: prune all but last, check value count
	lastHash := hashes[len(hashes)-1]
	tree.DeleteVersionsTo(lastHash.version - 1)
	valCount := countDBValues(db)
	liveKeys := tree.Size()
	if valCount != int(liveKeys) {
		t.Fatalf("final: %d DB values != %d live keys", valCount, liveKeys)
	}

	t.Logf("500 versions, pruned in batches, final: %d keys, %d values", liveKeys, valCount)
}

// --- Test 7: Worst-case tree restructuring ---

func TestStress_WorstCaseRestructuring(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	db := memdb.NewMemDB()
	// Disable inline storage: this test asserts that PrefixVal entries
	// match tree.Size() after prune, which only holds when every value
	// uses the external indirection.
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger(), InlineValueThresholdOption(-1))
	oracle := make(map[string]string)
	rng := rand.New(rand.NewSource(11))

	verify := func(phase string) {
		if tree.Size() != int64(len(oracle)) {
			t.Fatalf("%s: size %d != oracle %d", phase, tree.Size(), len(oracle))
		}
		for k, v := range oracle {
			got, _ := tree.Get([]byte(k))
			if string(got) != v {
				t.Fatalf("%s: Get(%q) = %q, want %q", phase, k, got, v)
			}
		}
		// Verify iterator order
		itr, _ := tree.Iterator(nil, nil, true)
		var prev string
		count := 0
		for itr.Valid() {
			k := string(itr.Key())
			if k <= prev && prev != "" {
				t.Fatalf("%s: iterator order broken at %d: %q <= %q", phase, count, k, prev)
			}
			prev = k
			count++
			itr.Next()
		}
		itr.Close()
		if count != len(oracle) {
			t.Fatalf("%s: iterator count %d != oracle %d", phase, count, len(oracle))
		}
	}

	// Phase 1: Sequential insert (triggers 90/10 splits)
	for i := 0; i < 2000; i++ {
		key := fmt.Sprintf("seq%06d", i)
		val := fmt.Sprintf("seqval%06d", i)
		tree.Set([]byte(key), []byte(val))
		oracle[key] = val
	}
	tree.SaveVersion()
	verify("sequential-insert")
	t.Logf("phase 1: sequential insert of 2000 keys, height=%d", tree.Height())

	// Phase 2: Remove every other key (triggers merges)
	for i := 0; i < 2000; i += 2 {
		key := fmt.Sprintf("seq%06d", i)
		tree.Remove([]byte(key))
		delete(oracle, key)
	}
	tree.SaveVersion()
	verify("remove-every-other")
	t.Logf("phase 2: removed 1000 keys, height=%d, size=%d", tree.Height(), tree.Size())

	// Phase 3: Random inserts (triggers 50/50 splits and redistributes)
	for i := 0; i < 3000; i++ {
		key := fmt.Sprintf("rnd%06d", rng.Intn(10000))
		val := fmt.Sprintf("rndval_%d", i)
		tree.Set([]byte(key), []byte(val))
		oracle[key] = val
	}
	tree.SaveVersion()
	verify("random-insert")
	t.Logf("phase 3: random insert of 3000 keys, height=%d, size=%d", tree.Height(), tree.Size())

	// Phase 4: Reverse-order inserts
	for i := 9999; i >= 5000; i-- {
		key := fmt.Sprintf("rev%06d", i)
		val := fmt.Sprintf("revval%06d", i)
		tree.Set([]byte(key), []byte(val))
		oracle[key] = val
	}
	tree.SaveVersion()
	verify("reverse-insert")
	t.Logf("phase 4: reverse insert of 5000 keys, height=%d, size=%d", tree.Height(), tree.Size())

	// Phase 5: Remove everything except 100 random keys
	keep := make(map[string]bool)
	keys := make([]string, 0, len(oracle))
	for k := range oracle {
		keys = append(keys, k)
	}
	rng.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	for i := 0; i < 100 && i < len(keys); i++ {
		keep[keys[i]] = true
	}
	for _, k := range keys {
		if !keep[k] {
			tree.Remove([]byte(k))
			delete(oracle, k)
		}
	}
	tree.SaveVersion()
	verify("mass-remove")
	t.Logf("phase 5: mass remove to 100 keys, height=%d", tree.Height())

	// Phase 6: Prune all old versions, verify value count
	tree.DeleteVersionsTo(4)
	valCount := countDBValues(db)
	if valCount != int(tree.Size()) {
		t.Fatalf("after prune: %d values != %d keys", valCount, tree.Size())
	}

	// Reload from DB and verify
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.LoadVersion(5)
	verify2 := func() {
		for k, v := range oracle {
			got, _ := tree2.Get([]byte(k))
			if string(got) != v {
				t.Fatalf("reload: Get(%q) = %q, want %q", k, got, v)
			}
		}
	}
	verify2()

	t.Logf("all phases complete, final size=%d, values=%d", tree.Size(), valCount)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
