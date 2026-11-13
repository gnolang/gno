package bptree

// A/B harness comparing the in-tree dual-tree-walk prune against the
// mark-and-sweep prune proposed in PR #5571 (originally #5570).
//
// The question this answers: mark-and-sweep marks every node reachable from
// the first RETAINED version before it deletes anything, so its cost is
// O(retained tree) per PruneVersionsTo CALL. The dual-tree-walk only descends
// where two adjacent versions differ, so its cost is O(divergence) per version
// pruned. Which one wins depends entirely on the ratio
//
//	(versions pruned per call) × (divergence per version)   vs   (tree size)
//
// gno.land prunes ONE version per block (store/bptree.Store.Commit calls
// DeleteVersionsTo(version-1-KeepRecent), which advances by one each block),
// with a divergence of one block's writes against a tree of millions of keys.
// That is the far left of the ratio. The benchmarks below parameterise both
// axes so the crossover can be located rather than assumed.
//
// This file is test-only: the mark-and-sweep implementation here is a faithful
// port of the PR's algorithm onto the current nodeDB API, kept next to the
// benchmark that judges it.

import (
	"encoding/binary"
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

var (
	abBackend = flag.String("abbackend", "memdb", "backend for the prune A/B benchmarks: memdb | pebbledb")
	// The interesting sizes are the big ones, but a 1M-key memdb fixture
	// takes ~20 s and a few GB to build, so the default stays small and the
	// scaling runs are opt-in: -absizes=1000000,3000000.
	abSizes = flag.String("absizes", "10000,100000", "comma-separated tree sizes for the prune A/B benchmarks")
	abCache = flag.Int("abcache", 10_000, "node LRU size for the prune A/B benchmarks (production default: 10000)")
)

func abSizeList(tb testing.TB) []int {
	tb.Helper()
	var out []int
	for s := range strings.SplitSeq(*abSizes, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			tb.Fatalf("bad -absizes entry %q: %v", s, err)
		}
		out = append(out, n)
	}
	return out
}

func abMakeDB(tb testing.TB) dbm.DB {
	tb.Helper()
	switch *abBackend {
	case "memdb":
		return memdb.NewMemDB()
	case "pebbledb":
		db, err := pebbledb.NewPebbleDBWithOpts("ab", tb.TempDir(), pebbledb.DefaultPebbleOptions())
		if err != nil {
			tb.Fatal(err)
		}
		return db
	default:
		tb.Fatalf("unknown -abbackend %q", *abBackend)
		return nil
	}
}

// -----------------------------------------------------------------------
// counting DB — backend-agnostic measure of the work a prune issues
// -----------------------------------------------------------------------

type abCountingDB struct {
	dbm.DB
	reads  atomic.Int64
	writes atomic.Int64
}

func (c *abCountingDB) Get(key []byte) ([]byte, error) {
	c.reads.Add(1)
	return c.DB.Get(key)
}

func (c *abCountingDB) Has(key []byte) (bool, error) {
	c.reads.Add(1)
	return c.DB.Has(key)
}

func (c *abCountingDB) Set(key, value []byte) error {
	c.writes.Add(1)
	return c.DB.Set(key, value)
}

func (c *abCountingDB) SetSync(key, value []byte) error {
	c.writes.Add(1)
	return c.DB.SetSync(key, value)
}

func (c *abCountingDB) Delete(key []byte) error {
	c.writes.Add(1)
	return c.DB.Delete(key)
}

func (c *abCountingDB) DeleteSync(key []byte) error {
	c.writes.Add(1)
	return c.DB.DeleteSync(key)
}

func (c *abCountingDB) NewBatch() dbm.Batch {
	return &abCountingBatch{Batch: c.DB.NewBatch(), writes: &c.writes}
}

func (c *abCountingDB) NewBatchWithSize(n int) dbm.Batch {
	return &abCountingBatch{Batch: c.DB.NewBatchWithSize(n), writes: &c.writes}
}

type abCountingBatch struct {
	dbm.Batch
	writes *atomic.Int64
}

func (b *abCountingBatch) Set(key, value []byte) error {
	b.writes.Add(1)
	return b.Batch.Set(key, value)
}

func (b *abCountingBatch) Delete(key []byte) error {
	b.writes.Add(1)
	return b.Batch.Delete(key)
}

// -----------------------------------------------------------------------
// mark-and-sweep prune (PR #5571 algorithm, ported to the current API)
// -----------------------------------------------------------------------

func nodeKeyArr(nk *NodeKey) [NodeKeySize]byte {
	var a [NodeKeySize]byte
	binary.BigEndian.PutUint64(a[:8], uint64(nk.Version))
	binary.BigEndian.PutUint32(a[8:], nk.Nonce)
	return a
}

