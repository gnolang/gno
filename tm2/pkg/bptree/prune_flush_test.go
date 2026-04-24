package bptree

import (
	"fmt"
	"sync/atomic"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// batchCountingDB wraps a DB and counts calls to NewBatch, which tells us
// how many times the pruner flushed its batch mid-run.
type batchCountingDB struct {
	dbm.DB
	newBatchCalls uint64
}

func (c *batchCountingDB) NewBatch() dbm.Batch {
	atomic.AddUint64(&c.newBatchCalls, 1)
	return c.DB.NewBatch()
}

func (c *batchCountingDB) NewBatchWithSize(size int) dbm.Batch {
	atomic.AddUint64(&c.newBatchCalls, 1)
	return c.DB.NewBatchWithSize(size)
}

func (c *batchCountingDB) batches() uint64 { return atomic.LoadUint64(&c.newBatchCalls) }

// TestPruneVersionsTo_FlushesBatchUnderThreshold verifies that pruning a
// long run of versions does not accumulate an unbounded batch. With a
// very low FlushThreshold, Commit should be called multiple times as the
// loop progresses, rather than only once at the end.
func TestPruneVersionsTo_FlushesBatchUnderThreshold(t *testing.T) {
	cdb := &batchCountingDB{DB: memdb.NewMemDB()}
	// Set an aggressively low threshold so we hit it within a few versions.
	tree := NewMutableTreeWithDB(cdb, 100, NewNopLogger(),
		FlushThresholdOption(128)) // 128 bytes

	// Create many versions so the prune loop iterates enough to flush.
	const numVersions = 20
	for v := 0; v < numVersions; v++ {
		for i := 0; i < 30; i++ {
			tree.Set(fmt.Appendf(nil, "k%04d", i+v*30), fmt.Appendf(nil, "v%d_%d", v, i))
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatalf("SaveVersion: %v", err)
		}
	}

	// Baseline: batches created so far (one per SaveVersion Commit).
	baseline := cdb.batches()

	// Prune most versions (keep the latest).
	if err := tree.DeleteVersionsTo(int64(numVersions - 1)); err != nil {
		t.Fatalf("DeleteVersionsTo: %v", err)
	}

	flushes := cdb.batches() - baseline
	if flushes < 2 {
		t.Fatalf("prune only created %d batches — expected multiple flushes under a 128-byte threshold over %d versions",
			flushes, numVersions-1)
	}

	// The latest version must still be readable and correct.
	tree2 := NewMutableTreeWithDB(cdb, 100, NewNopLogger())
	latestV, err := tree2.LoadVersion(int64(numVersions))
	if err != nil {
		t.Fatalf("LoadVersion: %v", err)
	}
	if latestV != int64(numVersions) {
		t.Fatalf("loaded version = %d, want %d", latestV, numVersions)
	}
	if tree2.Size() != int64(numVersions*30) {
		t.Fatalf("size after prune = %d, want %d", tree2.Size(), numVersions*30)
	}
}

// TestPruneVersionsTo_ZeroFlushThresholdUsesDefault verifies that the
// default/unset threshold does not degrade to per-version flushing (which
// would be correct but slow).
func TestPruneVersionsTo_ZeroFlushThresholdUsesDefault(t *testing.T) {
	cdb := &batchCountingDB{DB: memdb.NewMemDB()}
	tree := NewMutableTreeWithDB(cdb, 100, NewNopLogger()) // default threshold

	for v := 0; v < 10; v++ {
		tree.Set(fmt.Appendf(nil, "k%d", v), []byte("v"))
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatalf("SaveVersion: %v", err)
		}
	}

	baseline := cdb.batches()
	if err := tree.DeleteVersionsTo(9); err != nil {
		t.Fatalf("DeleteVersionsTo: %v", err)
	}
	flushes := cdb.batches() - baseline
	// At the default 4 MiB threshold, a small prune should fit in one batch.
	if flushes > 2 {
		t.Fatalf("default threshold flushed %d times for a tiny prune; expected 1", flushes)
	}
}
