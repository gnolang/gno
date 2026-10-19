package rootmulti_test

// Regression tests for https://github.com/gnolang/gno/issues/6011.
//
// The original bug chain: gnoland mounts its stores with a dedicated db, so
// constructStore routed the "immutable" query multistore's store reads (and,
// via a real writable batch, WRITES) to the LIVE cfg.DB instead of the query
// snapshot; the query ABCI connection runs concurrently with consensus; and
// bptree's Store.LoadVersion (Immutable branch) ran ensureFastIndex, whose
// version-scan (iterator, pre-commit view) and stamp-read (Get, post-commit)
// could straddle a concurrent commit for block N — misreading the index as
// stale and silently rebuilding ALL fast-index entries from root@N-1: old
// values persisted under a stamp that later commits re-validate, served on
// the consensus read path -> AppHash divergence.
//
// Fix layers exercised here, using a hook DB that commits block N at the
// exact moment the fast-index stamp is first read (the racing interleaving,
// made deterministic):
//
//  1. Immutable store loads use MutableTree.LoadReadonly, which performs NO
//     fast-index maintenance (never writes). The one stamp read they do make
//     is getImmutable's advisory snapshot gate, and the view stays
//     version-correct even when a commit races that read.
//  2. Even a FULL Load() straddling the commit (stamp ahead of the loaded
//     version) must NOT rebuild: ensureFastIndex treats stamp > version as
//     "out-of-contract racing reader or external rewind" and fails loud,
//     leaving the index untouched.

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	bp "github.com/gnolang/gno/tm2/pkg/bptree"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/rootmulti"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// hookDB delegates to db and runs onStampGet once, the first time the
// fast-index stamp key is read. It stands in for the consensus drain landing
// between a query-load's discoverVersions scan and its stamp read. The
// "fastidx" suffix couples to bptree's unexported metaFastVersionKey; drift
// is guarded by layer 2's mandatory hasFired assertion below.
type hookDB struct {
	dbm.DB
	mu         sync.Mutex
	onStampGet func()
	fired      bool
}

func (h *hookDB) Get(key []byte) ([]byte, error) {
	if bytes.HasSuffix(key, []byte("fastidx")) {
		h.mu.Lock()
		if !h.fired && h.onStampGet != nil {
			h.fired = true
			h.mu.Unlock()
			h.onStampGet()
		} else {
			h.mu.Unlock()
		}
	}
	return h.DB.Get(key)
}

func (h *hookDB) hasFired() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.fired
}