func nodeKeyBytesToArr(b []byte) ([NodeKeySize]byte, error) {
	var a [NodeKeySize]byte
	if len(b) != NodeKeySize {
		return a, fmt.Errorf("short serialized NodeKey ref: len=%d want %d", len(b), NodeKeySize)
	}
	copy(a[:], b)
	return a, nil
}

// pruneVersionsToMarkSweep mirrors PruneVersionsTo's guards and then runs the
// mark-and-sweep range instead of the dual-tree-walk range.
//
// One deliberate difference: it has no flush-threshold commits, so the batch
// it builds is unbounded. That favours mark-and-sweep in the batch benchmark
// (it skips the intermediate Commits the production path pays), which only
// strengthens the conclusion that it loses there anyway.
func pruneVersionsToMarkSweep(t *MutableTree, toVersion int64) error {
	first := t.ndb.getFirstVersion()
	latest := t.ndb.getLatestVersion()
	if toVersion >= latest {
		return fmt.Errorf("cannot prune latest version %d", latest)
	}
	if toVersion < first {
		return nil
	}
	if t.root != t.lastSaved || len(t.ndb.pendingVals) > 0 ||
		t.nextValueNonce > 0 || len(t.versionOrphans) > 0 {
		return fmt.Errorf("%w: SaveVersion or Rollback before pruning", ErrUncommittedChanges)
	}
	if t.version <= toVersion {
		return fmt.Errorf("%w: version %d is loaded by the working tree", ErrActiveReaders, t.version)
	}
	if err := t.ndb.beginPruning(first, toVersion); err != nil {
		return err
	}
	defer t.ndb.endPruning()

	if err := pruneRangeMarkSweep(t, first, toVersion, latest); err != nil {
		t.ndb.DiscardBatch()
		return err
	}
	t.ndb.setFirstVersion(toVersion + 1)
	return nil
}

func pruneRangeMarkSweep(t *MutableTree, first, toVersion, latest int64) error {
	// Mark phase: one walk of the first retained version's tree records every
	// live NodeKey. Sound by the COW invariant — a node saved at version W and
	// reachable from a retained version R is reachable from every existing
	// version in [W, R], so the earliest retained version's walk is a superset
	// of what the sweep can observe.
	reachable, err := buildRetainedReachableSet(t, toVersion+1, latest)
	if err != nil {
		return err
	}

	for v := first; v <= toVersion; v++ {
		exists, err := t.ndb.versionExistsE(v)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		nextV, err := t.findNextVersion(v, latest)
		if err != nil {
			return err
		}
		if nextV == 0 {
			return fmt.Errorf("bptree: pruning v%d found no successor", v)
		}
		if err := sweepAndOrphanVersion(t, v, nextV, reachable); err != nil {
			return err
		}
		if err := t.ndb.DeleteRoot(v); err != nil {
			return err
		}
	}
	return t.ndb.Commit()
}

// abMarkStats accumulates mark-phase instrumentation across a benchmark run.
var abMarkStats struct {
	calls     atomic.Int64
	setSize   atomic.Int64
	markReads atomic.Int64
}

func resetAbMarkStats() {
	abMarkStats.calls.Store(0)
	abMarkStats.setSize.Store(0)
	abMarkStats.markReads.Store(0)
}

func buildRetainedReachableSet(t *MutableTree, from, to int64) (map[[NodeKeySize]byte]struct{}, error) {
	reachable := make(map[[NodeKeySize]byte]struct{})
	if cdb, ok := t.ndb.db.(*abCountingDB); ok {
		before := cdb.reads.Load()
		defer func() {
			abMarkStats.calls.Add(1)
			abMarkStats.setSize.Add(int64(len(reachable)))
			abMarkStats.markReads.Add(cdb.reads.Load() - before)
		}()
	}
	for rv := from; rv <= to; rv++ {
		exists, err := t.ndb.versionExistsE(rv)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		rvRootNK, _, err := t.ndb.GetRoot(rv)
		if err != nil {
			return nil, err
		}
		if rvRootNK == nil {
			continue
		}
		rvRoot, err := t.ndb.GetNode(rvRootNK)
		if err != nil {
			return nil, err
		}
		if err := markReachable(t, rvRoot, reachable); err != nil {
			return nil, err
		}
		break // single mark pass suffices by the COW invariant
	}
	return reachable, nil
}

