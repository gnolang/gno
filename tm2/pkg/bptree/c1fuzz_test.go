package bptree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"

	ics23 "github.com/cosmos/ics23/go"
)

// The C1 fuzzer: model-based op programs interleaving prune with every other
// operation. One engine, three entry points: FuzzTreeOps (native fuzzing),
// TestSoak_TreeOps (env-gated continuous soak),
// TestStress_ConcurrentSanctionedReaders (seeded -race).
// Two decoders: runOpChunk (byte-frozen by the committed corpus) and
// runTideChunk (soak-only tide).

type fuzzCfg struct {
	keys        int   // keyspace size
	window      int64 // retained-version window W (forced prune cadence)
	holdBudget  int   // ops before a held snapshot auto-releases
	sessionCap  int   // mutations without a save before a forced SaveVersion
	maxOps      int   // decoded-op cap per program
	cacheSize   int
	allowImport bool
	maxImports  int // per-program cap: each import leaks values below the wall (M21)
	allowInject bool
	tideHigh    int // tide crest for runTideChunk (soak-only); read only by its flip check
}

func defaultFuzzCfg() fuzzCfg {
	return fuzzCfg{
		keys:        800,
		window:      4,
		holdBudget:  24,
		sessionCap:  256,
		maxOps:      2048,
		cacheSize:   256,
		allowImport: true,
		maxImports:  32,
		allowInject: true,
	}
}

type heldSnap struct {
	imm      *ImmutableTree
	expireAt int
}

// soakStats counts branch firings so a long soak run reports what it actually
// covered. Deterministic per seed — diffable across runs as a regression signal.
type soakStats struct {
	flips, emptySaves, twinSaves         int
	injectArmed, injectFired             int
	cell5Blocked                         int
	adoptions, adoptExistsErrs           int
	rollbacks, loadOlds, coldRestarts    int
	shrinkWraps, boundedIters, emptySets int
	maxPruneWidth                        int64
}

type fuzzState struct {
	tb   testing.TB
	cfg  fuzzCfg
	fdb  *failingGetDB // wraps the memdb; nodeDB captures this handle
	tree *MutableTree

	model  map[string]string           // working overlay (live keys)
	snaps  map[int64]map[string]string // committed version -> content
	hashes map[int64][]byte            // committed version -> root hash
	holds  map[int64]*heldSnap         // version -> held registered snapshot

	first, latest int64 // model's retained range; 0,0 = nothing saved
	dirty         bool  // effective mutation since the last session boundary
	opN           int
	mutSinceSave  int
	imports       int   // imports so far, capped at cfg.maxImports
	maxImportVer  int64 // vk-version wall (R2); 0 = no import yet
	rising        bool  // tide direction: growing toward tideHigh vs draining to 0
	lastFlipOp    int   // opN at the last tide flip (stall detection)
	stats         soakStats
}

func newFuzzState(tb testing.TB, cfg fuzzCfg) *fuzzState {
	tb.Helper()
	fdb := &failingGetDB{DB: memdb.NewMemDB()}
	return &fuzzState{
		tb:     tb,
		cfg:    cfg,
		fdb:    fdb,
		tree:   NewMutableTreeWithDB(fdb, cfg.cacheSize, NewNopLogger()),
		model:  map[string]string{},
		snaps:  map[int64]map[string]string{},
		hashes: map[int64][]byte{},
		holds:  map[int64]*heldSnap{},
	}
}

func (st *fuzzState) key(i int) string   { return fmt.Sprintf("fz%05d", i%st.cfg.keys) }
func (st *fuzzState) nzKey(i int) string { return fmt.Sprintf("nz%04d", i) } // disjoint sub-keyspace (R8)