func TestFastIndex_NoStaleRebuildOnRacingQueryLoad(t *testing.T) {
	db, err := dbm.NewDB("gnolang", dbm.PebbleDBBackend, t.TempDir())
	if err != nil {
		t.Fatalf("pebble: %v", err)
	}
	defer db.Close()

	mainKey := types.NewStoreKey("main")
	ms := rootmulti.NewMultiStore(db)
	ms.SetStoreOptions(types.StoreOptions{PruningOptions: types.NewPruningOptions(0, 1)})
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	if err := ms.LoadLatestVersion(); err != nil {
		t.Fatalf("load: %v", err)
	}
	defer ms.Close() // release the query snapshot before db.Close

	acct := []byte("/a/account-under-test")
	oldVal := []byte("balance=999999996959000000,seq=226") // written at an early block
	midVal := []byte("balance=999999996943000000,seq=227") // written at block N (races load 1)
	newVal := []byte("balance=999999996940000000,seq=228") // written at block N+1 (races load 2)

	commit := func(kvs map[string][]byte) {
		cms := ms.MultiCacheWrap()
		st := cms.GetStore(mainKey)
		for k, v := range kvs {
			st.Set(nil, []byte(k), v)
		}
		cms.MultiWrite()
		ms.Commit()
	}

	// Blocks 1..3: establish history; the account's last write lands at block 2.
	commit(map[string][]byte{"filler-a": []byte("x1")})
	commit(map[string][]byte{string(acct): oldVal, "filler-b": []byte("x2")})
	commit(map[string][]byte{"filler-c": []byte("x3")})

	// --- Layer 1: an immutable-store load is READ-ONLY and version-correct
	// even when a commit races it. The store is built by hand over the LIVE db
	// (the PRE-fix configuration that used to write the poison). The load reads
	// the fast-index stamp exactly once — the getImmutable snapshot gate, an
	// advisory read, not maintenance — and the hook lands block N (acct→midVal)
	// during that read. A view pinned at version 2 must return the version-2
	// value: the gate keeps the index on (stamp ≥ 2), and fastGet's per-entry
	// guard rejects the racing newer entry and walks to the version-2 root.
	// Write-freedom is asserted by the end-of-test raw-DB audit below.
	hdb := &hookDB{DB: db}
	hdb.onStampGet = func() {
		commit(map[string][]byte{string(acct): midVal, "filler-d": []byte("x4")})
	}
	queryStore := storebptree.FastStoreConstructor(
		dbm.NewPrefixDB(hdb, []byte("s/_/")),
		types.StoreOptions{Immutable: true},
	).(*storebptree.Store)
	if err := queryStore.LoadVersion(2); err != nil {
		t.Fatalf("query-path LoadVersion: %v", err)
	}
	if !hdb.hasFired() {
		t.Fatal("immutable load did not read the fast-index stamp (getImmutable gate probe unfired)")
	}
	if got := queryStore.Get(nil, acct); !bytes.Equal(got, oldVal) {
		t.Fatalf("version-2 view returned %q; want %q (gate on, per-entry guard rejects the racing entry)", got, oldVal)
	}

	// --- Layer 2: a FULL Load() (maintenance-capable) whose stamp read
	// straddles the next commit observes stamp > version and must FAIL LOUD —
	// neither rewriting the index from its older root (the gno#6011 poisoning)
	// nor silently proceeding.
	hdb2 := &hookDB{DB: db}
	hdb2.onStampGet = func() {
		commit(map[string][]byte{string(acct): newVal, "filler-e": []byte("x5")})
	}
	racingTree := bp.NewMutableTreeWithDB(
		dbm.NewPrefixDB(hdb2, []byte("s/_/")), 256, bp.NewNopLogger(), bp.FastIndexOption(true))
	if _, err := racingTree.Load(); err == nil {
		t.Fatal("racing full Load succeeded; want a stamp-ahead error")
	} else if !strings.Contains(err.Error(), "ahead of the loaded version") {
		t.Fatalf("racing full Load failed with the wrong error: %v", err)
	}
	if !hdb2.hasFired() {
		t.Fatal("test setup: full Load did not read the stamp (hook unfired)")
	}

	// More blocks; the stamp catches back up to latest.
	commit(map[string][]byte{"filler-f": []byte("x6")})
	commit(map[string][]byte{"filler-g": []byte("x7")})

	// Audit persisted state from fresh handles, as a restarted node would.
	pdb := dbm.NewPrefixDB(db, []byte("s/_/"))
	plain := bp.NewMutableTreeWithDB(pdb, 256, bp.NewNopLogger())
	if _, err := plain.Load(); err != nil {
		t.Fatalf("plain load: %v", err)
	}
	fast := bp.NewMutableTreeWithDB(pdb, 256, bp.NewNopLogger(), bp.FastIndexOption(true))
	if _, err := fast.Load(); err != nil {
		t.Fatalf("fast load: %v", err)
	}

	walkV, err := plain.Get(acct)
	if err != nil {
		t.Fatalf("walk get: %v", err)
	}
	fastV, err := fast.Get(acct)
	if err != nil {
		t.Fatalf("fast get: %v", err)
	}
	if !bytes.Equal(walkV, newVal) {
		t.Fatalf("tree corrupted: walk=%q want=%q", walkV, newVal)
	}
	if !bytes.Equal(fastV, walkV) {
		t.Fatalf("issue #6011 regressed: fast-index Get returned stale value %q, tree has %q (stale=%v)",
			fastV, walkV, bytes.Equal(fastV, midVal) || bytes.Equal(fastV, oldVal))
	}
}
