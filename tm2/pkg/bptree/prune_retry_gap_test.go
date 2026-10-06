package bptree

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// A mid-prune threshold flush commits the deletion of leading versions; when
// an error then aborts the prune, the retry must resume across the resulting
// version gap (the !exists -> continue skip) and complete. This is the
// crash-consistency story PruneVersionsTo documents ("Intermediate commits
// are safe: pruning is idempotent") — nothing else combines a flush with a
// mid-range failure and a retry.
func TestPruneVersionsTo_FlushThenErrorRetryResumes(t *testing.T) {
	fdb := &failingKeyHasDB{DB: memdb.NewMemDB()}
	tree := NewMutableTreeWithDB(fdb, 0, NewNopLogger(), FlushThresholdOption(128))

	// Overwrite the same keys every version so each pruned version carries a
	// full set of orphaned value records — enough delete volume to cross the
	// flush threshold on every iteration.
	const versions = 8
	for v := 1; v <= versions; v++ {
		for k := range 30 {
			if _, err := tree.Set(fmt.Appendf(nil, "k%04d", k), fmt.Appendf(nil, "v%d_%d", v, k)); err != nil {
				t.Fatal(err)
			}
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatalf("save v%d: %v", v, err)
		}
	}

	// Arm: versionExistsE(failVer) errors mid-range, after the 128-byte
	// threshold has already flushed earlier versions' deletions.
	const failVer = 5
	fdb.failKey = rootDBKey(failVer)
	fdb.armed = true
	to := int64(versions - 1)
	if err := tree.DeleteVersionsTo(to); err == nil {
		t.Fatal("armed prune succeeded; want error")
	}
	fdb.armed = false

	// Partial progress was committed (leading versions gone) while everything
	// from just below the failure point on survives intact.
	if tree.VersionExists(1) {
		t.Fatal("v1 still exists; expected its flushed deletion to have committed")
	}
	for v := int64(failVer - 1); v <= int64(versions); v++ {
		if !tree.VersionExists(v) {
			t.Fatalf("v%d missing after aborted prune", v)
		}
	}

	// The disarmed retry resumes across the gap and completes.
	if err := tree.DeleteVersionsTo(to); err != nil {
		t.Fatalf("retry: %v", err)
	}
	for v := int64(1); v <= to; v++ {
		if tree.VersionExists(v) {
			t.Fatalf("v%d still exists after retry", v)
		}
	}
	if !tree.VersionExists(versions) {
		t.Fatalf("v%d (latest) missing after retry", versions)
	}

	// Cold reload: discoverVersions over the post-gap state.
	tree2 := NewMutableTreeWithDB(fdb, 0, NewNopLogger())
	v, err := tree2.Load()
	if err != nil {
		t.Fatalf("cold load: %v", err)
	}
	if v != versions {
		t.Fatalf("cold load at v%d, want %d", v, versions)
	}
	got, err := tree2.Get([]byte("k0000"))
	if err != nil || string(got) != fmt.Sprintf("v%d_0", versions) {
		t.Fatalf("post-reload Get = %q, %v", got, err)
	}
}