func markReachable(t *MutableTree, node Node, reachable map[[NodeKeySize]byte]struct{}) error {
	if node == nil {
		return nil
	}
	if nk := node.GetNodeKey(); nk != nil {
		key := nodeKeyArr(nk)
		if _, seen := reachable[key]; seen {
			return nil
		}
		reachable[key] = struct{}{}
	}
	inner, ok := node.(*InnerNode)
	if !ok {
		return nil
	}
	leafChildren := inner.height == 1
	for i := 0; i < inner.NumChildren(); i++ {
		if inner.childNodes[i] != nil {
			if err := markReachable(t, inner.childNodes[i], reachable); err != nil {
				return err
			}
			continue
		}
		if inner.children[i] == nil {
			continue
		}
		arr, err := nodeKeyBytesToArr(inner.children[i])
		if err != nil {
			return err
		}
		if _, seen := reachable[arr]; seen {
			continue
		}
		if leafChildren {
			reachable[arr] = struct{}{}
			continue
		}
		child, err := t.ndb.GetNode(inner.children[i])
		if err != nil {
			return err
		}
		if err := markReachable(t, child, reachable); err != nil {
			return err
		}
	}
	return nil
}

func sweepAndOrphanVersion(t *MutableTree, v, nextV int64, reachable map[[NodeKeySize]byte]struct{}) error {
	vRootNK, _, err := t.ndb.GetRoot(v)
	if err != nil {
		return err
	}
	if vRootNK != nil {
		vRoot, err := t.ndb.GetNode(vRootNK)
		if err != nil {
			return err
		}
		if err := sweepOld(t, vRoot, reachable); err != nil {
			return err
		}
	}
	for _, version := range [2]int64{v, nextV} {
		orphans, err := t.ndb.LoadOrphans(version)
		if err != nil {
			return err
		}
		for _, vk := range orphans {
			if err := t.ndb.DeleteValue(vk); err != nil {
				return err
			}
		}
	}
	if err := t.ndb.DeleteOrphans(nextV); err != nil {
		return err
	}
	return t.ndb.DeleteOrphans(v)
}

func sweepOld(t *MutableTree, node Node, reachable map[[NodeKeySize]byte]struct{}) error {
	if node == nil {
		return nil
	}
	if nk := node.GetNodeKey(); nk != nil {
		if _, shared := reachable[nodeKeyArr(nk)]; shared {
			return nil
		}
		// DeleteNodeKey, like the production walk, so the A/B is not
		// measuring an allocation difference in the delete path.
		if err := t.ndb.DeleteNodeKey(nk); err != nil {
			return err
		}
	}
	inner, ok := node.(*InnerNode)
	if !ok {
		return nil
	}
	leafChildren := inner.height == 1
	for i := 0; i < inner.NumChildren(); i++ {
		var child Node
		switch {
		case inner.childNodes[i] != nil:
			if ck := inner.childNodes[i].GetNodeKey(); ck != nil {
				if _, shared := reachable[nodeKeyArr(ck)]; shared {
					continue
				}
			}
			child = inner.childNodes[i]
		case inner.children[i] != nil:
			arr, err := nodeKeyBytesToArr(inner.children[i])
			if err != nil {
				return err
			}
			if _, shared := reachable[arr]; shared {
				continue
			}
			if leafChildren {
				if err := t.ndb.DeleteNode(inner.children[i]); err != nil {
					return err
				}
				continue
			}
			c, err := t.ndb.GetNode(inner.children[i])
			if err != nil {
				return err
			}
			child = c
		}
		if child == nil {
			continue
		}
		if err := sweepOld(t, child, reachable); err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------------------
// fixture
// -----------------------------------------------------------------------

type pruneFixture struct {
	tree *MutableTree
	db   *abCountingDB
	keys [][]byte
}

func abKey(r *rand.Rand) []byte {
	k := make([]byte, 16)
	r.Read(k)
	return k
}

// buildPruneFixture creates a tree of size keys at v1, then appends
// versions-1 more versions, each applying writesPerVersion key writes
// (half updates of existing keys, half fresh inserts) — the shape of a
// gno.land block.
func buildPruneFixture(tb testing.TB, size, versions, writesPerVersion, cacheSize int) *pruneFixture {
	tb.Helper()
	r := rand.New(rand.NewSource(1))
	cdb := &abCountingDB{DB: abMakeDB(tb)}
	tree := NewMutableTreeWithDB(cdb, cacheSize, NewNopLogger())

	keys := make([][]byte, 0, size+versions*writesPerVersion)
	val := make([]byte, 40)
	r.Read(val)
	for range size {
		k := abKey(r)
		if _, err := tree.Set(k, val); err != nil {
			tb.Fatal(err)
		}
		keys = append(keys, k)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		tb.Fatal(err)
	}
	for range versions - 1 {
		for j := range writesPerVersion {
			if j%2 == 0 {
				k := abKey(r)
				if _, err := tree.Set(k, val); err != nil {
					tb.Fatal(err)
				}
				keys = append(keys, k)
			} else {
				k := keys[r.Intn(len(keys))]
				if _, err := tree.Set(k, val); err != nil {
					tb.Fatal(err)
				}
			}
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			tb.Fatal(err)
		}
	}
	return &pruneFixture{tree: tree, db: cdb, keys: keys}
}

// -----------------------------------------------------------------------
// benchmarks
// -----------------------------------------------------------------------

type pruneAlgo struct {
	name string
	run  func(t *MutableTree, to int64) error
}

var pruneAlgos = []pruneAlgo{
	{"dualwalk", func(t *MutableTree, to int64) error { return t.PruneVersionsTo(to) }},
	{"marksweep", pruneVersionsToMarkSweep},
}

// BenchmarkPruneAB_OneVersion is the gno.land shape: one DeleteVersionsTo call
// retires exactly one version. Run with -benchtime=<N>x where N <= versions-2.
func BenchmarkPruneAB_OneVersion(b *testing.B) {
	for _, sz := range abSizeList(b) {
		for _, algo := range pruneAlgos {
			b.Run(fmt.Sprintf("%s/%dk", algo.name, sz/1000), func(b *testing.B) {
				fx := buildPruneFixture(b, sz, 64, 100, *abCache)
				defer fx.tree.Close()
				if b.N > 60 {
					b.Skipf("needs -benchtime<=60x (got N=%d)", b.N)
				}
				r0, w0 := fx.db.reads.Load(), fx.db.writes.Load()
				resetAbMarkStats()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := algo.run(fx.tree, int64(i+1)); err != nil {
						b.Fatal(err)
					}
				}
				b.StopTimer()
				reportPruneStats(b, fx, r0, w0)
			})
		}
	}
}

// BenchmarkPruneAB_BlockLoop is the whole gno.land block cycle: apply a
// block's writes, SaveVersion, then retire one version. It is the benchmark
// that matters, because it also prices what the per-version benchmarks hide —
// the mark walk evicts the block's working set from the shared node LRU, so
// the NEXT block's writes pay to reload it.
func BenchmarkPruneAB_BlockLoop(b *testing.B) {
	const writesPerBlock = 100
	for _, sz := range abSizeList(b) {
		for _, algo := range pruneAlgos {
			b.Run(fmt.Sprintf("%s/%dk", algo.name, sz/1000), func(b *testing.B) {
				fx := buildPruneFixture(b, sz, 4, writesPerBlock, *abCache)
				defer fx.tree.Close()
				r := rand.New(rand.NewSource(7))
				val := make([]byte, 40)
				r.Read(val)
				r0, w0 := fx.db.reads.Load(), fx.db.writes.Load()
				resetAbMarkStats()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for j := range writesPerBlock {
						if j%2 == 0 {
							k := abKey(r)
							if _, err := fx.tree.Set(k, val); err != nil {
								b.Fatal(err)
							}
							fx.keys = append(fx.keys, k)
						} else if _, err := fx.tree.Set(fx.keys[r.Intn(len(fx.keys))], val); err != nil {
							b.Fatal(err)
						}
					}
					if _, _, err := fx.tree.SaveVersion(); err != nil {
						b.Fatal(err)
					}
					// One version retired per block, exactly as
					// store/bptree.Store.Commit does.
					if err := algo.run(fx.tree, fx.tree.Version()-3); err != nil {
						b.Fatal(err)
					}
				}
				b.StopTimer()
				reportPruneStats(b, fx, r0, w0)
			})
		}
	}
}