func snapCopy(m map[string]string) map[string]string {
	c := make(map[string]string, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

func (st *fuzzState) sortedModelKeys() []string {
	ks := make([]string, 0, len(st.model))
	for k := range st.model {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// --- expectedPrune: the R6 precedence table as a single predicate ---

type pruneExp int

const (
	expNil pruneExp = iota
	expPlainErr
	expUncommitted
	expActiveReaders
)

func (st *fuzzState) expectedPrune(toVersion int64) pruneExp {
	switch {
	case toVersion >= st.latest:
		return expPlainErr // cell 1: plain "cannot prune latest" (incl. latest==0)
	case toVersion < st.first:
		return expNil // cell 2: no-op
	case st.dirty:
		return expUncommitted // cell 3
	case st.tree.Version() <= toVersion:
		return expActiveReaders // cell 4: loaded-version guard
	default:
		for v := range st.holds {
			if v >= st.first && v <= toVersion {
				return expActiveReaders // cell 5: held reader in range
			}
		}
		return expNil // cell 6
	}
}

// doPrune executes a prune, asserts the predicate's outcome, applies the
// model transition on success, and runs the full oracle after real work.
func (st *fuzzState) doPrune(toVersion int64) {
	exp := st.expectedPrune(toVersion)
	if exp == expActiveReaders && st.tree.Version() > toVersion {
		st.stats.cell5Blocked++ // held-reader block (cell 5), not the loaded-version guard
	}
	err := st.tree.DeleteVersionsTo(toVersion)
	switch exp {
	case expNil:
		if err != nil {
			st.tb.Fatalf("op %d: prune(%d) expected nil, got %v (first=%d latest=%d loaded=%d dirty=%v holds=%d)",
				st.opN, toVersion, err, st.first, st.latest, st.tree.Version(), st.dirty, len(st.holds))
		}
		if toVersion >= st.first { // real work (not the cell-2 no-op)
			if w := toVersion - st.first + 1; w > st.stats.maxPruneWidth {
				st.stats.maxPruneWidth = w
			}
			for v := range st.snaps {
				if v <= toVersion {
					delete(st.snaps, v)
					delete(st.hashes, v)
				}
			}
			st.first = toVersion + 1
			st.fullOracle()
		}
	case expPlainErr:
		if err == nil || errors.Is(err, ErrUncommittedChanges) || errors.Is(err, ErrActiveReaders) {
			st.tb.Fatalf("op %d: prune(%d) expected plain error, got %v", st.opN, toVersion, err)
		}
	case expUncommitted:
		if !errors.Is(err, ErrUncommittedChanges) {
			st.tb.Fatalf("op %d: prune(%d) expected ErrUncommittedChanges, got %v", st.opN, toVersion, err)
		}
	case expActiveReaders:
		if !errors.Is(err, ErrActiveReaders) {
			st.tb.Fatalf("op %d: prune(%d) expected ErrActiveReaders, got %v", st.opN, toVersion, err)
		}
	}
}

// --- ops ---

func (st *fuzzState) doSet(k string) {
	v := fmt.Sprintf("v%d", st.opN)
	if st.opN%101 == 0 {
		v = "" // empty value: the only value-shape-dependent path (empty-vs-missing)
		st.stats.emptySets++
	}
	if _, err := st.tree.Set([]byte(k), []byte(v)); err != nil {
		st.tb.Fatalf("op %d: Set(%s): %v", st.opN, k, err)
	}
	st.model[k] = v
	st.dirty = true
	st.mutSinceSave++
}

func (st *fuzzState) doSetSame(k string) {
	cur, ok := st.model[k]
	if !ok {
		return // no-op byte
	}
	if _, err := st.tree.Set([]byte(k), []byte(cur)); err != nil {
		st.tb.Fatalf("op %d: SetSame(%s): %v", st.opN, k, err)
	}
	st.dirty = true
	st.mutSinceSave++
}

func (st *fuzzState) doRemove(k string) {
	if _, ok := st.model[k]; !ok {
		return // no-op byte (Remove of an absent key doesn't dirty the session)
	}
	if _, _, err := st.tree.Remove([]byte(k)); err != nil {
		st.tb.Fatalf("op %d: Remove(%s): %v", st.opN, k, err)
	}
	delete(st.model, k)
	st.dirty = true
	st.mutSinceSave++
}

func (st *fuzzState) doNetZero(i int) {
	k := st.nzKey(i)
	if _, err := st.tree.Set([]byte(k), []byte("nz")); err != nil {
		st.tb.Fatalf("op %d: NetZero set: %v", st.opN, err)
	}
	if _, _, err := st.tree.Remove([]byte(k)); err != nil {
		st.tb.Fatalf("op %d: NetZero remove: %v", st.opN, err)
	}
	st.dirty = true // nonce advanced; session is dirty even though content is unchanged
	st.mutSinceSave++
}

func (st *fuzzState) doSave() {
	if len(st.model) == 0 {
		st.stats.emptySaves++
	}
	if st.mutSinceSave == 0 {
		st.stats.twinSaves++ // clean-session save: content-identical adjacent versions
	}
	h, v, err := st.tree.SaveVersion()
	if err != nil {
		st.tb.Fatalf("op %d: SaveVersion: %v", st.opN, err)
	}
	st.snaps[v] = snapCopy(st.model)
	st.hashes[v] = append([]byte(nil), h...)
	if st.first == 0 {
		st.first = v
	}
	st.latest = v
	st.dirty = false
	st.mutSinceSave = 0
	if got := countPinned(st.tree.root); st.tree.root != nil && got != 1 {
		st.tb.Fatalf("op %d: post-save pinned nodes = %d, want 1", st.opN, got)
	}
	// Forced prune cadence: keep the retained window bounded by construction.
	if st.latest-st.first+1 > st.cfg.window {
		st.catchUp()
	}
}

func (st *fuzzState) catchUp() {
	st.releaseHolds(true)
	if st.dirty {
		st.doRollback()
	}
	target := st.latest - st.cfg.window
	if target < st.first {
		return
	}
	if exp := st.expectedPrune(target); exp != expNil {
		st.tb.Fatalf("op %d: catch-up prune(%d) blocked (exp=%d) — reader leak?", st.opN, target, exp)
	}
	st.doPrune(target)
	if st.latest-st.first+1 > 2*st.cfg.window {
		st.tb.Fatalf("op %d: window %d exceeds 2×W — genuine reader leak", st.opN, st.latest-st.first+1)
	}
}

func (st *fuzzState) doRollback() {
	st.stats.rollbacks++
	st.tree.Rollback()
	if st.latest > 0 {
		st.model = snapCopy(st.snaps[st.latest])
	} else {
		st.model = map[string]string{}
	}
	st.dirty = false
	st.mutSinceSave = 0
}

func (st *fuzzState) doGrowWave(start, n int) {
	for i := 0; i < n; i++ {
		st.doSet(st.key(start + i))
	}
}

// doShrinkWave removes up to n live keys starting at the first live key at or
// above the start slot, wrapping to the smallest live keys if too few remain.
// Targeting live keys (unlike doRemove's blind selector) is what lets the
// falling tide actually reach 0 instead of decaying asymptotically. The single
// model scan keeps only the n smallest candidates — sorting all live keys per
// wave would dominate the falling phase.
func (st *fuzzState) doShrinkWave(start, n int) {
	from := st.key(start)
	pick := make([]string, 0, n)
	for k := range st.model {
		if k >= from {
			pick = insertSmallest(pick, k, n)
		}
	}
	if rest := n - len(pick); rest > 0 {
		low := make([]string, 0, rest)
		for k := range st.model {
			if k < from {
				low = insertSmallest(low, k, rest)
			}
		}
		if len(low) > 0 {
			st.stats.shrinkWraps++
		}
		pick = append(pick, low...)
	}
	for _, k := range pick {
		st.doRemove(k)
	}
}

// insertSmallest inserts k into the sorted slice, keeping only the limit smallest.
func insertSmallest(ks []string, k string, limit int) []string {
	if len(ks) == limit && k >= ks[limit-1] {
		return ks // not among the smallest — the common case at scale
	}
	i, _ := slices.BinarySearch(ks, k)
	if i >= limit {
		return ks
	}
	ks = slices.Insert(ks, i, k)
	if len(ks) > limit {
		ks = ks[:limit]
	}
	return ks
}

func (st *fuzzState) doDrainAll() {
	for _, k := range st.sortedModelKeys() {
		st.doRemove(k)
	}
}

func (st *fuzzState) retainedVersions() []int64 {
	vs := make([]int64, 0, len(st.snaps))
	for v := range st.snaps {
		vs = append(vs, v)
	}
	sort.Slice(vs, func(i, j int) bool { return vs[i] < vs[j] })
	return vs
}

func (st *fuzzState) pickRetained(sel byte) (int64, bool) {
	vs := st.retainedVersions()
	if len(vs) == 0 {
		return 0, false
	}
	return vs[int(sel)%len(vs)], true
}

func (st *fuzzState) doHold(sel byte) {
	v, ok := st.pickRetained(sel)
	if !ok {
		return
	}
	if _, held := st.holds[v]; held {
		return
	}
	imm, err := st.tree.GetImmutable(v)
	if err != nil {
		st.tb.Fatalf("op %d: Hold GetImmutable(%d): %v", st.opN, v, err)
	}
	st.holds[v] = &heldSnap{imm: imm, expireAt: st.opN + st.cfg.holdBudget}
}

// releaseHolds releases expired holds (or all of them).
func (st *fuzzState) releaseHolds(all bool) {
	for v, h := range st.holds {
		if all || st.opN >= h.expireAt {
			h.imm.Close()
			delete(st.holds, v)
		}
	}
}

// doSnapshotReads verifies up to 3 entries of a retained version against its
// snapshot, iterating from a selector-derived interior point — an O(log n)
// seek that reaches arbitrary paths without sorting the snapshot.
func (st *fuzzState) doSnapshotReads(sel byte) {
	v, ok := st.pickRetained(sel)
	if !ok {
		return
	}
	imm, err := st.tree.GetImmutable(v)
	if err != nil {
		st.tb.Fatalf("op %d: GetImmutable(%d): %v", st.opN, v, err)
	}
	defer imm.Close()
	it, err := imm.Iterator([]byte(st.key(int(sel)<<8)), nil, true)
	if err != nil {
		st.tb.Fatalf("op %d: v%d iterator: %v", st.opN, v, err)
	}
	defer it.Close()
	snap := st.snaps[v]
	for i := 0; i < 3 && it.Valid(); i++ {
		k := string(it.Key())
		if string(it.Value()) != snap[k] {
			st.tb.Fatalf("op %d: v%d iter %s = %q (%v); want %q", st.opN, v, k, it.Value(), it.Error(), snap[k])
		}
		got, err := imm.Get(it.Key())
		if err != nil || string(got) != snap[k] {
			st.tb.Fatalf("op %d: v%d Get(%s) = %q, %v; want %q", st.opN, v, k, got, err, snap[k])
		}
		it.Next()
	}
	if err := it.Error(); err != nil {
		st.tb.Fatalf("op %d: v%d iterator error: %v", st.opN, v, err)
	}
}

// doIteratePartial walks up to sel%97 steps, descending when sel >= 128 — the
// harness's only descending-iterator coverage; strides past 32 cross leaf
// boundaries (B=32) in both directions. Every third selector bounds the walk
// to a selector-derived [start, end) window — the only end-bounded iteration
// at depth. Any yield outside the window (including from a wrapped/inverted
// window, which must yield nothing) fails.
func (st *fuzzState) doIteratePartial(sel byte) {
	v, ok := st.pickRetained(sel)
	if !ok {
		return
	}
	imm, err := st.tree.GetImmutable(v)
	if err != nil {
		st.tb.Fatalf("op %d: GetImmutable(%d): %v", st.opN, v, err)
	}
	defer imm.Close()
	var start, end []byte
	if sel%3 == 0 {
		lo := (int(sel) * 257) % st.cfg.keys
		span := 1 + (int(sel)*7)%500
		start, end = []byte(st.key(lo)), []byte(st.key(lo+span))
		st.stats.boundedIters++
	}
	it, err := imm.Iterator(start, end, sel < 128)
	if err != nil {
		st.tb.Fatalf("op %d: Iterator(%d): %v", st.opN, v, err)
	}
	defer it.Close()
	for i := 0; i < int(sel)%97 && it.Valid(); i++ {
		if k := it.Key(); start != nil && (bytes.Compare(k, start) < 0 || bytes.Compare(k, end) >= 0) {
			st.tb.Fatalf("op %d: v%d bounded iterator yielded %s outside [%s,%s)", st.opN, v, k, start, end)
		}
		it.Next()
	}
	if err := it.Error(); err != nil {
		st.tb.Fatalf("op %d: v%d iterator error: %v", st.opN, v, err)
	}
}

func (st *fuzzState) doLoadOld(sel byte) {
	vs := st.retainedVersions()
	if len(vs) < 2 {
		return // R7: with one retained version a covering prune is the cell-1 plain error
	}
	st.stats.loadOlds++
	v := vs[int(sel)%(len(vs)-1)] // any retained version < latest
	// Entry pin: LoadVersion discards any staged session; the model drops its
	// overlay at op START and the working view becomes the v snapshot.
	if _, err := st.tree.LoadVersion(v); err != nil {
		st.tb.Fatalf("op %d: LoadVersion(%d): %v", st.opN, v, err)
	}
	st.model = snapCopy(st.snaps[v])
	st.dirty = false
	st.mutSinceSave = 0

	// Covering prune: per expectedPrune (cell 4).
	st.doPrune(v)
	// Below prune: per expectedPrune (nil unless a hold covers the range).
	if v-1 >= st.first {
		below := v - 1
		exp := st.expectedPrune(below)
		st.doPrune(below)
		if exp == expNil {
			// The loaded view must still read correctly after a below-prune.
			for _, k := range st.sortedModelKeys() {
				got, err := st.tree.Get([]byte(k))
				if err != nil || string(got) != st.model[k] {
					st.tb.Fatalf("op %d: loaded view broken after below-prune: Get(%s)=%q,%v", st.opN, k, got, err)
				}
				break // one key suffices per op
			}
		}
	}
	// SaveVersion while loaded: decidable (R7) — adoption iff hashes match.
	if next, ok := st.hashes[v+1]; ok {
		_, sv, err := st.tree.SaveVersion()
		if bytes.Equal(st.hashes[v], next) {
			if err != nil || sv != v+1 {
				st.tb.Fatalf("op %d: idempotent adoption at v%d: got v%d, %v", st.opN, v+1, sv, err)
			}
			st.stats.adoptions++
			st.model = snapCopy(st.snaps[v+1])
		} else {
			if err == nil || !strings.Contains(err.Error(), "already exists with a different hash") {
				st.tb.Fatalf("op %d: save-while-loaded at v%d: want exists-error, got %v", st.opN, v+1, err)
			}
			st.stats.adoptExistsErrs++
		}
	}
	// Closing recovery: back to latest.
	if _, err := st.tree.Load(); err != nil {
		st.tb.Fatalf("op %d: Load after LoadOld: %v", st.opN, err)
	}
	st.model = snapCopy(st.snaps[st.latest])
	st.dirty = false
	st.mutSinceSave = 0
}

func (st *fuzzState) doColdRestart() {
	// R5: holds register in the old nodeDB's memory — close them first.
	st.releaseHolds(true)
	if n := len(st.tree.ndb.versionReaders); n != 0 {
		st.tb.Fatalf("op %d: %d version readers outstanding before cold restart", st.opN, n)
	}
	if st.latest == 0 {
		return // nothing committed; a fresh tree would have nothing to Load
	}
	st.stats.coldRestarts++
	st.tree = NewMutableTreeWithDB(st.fdb, st.cfg.cacheSize, NewNopLogger())
	v, err := st.tree.Load()
	if err != nil {
		st.tb.Fatalf("op %d: cold Load: %v", st.opN, err)
	}
	if v != st.latest {
		st.tb.Fatalf("op %d: cold Load at v%d, want %d", st.opN, v, st.latest)
	}
	// The model drops its uncommitted overlay (Rollback semantics).
	st.model = snapCopy(st.snaps[st.latest])
	st.dirty = false
	st.mutSinceSave = 0
	for _, k := range st.sortedModelKeys() {
		got, err := st.tree.Get([]byte(k))
		if err != nil || string(got) != st.model[k] {
			st.tb.Fatalf("op %d: cold restart Get(%s)=%q,%v; want %q", st.opN, k, got, err, st.model[k])
		}
	}
}

func (st *fuzzState) doExportImport(sel byte) {
	if !st.cfg.allowImport || st.dirty || st.imports >= st.cfg.maxImports {
		return
	}
	// Source: a retained NON-EMPTY version (Export of an empty tree errors).
	var src int64
	found := false
	for _, v := range st.retainedVersions() {
		if len(st.snaps[v]) > 0 {
			src, found = v, true
			if int(sel)%2 == 0 {
				break // sometimes the oldest, sometimes the newest
			}
		}
	}
	if !found {
		return
	}
	imm, err := st.tree.GetImmutable(src)
	if err != nil {
		st.tb.Fatalf("op %d: export GetImmutable(%d): %v", st.opN, src, err)
	}
	target := st.latest + 1
	exportInto(st.tb, st.tree, imm, target)
	imm.Close()
	// Import Rollbacks the session first and leaves the working tree = the
	// imported content; the model mirrors both.
	st.snaps[target] = snapCopy(st.snaps[src])
	h, err2 := st.tree.GetImmutable(target)
	if err2 != nil {
		st.tb.Fatalf("op %d: post-import GetImmutable(%d): %v", st.opN, target, err2)
	}
	st.hashes[target] = append([]byte(nil), h.Hash()...)
	h.Close()
	st.latest = target
	st.model = snapCopy(st.snaps[src])
	st.dirty = false
	st.mutSinceSave = 0
	st.maxImportVer = target // the vk-version wall (R2) — must precede catchUp's oracle
	st.imports++
	// Forced prune cadence. Imports advance latest like saves do, but at a 4×
	// wider gate: import runs are the only source of wide multi-version prunes,
	// so let them build up before catching up.
	if st.latest-st.first+1 > 4*st.cfg.window {
		st.catchUp()
	}
}

func (st *fuzzState) doInjectError(n byte) {
	if !st.cfg.allowInject {
		return
	}
	// Commit any staged session first: with ~1000-mutation sessions a dirty
	// early-return would make this op nearly dead. The save may cascade into
	// catchUp (holds released, window pruned) before arming — benign.
	if st.dirty {
		st.doSave()
	}
	to := st.latest - 1
	if to < st.first || st.expectedPrune(to) != expNil {
		return // injection branches apply only to a would-succeed (cell-6) prune
	}
	// A warm cache serves every prune read — purge so the injector is reachable.
	if st.tree.ndb.nodeCache != nil {
		st.tree.ndb.nodeCache.Purge()
	}
	// n<<4 spreads the failure 0..4080 reads into the walk: deep enough to land
	// mid-prune at soak tree sizes, shallow enough that the not-fired branch
	// stays reachable. A width-3 prune of ~1k-mutation diffs can stage close
	// to the 100KB flush threshold (options.go), so a mid-prune flush before
	// the error is possible — the fired path below reconciles for it.
	st.stats.injectArmed++
	atomic.StoreInt32(&st.fdb.allow, int32(n)<<4)
	atomic.StoreInt32(&st.fdb.armed, 1)
	err := st.tree.DeleteVersionsTo(to)
	fired := atomic.LoadInt32(&st.fdb.allow) < 0
	atomic.StoreInt32(&st.fdb.armed, 0) // disarm before any oracle runs
	if fired {
		if err == nil {
			st.tb.Fatalf("op %d: injector fired but prune(%d) succeeded", st.opN, to)
		}
		st.stats.injectFired++
		// A mid-prune threshold flush may have committed the deletion of
		// leading versions before the error fired — sanctioned idempotent
		// behavior. Reconcile the model to what actually survived.
		for v := range st.snaps {
			if !st.tree.VersionExists(v) {
				delete(st.snaps, v)
				delete(st.hashes, v)
				if v >= st.first {
					st.first = v + 1
				}
			}
		}
		st.perVersionOracle() // surviving retained versions intact after the failed prune
		// Disarmed retry succeeds — the continuous L2 property.
		st.doPrune(to)
	} else {
		if err != nil {
			st.tb.Fatalf("op %d: injector not fired but prune(%d) errored: %v", st.opN, to, err)
		}
		for v := range st.snaps {
			if v <= to {
				delete(st.snaps, v)
				delete(st.hashes, v)
			}
		}
		st.first = to + 1
		st.fullOracle()
	}
}

// --- oracles ---

func (st *fuzzState) fullOracle() {
	st.garbageOracle()
	st.perVersionOracle()
	st.proofOracle()
	st.bookkeepingOracle()
}

// garbageOracle: exact node accounting always; value accounting is exact
// until the first import, then governed by the vk-version wall (R2).
func (st *fuzzState) garbageOracle() {
	nodes, values := collectReachable(st.tb, st.tree)
	it, err := st.tree.ndb.db.Iterator(nil, nil)
	if err != nil {
		st.tb.Fatal(err)
	}
	defer it.Close()
	for ; it.Valid(); it.Next() {
		k := it.Key()
		if len(k) == 0 {
			continue
		}
		switch k[0] {
		case PrefixNode:
			if !nodes[string(k[1:])] {
				st.tb.Fatalf("op %d: LEAK: node record %x unreachable", st.opN, k[1:])
			}
		case PrefixOrphan:
			// Orphan records exist only for (first, latest]: orphans[first]
			// is consumed when its predecessor is pruned.
			if v := int64(binary.BigEndian.Uint64(k[1:])); v <= st.first || v > st.latest {
				st.tb.Fatalf("op %d: LEAK: orphan record for v%d outside (%d,%d]", st.opN, v, st.first, st.latest)
			}
		case PrefixRoot:
			if v := int64(binary.BigEndian.Uint64(k[1:])); v < st.first || v > st.latest {
				st.tb.Fatalf("op %d: LEAK: root record for v%d outside [%d,%d]", st.opN, v, st.first, st.latest)
			}
		case PrefixVal:
			if values[string(k[1:])] {
				continue
			}
			vk := k[1:]
			if st.maxImportVer > 0 && len(vk) >= 8 &&
				int64(binary.BigEndian.Uint64(vk[:8])) < st.maxImportVer {
				continue // tolerated below the import wall (M21)
			}
			st.tb.Fatalf("op %d: LEAK: value record %x referenced by no retained leaf", st.opN, k[1:])
		default:
			st.tb.Fatalf("op %d: unknown record prefix %q (key %x)", st.opN, k[0], k)
		}
	}
}

func (st *fuzzState) perVersionOracle() {
	for v, snap := range st.snaps {
		imm, err := st.tree.GetImmutable(v)
		if err != nil {
			st.tb.Fatalf("op %d: oracle GetImmutable(%d): %v", st.opN, v, err)
		}
		if !bytes.Equal(imm.Hash(), st.hashes[v]) {
			st.tb.Fatalf("op %d: v%d hash drift", st.opN, v)
		}
		got := 0
		imm.Iterate(func(k, val []byte) bool {
			if snap[string(k)] != string(val) {
				st.tb.Fatalf("op %d: v%d key %s = %q, want %q", st.opN, v, k, val, snap[string(k)])
			}
			got++
			return false
		})
		if got != len(snap) {
			st.tb.Fatalf("op %d: v%d has %d keys, want %d", st.opN, v, got, len(snap))
		}
		imm.Close()
	}
}

func (st *fuzzState) proofOracle() {
	for v, snap := range st.snaps {
		if len(snap) == 0 {
			continue // empty version: both proof kinds return ErrEmptyTree — skip
		}
		imm, err := st.tree.GetImmutable(v)
		if err != nil {
			st.tb.Fatalf("op %d: proof GetImmutable(%d): %v", st.opN, v, err)
		}
		ks := make([]string, 0, len(snap))
		for k := range snap {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		for _, i := range []int{0, len(ks) / 2, len(ks) - 1} {
			k := ks[i]
			if snap[k] == "" {
				continue // ics23 LeafOp rejects empty values — unprovable by construction (the M24 empty-key constraint, value side)
			}
			proof, err := imm.GetMembershipProof([]byte(k))
			if err != nil {
				st.tb.Fatalf("op %d: v%d membership proof(%s): %v", st.opN, v, k, err)
			}
			if !ics23.VerifyMembership(BptreeSpec, st.hashes[v], proof, []byte(k), []byte(snap[k])) {
				st.tb.Fatalf("op %d: v%d membership proof for %s does not verify", st.opN, v, k)
			}
		}
		// Absent-key probes: an edge bracket, plus an interior gap (strictly
		// between a present key and its successor) — the only probe shape that
		// yields a two-sided (left AND right neighbor) NonExist. A probe is
		// skipped when an adjacent leaf holds an empty value: the embedded
		// neighbor ExistenceProof is unprovable (ics23 LeafOp, cf. M24).
		mid := (len(ks) - 1) / 2
		probes := make([]string, 0, 2)
		if v%2 == 0 {
			if snap[ks[len(ks)-1]] != "" {
				probes = append(probes, "zz!") // sorts after the fz/nz keyspaces
			}
		} else if snap[ks[0]] != "" {
			probes = append(probes, "a!") // and before
		}
		if snap[ks[mid]] != "" && (mid+1 >= len(ks) || snap[ks[mid+1]] != "") {
			probes = append(probes, ks[mid]+"!")
		}
		for _, probe := range probes {
			proof, err := imm.GetNonMembershipProof([]byte(probe))
			if err != nil {
				st.tb.Fatalf("op %d: v%d non-membership proof(%s): %v", st.opN, v, probe, err)
			}
			if !ics23.VerifyNonMembership(BptreeSpec, st.hashes[v], proof, []byte(probe)) {
				st.tb.Fatalf("op %d: v%d non-membership proof for %s does not verify", st.opN, v, probe)
			}
		}
		imm.Close()
	}
}

func (st *fuzzState) bookkeepingOracle() {
	avail := st.tree.AvailableVersions()
	if len(avail) != len(st.snaps) {
		st.tb.Fatalf("op %d: AvailableVersions has %d entries, model %d", st.opN, len(avail), len(st.snaps))
	}
	for _, v := range avail {
		if _, ok := st.snaps[int64(v)]; !ok {
			st.tb.Fatalf("op %d: AvailableVersions reports unknown v%d", st.opN, v)
		}
	}
	if got, want := len(st.tree.ndb.versionReaders), len(st.holds); got != want {
		st.tb.Fatalf("op %d: %d version readers, %d holds outstanding", st.opN, got, want)
	}
}

// --- decoder ---

// pruneTarget maps a selector byte to a toVersion exercising each table cell.
func (st *fuzzState) pruneTarget(sel byte) (int64, bool) {
	switch sel % 8 {
	case 0:
		return st.first - 1, true // cell 2 (or cell 1 when nothing saved)
	case 1:
		return -int64(sel%5) - 1, true // negative
	case 2:
		return st.latest, true // cell 1
	case 3:
		return st.first, true // width-1 (cell 6 / cell 1 when first==latest)
	case 4:
		return st.latest - 1, true // wide catch-up
	case 5, 6:
		if st.latest <= st.first {
			return 0, false
		}
		return st.first + int64(sel)%(st.latest-st.first), true // mid-range
	default:
		// Dirty-session composite: arrange a dirty session, prune in-range,
		// then roll back. Requires an in-range target (cell 3 reachability).
		if st.latest <= st.first {
			return 0, false
		}
		st.doSet(st.key(int(sel)))
		st.doPrune(st.first) // expectedPrune sees dirty → ErrUncommittedChanges
		st.doRollback()
		return 0, false
	}
}

// runOpChunk decodes and executes ops against persistent state.
// makePops returns the byte-cursor primitives shared by both decoders; every
// op's byte consumption flows through these, and the corpus freezes it.
func makePops(data []byte) (pop func() (byte, bool), pop2 func() (int, bool)) {
	pos := 0
	pop = func() (byte, bool) {
		if pos >= len(data) {
			return 0, false
		}
		b := data[pos]
		pos++
		return b, true
	}
	pop2 = func() (int, bool) {
		a, ok1 := pop()
		b, ok2 := pop()
		return int(a)<<8 | int(b), ok1 && ok2
	}
	return pop, pop2
}

func runOpChunk(st *fuzzState, data []byte) {
	pop, pop2 := makePops(data)
	ops := 0
	for ops < st.cfg.maxOps {
		op, ok := pop()
		if !ok {
			return
		}
		ops++
		st.opN++
		st.releaseHolds(false)
		if st.mutSinceSave > st.cfg.sessionCap {
			st.doSave() // session-ops cap (R4)
		}
		switch {
		case op < 80:
			if k, ok := pop2(); ok {
				st.doSet(st.key(k))
			}
		case op < 100:
			if k, ok := pop2(); ok {
				st.doSetSame(st.key(k))
			}
		case op < 140:
			if k, ok := pop2(); ok {
				st.doRemove(st.key(k))
			}
		case op < 150:
			if k, ok := pop(); ok {
				st.doNetZero(int(k))
			}
		case op < 175:
			st.doSave()
		case op < 190:
			if sel, ok := pop(); ok {
				if to, doIt := st.pruneTarget(sel); doIt {
					st.doPrune(to)
				}
			}
		case op < 200:
			start, ok1 := pop2()
			n, ok2 := pop()
			if ok1 && ok2 {
				st.doGrowWave(start, 1+int(n)%24)
			}
		case op < 204:
			st.doDrainAll()
		case op < 210:
			st.doRollback()
		case op < 220:
			if sel, ok := pop(); ok {
				st.doHold(sel)
			}
		case op < 228:
			if sel, ok := pop(); ok {
				st.doSnapshotReads(sel)
			}
		case op < 234:
			if sel, ok := pop(); ok {
				st.doIteratePartial(sel)
			}
		case op < 242:
			if sel, ok := pop(); ok {
				st.doLoadOld(sel)
			}
		case op < 246:
			st.doColdRestart()
		case op < 250:
			if sel, ok := pop(); ok {
				st.doExportImport(sel)
			}
		default:
			if n, ok := pop(); ok {
				st.doInjectError(n)
			}
		}
	}
}

// runTideChunk decodes one chunk in tide mode: mutations dominate and follow
// the rising/falling state (grow to cfg.tideHigh, drain to 0, repeat), and
// waves run 1-64 consecutive keys. Sessions commit on the ~1024-mutation
// sessionCap backbone; the byte-247 save is rare jitter. Cheap lifecycle ops
// run at 1/256, but the four session-killers (rollback, load-old, cold
// restart, error injection) are gated to ~1/4096 — at 1/256 they discarded
// ~90% of tide progress before it could commit and the soak never crested.
// The tide is what carries the tree through height transitions both ways
// (1↔4 levels at tideHigh=50k); the fixed-equilibrium decoder above never
// leaves height 2.
func runTideChunk(st *fuzzState, data []byte) {
	pop, pop2 := makePops(data)
	ops := 0
	for ops < st.cfg.maxOps {
		op, ok := pop()
		if !ok {
			return
		}
		ops++
		st.opN++
		st.releaseHolds(false)
		if st.mutSinceSave > st.cfg.sessionCap {
			st.doSave()
		}
		if st.rising && len(st.model) >= st.cfg.tideHigh {
			st.rising = false
			st.lastFlipOp = st.opN
			st.stats.flips++
		} else if !st.rising && len(st.model) == 0 {
			// Commit the drained tree so every cycle retains an empty version;
			// the rising window's catch-up then prunes through it.
			if st.dirty {
				st.doSave()
			}
			st.rising = true
			st.lastFlipOp = st.opN
			st.stats.flips++
		}
		// ~8× the expected half-cycle (rise ≈ 38k ops, fall ≈ 21k): never
		// trips on variance, catches a drift regression within minutes.
		if st.opN-st.lastFlipOp > 300_000 {
			st.tb.Fatalf("op %d: tide stalled (rising=%v, %d live keys, versions [%d,%d])",
				st.opN, st.rising, len(st.model), st.first, st.latest)
		}
		switch {
		case op < 130: // dominant single key: Set rising, Remove falling
			if k, ok := pop2(); ok {
				if st.rising {
					st.doSet(st.key(k))
				} else {
					st.doRemove(st.key(k))
				}
			}
		case op < 195: // counter single key: churn against the tide
			if k, ok := pop2(); ok {
				if st.rising {
					st.doRemove(st.key(k))
				} else {
					st.doSet(st.key(k))
				}
			}
		case op < 215:
			if k, ok := pop2(); ok {
				st.doSetSame(st.key(k))
			}
		case op < 239: // wave: 1-64 consecutive keys in the tide's direction
			start, ok1 := pop2()
			n, ok2 := pop()
			if ok1 && ok2 {
				if st.rising {
					st.doGrowWave(start, 1+int(n)%64)
				} else {
					st.doShrinkWave(start, 1+int(n)%64)
				}
			}
		case op < 247:
			if k, ok := pop(); ok {
				st.doNetZero(int(k))
			}
		case op == 247: // save jitter ~1/1024; the sessionCap backbone sets the cadence
			if b, ok := pop(); ok && b < 64 {
				st.doSave()
			}
		case op == 248:
			if sel, ok := pop(); ok {
				if to, doIt := st.pruneTarget(sel); doIt {
					st.doPrune(to)
				}
			}
		case op == 249:
			if sel, ok := pop(); ok {
				st.doHold(sel)
			}
		case op == 250:
			if sel, ok := pop(); ok {
				st.doSnapshotReads(sel)
			}
		case op == 251:
			if sel, ok := pop(); ok {
				st.doIteratePartial(sel)
			}
		default: // 252-255: session-killers, gated to ~1/4096 (gate byte always
			// consumed): ~13 firings per tide cycle each — useful, not in the way.
			if g, ok := pop(); ok && g < 16 {
				switch op {
				case 252:
					st.doRollback()
				case 253:
					st.doLoadOld(g)
				case 254:
					st.doColdRestart()
				default: // 255: inject; depth byte popped only on gate-pass
					if n, ok := pop(); ok {
						st.doInjectError(n)
					}
				}
			}
		}
	}
}

func runOpProgram(tb testing.TB, data []byte, cfg fuzzCfg) {
	tb.Helper()
	st := newFuzzState(tb, cfg)
	runOpChunk(st, data)
	st.releaseHolds(true)
	if st.latest > 0 {
		st.fullOracle()
	}
}

// --- entry point 1: native fuzzing ---

func FuzzTreeOps(f *testing.F) {
	// Seed corpus: known-nasty shapes (opcode bytes per the decoder above).
	seed := func(ops ...byte) { f.Add(ops) }
	const (
		opSet, opSetSame, opRemove, opNetZero = 0, 80, 100, 140
		opSave, opPrune, opGrow, opDrain      = 150, 175, 190, 200
		opRollback, opHold, opSnap, opIter    = 204, 210, 220, 228
		opLoadOld, opCold, opImport, opInject = 234, 242, 246, 250
	)
	// Net-zero twin then prune.
	seed(opGrow, 0, 0, 12, opSave, opNetZero, 1, opSave, opSet, 0, 50, opSave, opPrune, 4)
	// Same-value rewrite twin then prune.
	seed(opGrow, 0, 0, 12, opSave, opSetSame, 0, 3, opSave, opSet, 0, 60, opSave, opPrune, 4)
	// Drain → empty saves → prune through empties → regrow.
	seed(opGrow, 0, 0, 20, opSave, opDrain, opSave, opSave, opSave, opPrune, 4, opGrow, 0, 30, 10, opSave)
	// Repeated width-1 prunes through churn.
	seed(opSet, 0, 1, opSave, opSet, 0, 2, opSave, opPrune, 3, opSet, 0, 3, opSave, opPrune, 3, opSave, opPrune, 3)
	// Import then prune.
	seed(opGrow, 0, 0, 20, opSave, opSet, 0, 5, opSave, opImport, 1, opPrune, 4)
	// LoadOld: covering prune, idempotent save, recovery.
	seed(opSet, 0, 1, opSave, opSet, 0, 2, opSave, opSet, 0, 3, opSave, opLoadOld, 0, opPrune, 4)
	// Injected error mid-prune then retry (depth 0: fires on the first armed read).
	seed(opGrow, 0, 0, 20, opSave, opSet, 0, 7, opSave, opSet, 0, 8, opSave, opInject, 0)
	// Held snapshot blocks then releases.
	seed(opSet, 0, 1, opSave, opSet, 0, 2, opSave, opHold, 0, opPrune, 3, opSnap, 0, opPrune, 3)
	// Separator-shift deletes (first keys) then prune.
	seed(opGrow, 0, 0, 24, opSave, opRemove, 0, 0, opRemove, 0, 1, opSave, opSet, 0, 90, opSave, opPrune, 4)
	// Cold restart mid-churn.
	seed(opGrow, 0, 0, 16, opSave, opSet, 0, 4, opSave, opCold, opSet, 0, 5, opSave, opPrune, 4)
	// Import run crossing the 4×W gate: wide catch-up prune mid-run, then save.
	importRun := []byte{opGrow, 0, 0, 12, opSave}
	for i := 0; i < 17; i++ {
		importRun = append(importRun, opImport, byte(i%2))
	}
	seed(append(importRun, opSave, opPrune, 4)...)

	f.Fuzz(func(t *testing.T, data []byte) {
		runOpProgram(t, data, defaultFuzzCfg())
	})
}

// --- entry point 2: env-gated continuous soak ---

func TestSoak_TreeOps(t *testing.T) {
	spec := os.Getenv("BPTREE_SOAK")
	if spec == "" {
		t.Skip("set BPTREE_SOAK=<duration|forever> to run the soak (and -timeout=0 for long runs)")
	}
	var deadline time.Time
	if spec != "forever" {
		d, err := time.ParseDuration(spec)
		if err != nil {
			t.Fatalf("BPTREE_SOAK=%q: %v", spec, err)
		}
		deadline = time.Now().Add(d)
	}
	seed := int64(1)
	if s := os.Getenv("BPTREE_SOAK_SEED"); s != "" {
		fmt.Sscanf(s, "%d", &seed)
	}
	t.Logf("soak: seed=%d duration=%s", seed, spec)

	cfg := defaultFuzzCfg()
	cfg.allowImport = false // M21: repeated imports leak unboundedly
	cfg.maxOps = 1 << 30    // the chunk loop, not the op cap, bounds each pass
	cfg.keys = 1 << 16      // full 2-byte selector range; tideHigh must stay below ~0.93×keys (drift equilibrium) or the rise stalls
	cfg.tideHigh = 50_000   // crest at height-4 trees (non-root inners with inner children)
	cfg.sessionCap = 1024   // the session backbone: most sessions commit at ~1024 mutations
	st := newFuzzState(t, cfg)
	st.rising = true
	rng := rand.New(rand.NewSource(seed))
	chunk := make([]byte, 4096)
	// peak/low are per-log-window (chunk-boundary samples) so the line stays
	// diagnostic on long runs; the exact floor (0) is what flips the phase.
	chunks, peakKeys, peakLevels := 0, 0, 0
	lowKeys := cfg.tideHigh
	for spec == "forever" || time.Now().Before(deadline) {
		rng.Read(chunk)
		runTideChunk(st, chunk)
		chunks++
		peakKeys = max(peakKeys, len(st.model))
		lowKeys = min(lowKeys, len(st.model))
		levels := 0
		if st.tree.root != nil {
			levels = int(nodeHeight(st.tree.root)) + 1
		}
		peakLevels = max(peakLevels, levels)
		if chunks%16 == 0 {
			phase := "rising"
			if !st.rising {
				phase = "falling"
			}
			t.Logf("soak: %d chunks, %d ops, versions [%d,%d], %d live keys, %d levels, %s (window peak %d / low %d keys, max %d levels)",
				chunks, st.opN, st.first, st.latest, len(st.model), levels, phase, peakKeys, lowKeys, peakLevels)
			peakKeys, lowKeys = 0, cfg.tideHigh
		}
	}
	st.releaseHolds(true)
	if st.latest > 0 {
		st.fullOracle()
	}
	t.Logf("soak done: %d chunks, %d ops", chunks, st.opN)
	t.Logf("soak stats: %+v", st.stats)
}

// --- entry point 3: seeded -race stress on the sanctioned concurrent surface ---

func TestStress_ConcurrentSanctionedReaders(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 512, NewNopLogger())
	rng := rand.New(rand.NewSource(2))

	// Commit a baseline so readers always have versions to read.
	for i := 0; i < 200; i++ {
		tree.Set([]byte(fmt.Sprintf("cs%04d", i)), []byte("v0"))
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	var latest atomic.Int64
	latest.Store(1)

	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() { // single writer: mutate, save, prune (per the contract)
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			for j := 0; j < 20; j++ {
				tree.Set([]byte(fmt.Sprintf("cs%04d", rng.Intn(400))), []byte(fmt.Sprintf("w%d_%d", i, j)))
			}
			_, v, err := tree.SaveVersion()
			if err != nil {
				t.Error(err)
				return
			}
			latest.Store(v)
			if v > 4 {
				// Readers may hold registered snapshots — both outcomes valid.
				if err := tree.DeleteVersionsTo(v - 4); err != nil && !errors.Is(err, ErrActiveReaders) {
					t.Error(err)
					return
				}
			}
		}
	}()
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func(r int) {
			defer wg.Done()
			rrng := rand.New(rand.NewSource(int64(100 + r)))
			for i := 0; i < 1500; i++ {
				v := latest.Load() - int64(rrng.Intn(3))
				if v < 1 {
					v = 1
				}
				imm, err := tree.GetImmutable(v)
				if err != nil {
					continue // raced with a prune — sanctioned outcome
				}
				k := []byte(fmt.Sprintf("cs%04d", rrng.Intn(400)))
				if _, err := imm.Has(k); err != nil {
					t.Errorf("reader %d: Has on held v%d: %v", r, v, err)
				}
				if _, err := imm.Get(k); err != nil {
					t.Errorf("reader %d: Get on held v%d: %v", r, v, err)
				}
				if rrng.Intn(4) == 0 {
					if _, err := imm.GetMembershipProof(k); err != nil && !errors.Is(err, ErrKeyDoesNotExist) &&
						!strings.Contains(err.Error(), "key not found") {
						t.Errorf("reader %d: proof on held v%d: %v", r, v, err)
					}
				}
				if rrng.Intn(4) == 0 {
					it, err := imm.Iterator(nil, nil, true)
					if err != nil {
						t.Errorf("reader %d: iterator: %v", r, err)
					} else {
						for j := 0; j < 5 && it.Valid(); j++ {
							it.Next()
						}
						it.Close()
					}
				}
				_ = tree.VersionExists(v)
				imm.Close()
			}
		}(r)
	}
	// Let writer and readers overlap briefly, then stop the writer; readers
	// finish their fixed iteration counts against whatever versions remain.
	time.Sleep(50 * time.Millisecond)
	close(stop)
	wg.Wait()
}