// BenchmarkPruneAB_Batch retires many versions in ONE call — the catch-up
// shape (a node restarting with a large backlog, or a retention policy change).
// This is where a single mark pass amortises.
func BenchmarkPruneAB_Batch(b *testing.B) {
	const sz = 100_000
	for _, batch := range []int{1, 8, 32} {
		for _, algo := range pruneAlgos {
			b.Run(fmt.Sprintf("%s/batch-%d", algo.name, batch), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					fx := buildPruneFixture(b, sz, batch+2, 100, *abCache)
					r0, w0 := fx.db.reads.Load(), fx.db.writes.Load()
					resetAbMarkStats()
					b.StartTimer()
					if err := algo.run(fx.tree, int64(batch)); err != nil {
						b.Fatal(err)
					}
					b.StopTimer()
					reportPruneStats(b, fx, r0, w0)
					fx.tree.Close()
				}
			})
		}
	}
}

func reportPruneStats(b *testing.B, fx *pruneFixture, r0, w0 int64) {
	b.Helper()
	n := int64(b.N)
	if n == 0 {
		n = 1
	}
	b.ReportMetric(float64(fx.db.reads.Load()-r0)/float64(n), "dbreads/op")
	b.ReportMetric(float64(fx.db.writes.Load()-w0)/float64(n), "dbwrites/op")
	if calls := abMarkStats.calls.Load(); calls > 0 {
		b.ReportMetric(float64(abMarkStats.setSize.Load())/float64(calls), "markset")
		b.ReportMetric(float64(abMarkStats.markReads.Load())/float64(calls), "markreads")
	}
}
