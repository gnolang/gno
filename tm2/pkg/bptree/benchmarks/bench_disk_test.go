package benchmarks

// Large-scale, disk-bound benchmark comparison (IAVL vs B+32) on a real disk
// DB (pebbledb by default; switch with -disk-backend).
//
// Unlike the warm benchmarks in bench_test.go (which build a small tree and
// then read it back from hot caches), this builds a fixture large enough that
// the working set dwarfs every cache layer — the in-process node LRU, pebble's
// 500MB block cache, and (at 100M keys / ~15-20GB per tree) the OS page cache.
// Random reads and block commits therefore exercise the real on-disk paths
// without any artificial cache-dropping.
//
// The fixture is built ONCE into a persistent pebbledb directory and reused
// across runs (resumable). Keys are derived deterministically from an integer
// index, so reads can pick a random *existing* key without storing all of them,
// and a partially-built fixture can be resumed.
//
// Realistic 100M comparison (needs ~40GB free at -disk-dir; build is one-time
// and can take a while):
//
//	go test ./tm2/pkg/bptree/benchmarks/ -run=^$ \
//	  -bench='BenchmarkDisk(GetRandom|GetMiss|BlockWrite)' \
//	  -disk-dir=/data/bptree-bench -disk-keys=100000000 \
//	  -benchtime=20000x -timeout=24h
//
// Swap the backend (default pebbledb) with -disk-backend; lmdbdb/mdbxdb need a
// cgo build (CGO_ENABLED=1):
//
//	go test ./tm2/pkg/bptree/benchmarks/ -run=TestDiskPopulate -v \
//	  -disk-dir=/data/pop -disk-keys=10000000 -disk-backend=lmdbdb -timeout=2h
//
// Quick smoke (default 1M keys, ephemeral temp dir):
//
//	go test ./tm2/pkg/bptree/benchmarks/ -run=^$ -bench='BenchmarkDisk'

import (
	"encoding/binary"
	"flag"
	"fmt"
	mrand "math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/boltdb"    // -disk-backend=boltdb
	_ "github.com/gnolang/gno/tm2/pkg/db/goleveldb" // -disk-backend=goleveldb
	"github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

var (
	diskDir          = flag.String("disk-dir", "", "persistent dir for disk fixtures; empty = ephemeral TempDir (fixture rebuilt each run)")
	diskKeys         = flag.Int64("disk-keys", 1_000_000, "fixture size N in keys (set 100000000 for the realistic disk-bound comparison)")
	diskBlock        = flag.Int("disk-block", 1000, "writes per block (SaveVersion cadence) for the block-write benchmark")
	diskNodeCache    = flag.Int("disk-node-cache", 10000, "in-process node LRU cache size, in nodes (production-realistic)")
	diskUpdateFrac   = flag.Float64("disk-update-frac", 0.5, "fraction of block writes that update existing keys (rest insert new keys)")
	diskBuildBatch   = flag.Int64("disk-build-batch", 100_000, "keys per SaveVersion while building the fixture")
	diskReloadEvery  = flag.Int("disk-reload-every", 100_000, "reload latest every N read ops to bound resident memory (the node LRU stays warm across reloads)")
	diskFactory      = flag.String("disk-factory", "", "limit disk populate/benchmarks to one backend: iavl|bptree (empty = both). Lets two processes populate in parallel into one -disk-dir.")
	diskVerbose      = flag.Bool("disk-verbose", false, "stream live populate progress to stderr: keys/sec + time split across set/save/prune/reload")
	diskVerboseEvery = flag.Duration("disk-verbose-every", time.Minute, "reporting interval for -disk-verbose")
	diskBackend      = flag.String("disk-backend", "pebbledb", "db backend for fixtures: pebbledb (tuned 500MB cache+bloom), goleveldb, boltdb; lmdbdb/mdbxdb need a cgo build")
)

// openDiskDB opens fixture sub-DB `name` under dir, honoring -disk-backend.
// pebbledb is opened with the production-tuned options (500MB block cache +
// bloom filter) the disk benchmarks are calibrated against — the generic
// registry's pebbledb creator uses empty options, which would silently drop
// both. Every other backend goes through dbm.NewDB and so must be linked in:
// goleveldb/boltdb always are; lmdbdb/mdbxdb only in a cgo build (see
// backends_cgo_test.go).
func openDiskDB(name, dir string) (dbm.DB, error) {
	if dbm.BackendType(*diskBackend) == dbm.PebbleDBBackend {
		return pebbledb.NewPebbleDBWithOpts(name, dir, pebbledb.DefaultPebbleOptions())
	}
	return dbm.NewDB(name, dbm.BackendType(*diskBackend), dir)
}

// selectedFactories returns the factories to run, filtered by -disk-factory
// (empty = all). Two processes with -disk-factory=iavl and -disk-factory=bptree
// can populate the same -disk-dir in parallel: distinct sub-DBs, no lock conflict.
func selectedFactories() []treeFactory {
	if *diskFactory == "" {
		return factories
	}
	for _, f := range factories {
		if f.name == *diskFactory {
			return []treeFactory{f}
		}
	}
	panic(fmt.Sprintf("unknown -disk-factory %q (want iavl|bptree)", *diskFactory))
}

const (
	diskKeyLen = 16
	diskValLen = 40
)

// mix64 is splitmix64 — a fast, deterministic bijection on uint64. Being a
// bijection guarantees distinct inputs map to distinct outputs, so the "hit"
// keyspace (input = i) and the "miss" keyspace (input = i with the top bit set)
// never collide.
func mix64(z uint64) uint64 {
	z += 0x9E3779B97F4A7C15
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// putDiskKey writes the 16-byte key for fixture index i into buf (no alloc).
// Both halves come from bijective mixes of i, so the full key is a bijection of
// i — distinct indices yield distinct keys.
func putDiskKey(buf []byte, i uint64) {
	a := mix64(i)
	b := mix64(a)
	binary.BigEndian.PutUint64(buf[0:8], a)
	binary.BigEndian.PutUint64(buf[8:16], b)
}

// putDiskMissKey writes a key guaranteed NOT to be in the fixture: it uses the
// top input bit, which fixture indices (< 2^40 in practice) never set.
func putDiskMissKey(buf []byte, i uint64) {
	putDiskKey(buf, i|(1<<63))
}

// putDiskVal writes a deterministic 40-byte value into buf (content is
// irrelevant to tree timing; the tree hashes it regardless).
func putDiskVal(buf []byte, i uint64) {
	z := i
	for off := 0; off < len(buf); off += 8 {
		z = mix64(z)
		var t [8]byte
		binary.BigEndian.PutUint64(t[:], z)
		copy(buf[off:], t[:])
	}
}

type diskFixture struct {
	tree  TreeBench
	db    dbm.DB // retained for block-cache (disk-read) stats; see diskBlockReads
	n     uint64
	close func()
}

// blockCacheStatter is implemented by pebbledb: cumulative block-cache hits and
// misses, where a miss ≈ one filesystem (disk) block read. Other backends don't
// expose it, so the disk-read metrics below are pebbledb-only.
type blockCacheStatter interface {
	BlockCacheStats() (hits, misses int64)
}

// diskBlockReads reads the backend's cumulative block-cache (hits, misses).
// ok is false for backends that don't expose it. misses ≈ physical disk reads;
// hits+misses ≈ logical block accesses (≈ traversal depth).
func diskBlockReads(db dbm.DB) (hits, misses int64, ok bool) {
	if s, isStatter := db.(blockCacheStatter); isStatter {
		h, m := s.BlockCacheStats()
		return h, m, true
	}
	return 0, 0, false
}

// readMeter accumulates block-cache deltas across timed segments so a benchmark
// can report disk reads (misses) and logical block accesses (hits+misses) per
// op. snap() opens a fresh segment — call it around untimed work (prune/reload)
// to exclude those reads; fold() accumulates the delta since the last snap/fold
// and advances the baseline. All methods no-op when the backend lacks stats.
type readMeter struct {
	db               dbm.DB
	ok               bool
	h0, m0           int64 // baseline at the current segment's start
	misses, accesses int64 // accumulated over folded segments
}

func newReadMeter(db dbm.DB) *readMeter {
	rm := &readMeter{db: db}
	rm.h0, rm.m0, rm.ok = diskBlockReads(db)
	return rm
}

func (rm *readMeter) snap() {
	if rm.ok {
		rm.h0, rm.m0, _ = diskBlockReads(rm.db)
	}
}

func (rm *readMeter) fold() {
	if !rm.ok {
		return
	}
	h, m, _ := diskBlockReads(rm.db)
	rm.misses += m - rm.m0
	rm.accesses += (h + m) - (rm.h0 + rm.m0)
	rm.h0, rm.m0 = h, m
}

// report emits diskreads/<unit> (misses ≈ physical reads) and blocks/<unit>
// (hits+misses ≈ logical traversal depth). No-op without stats or denom.
func (rm *readMeter) report(b *testing.B, denom float64, missMetric, accessMetric string) {
	if rm.ok && denom > 0 {
		b.ReportMetric(float64(rm.misses)/denom, missMetric)
		b.ReportMetric(float64(rm.accesses)/denom, accessMetric)
	}
}

// humanCount formats a key count for benchmark sub-names (20000->"20k", 1000000->"1M").
func humanCount(n uint64) string {
	switch {
	case n >= 1_000_000 && n%1_000_000 == 0:
		return fmt.Sprintf("%dM", n/1_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1e6)
	case n >= 1_000:
		return fmt.Sprintf("%dk", n/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// buildDiskFixture inserts keys [from, to) into tree using deterministic
// index->key/value derivation, committing every `batch` keys. It reloads latest
// between batches so resident memory stays bounded by the node LRU instead of
// materializing the whole tree (required for 100M-scale builds). Fresh slices
// per Set: IAVL retains the key slice by reference (bptree copies internally),
// so a reused buffer would alias every insert to a single key.
func buildDiskFixture(tb testing.TB, tree TreeBench, from, to, batch uint64, label string, logProgress bool) {
	tb.Helper()

	// Phase timers (accumulated across batches). Cheap: 4 time.Now per batch.
	var tSet, tSave, tPrune, tReload time.Duration
	start := time.Now()
	lastReport, lastKeys := start, from

	// report streams a live line to stderr (unbuffered, so it shows during the
	// run regardless of go test log buffering): rate + time split by phase.
	report := func(done uint64) {
		now := time.Now()
		win := now.Sub(lastReport).Seconds()
		rate := 0.0
		if win > 0 {
			rate = float64(done-lastKeys) / win
		}
		overall := float64(done-from) / now.Sub(start).Seconds()
		busy := tSet + tSave + tPrune + tReload
		pct := func(d time.Duration) float64 {
			if busy == 0 {
				return 0
			}
			return 100 * float64(d) / float64(busy)
		}
		fmt.Fprintf(os.Stderr,
			"[populate %-6s] %d/%d | %.0f keys/s (%.0fs win), %.0f overall | elapsed %s | "+
				"set %s/%.0f%% save %s/%.0f%% prune %s/%.0f%% reload %s/%.0f%%\n",
			label, done, to, rate, win, overall, now.Sub(start).Round(time.Second),
			tSet.Round(time.Millisecond), pct(tSet),
			tSave.Round(time.Millisecond), pct(tSave),
			tPrune.Round(time.Millisecond), pct(tPrune),
			tReload.Round(time.Millisecond), pct(tReload))
		lastReport, lastKeys = now, done
	}

	for i := from; i < to; {
		end := i + batch
		if end > to {
			end = to
		}
		t0 := time.Now()
		for ; i < end; i++ {
			k := make([]byte, diskKeyLen)
			v := make([]byte, diskValLen)
			putDiskKey(k, i)
			putDiskVal(v, i)
			if _, err := tree.Set(k, v); err != nil {
				tb.Fatalf("%s build Set: %v", label, err)
			}
		}
		tSet += time.Since(t0)

		t0 = time.Now()
		_, ver, err := tree.SaveVersion()
		if err != nil {
			tb.Fatalf("%s build SaveVersion: %v", label, err)
		}
		tSave += time.Since(t0)

		if ver > historySize {
			t0 = time.Now()
			if err := tree.DeleteVersionsTo(ver - historySize); err != nil {
				tb.Fatalf("%s build prune: %v", label, err)
			}
			tPrune += time.Since(t0)
		}

		t0 = time.Now()
		if _, err := tree.Load(); err != nil { // drop in-mem tree; node LRU stays warm
			tb.Fatalf("%s build reload: %v", label, err)
		}
		tReload += time.Since(t0)

		switch {
		case *diskVerbose && time.Since(lastReport) >= *diskVerboseEvery:
			report(i)
		case !*diskVerbose && logProgress && (i%(batch*10) == 0 || i == to):
			tb.Logf("  %s: %d/%d keys", label, i, to)
		}
	}
	if *diskVerbose {
		report(to) // final summary line
	}
}

// ensureDiskFixture opens (or creates) a per-factory pebbledb fixture and builds
// it to n keys, resuming if it already has some. All build work happens here,
// OUTSIDE any b.N loop, so it is never timed and never rebuilt during calibration.
// The build commits in batches and reloads latest between batches so resident
// memory stays bounded by the node LRU instead of materializing the whole tree.
func ensureDiskFixture(b *testing.B, f treeFactory, n uint64) diskFixture {
	b.Helper()
	dir := *diskDir
	ephemeral := dir == ""
	if ephemeral {
		dir = b.TempDir()
	} else {
		require.NoError(b, os.MkdirAll(dir, 0o755))
	}
	// Distinct sub-DB per factory so iavl and bptree don't share a directory.
	name := fmt.Sprintf("%s-disk", f.name)
	pdb, err := openDiskDB(name, dir)
	require.NoError(b, err)

	tree := f.newTree(pdb, *diskNodeCache)
	if _, err := tree.Load(); err != nil {
		b.Fatalf("load %s fixture: %v", f.name, err)
	}
	have := uint64(tree.Size())

	closeFn := func() { tree.Close(); pdb.Close() }

	if have < n {
		if !ephemeral {
			b.Logf("building %s fixture in %s: %d -> %d keys (one-time)...", f.name, dir, have, n)
		}
		buildDiskFixture(b, tree, have, n, uint64(*diskBuildBatch), f.name, !ephemeral)
	}
	if got := uint64(tree.Size()); got < n {
		closeFn()
		b.Fatalf("%s fixture size %d < requested %d — fixture build/persistence is broken", f.name, got, n)
	}
	return diskFixture{tree: tree, db: pdb, n: n, close: closeFn}
}

// BenchmarkDiskGetRandom measures random point reads of existing keys against a
// large on-disk fixture. Each op derives a fresh random existing key, so reads
// are genuinely scattered across the whole keyspace (not a small repeating pool
// that would warm into cache).
func BenchmarkDiskGetRandom(b *testing.B) {
	n := uint64(*diskKeys)
	for _, f := range selectedFactories() {
		fx := ensureDiskFixture(b, f, n)
		b.Run(fmt.Sprintf("%s/%s", f.name, humanCount(n)), func(b *testing.B) {
			b.ReportAllocs()
			b.ReportMetric(float64(fx.tree.Height()), "height")
			rng := mrand.New(mrand.NewSource(1))
			var key [diskKeyLen]byte
			// diskreads/op ≈ physical reads per Get (IAVL+fast ~1, bp32 ~2);
			// blocks/op ≈ logical traversal depth. Fold around the untimed
			// reload so its reads aren't counted.
			rm := newReadMeter(fx.db)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i > 0 && *diskReloadEvery > 0 && i%*diskReloadEvery == 0 {
					b.StopTimer()
					rm.fold()             // close segment before the untimed reload
					_, _ = fx.tree.Load() // bound memory; node LRU stays warm
					rm.snap()             // reopen segment after reload
					b.StartTimer()
				}
				putDiskKey(key[:], uint64(rng.Int63n(int64(n))))
				if _, err := fx.tree.Get(key[:]); err != nil {
					b.Fatalf("Get: %v", err)
				}
			}
			rm.fold() // final segment
			rm.report(b, float64(b.N), "diskreads/op", "blocks/op")
		})
		fx.close()
	}
}

// BenchmarkDiskGetMiss measures random point reads of absent keys (exercises the
// bloom-filter / negative-lookup path, where B+32 rejects in-memory and IAVL
// must consult disk).
func BenchmarkDiskGetMiss(b *testing.B) {
	n := uint64(*diskKeys)
	for _, f := range selectedFactories() {
		fx := ensureDiskFixture(b, f, n)
		b.Run(fmt.Sprintf("%s/%s", f.name, humanCount(n)), func(b *testing.B) {
			b.ReportAllocs()
			rng := mrand.New(mrand.NewSource(3))
			var key [diskKeyLen]byte
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i > 0 && *diskReloadEvery > 0 && i%*diskReloadEvery == 0 {
					b.StopTimer()
					_, _ = fx.tree.Load()
					b.StartTimer()
				}
				putDiskMissKey(key[:], uint64(rng.Int63n(int64(n))))
				if _, err := fx.tree.Get(key[:]); err != nil {
					b.Fatalf("Get: %v", err)
				}
			}
		})
		fx.close()
	}
}

// BenchmarkDiskBlockWrite measures the cost of committing a block: -disk-block
// writes (a configurable mix of updates to existing keys and new inserts)
// followed by SaveVersion, against the large on-disk fixture. ns/op is the
// per-block latency; ns/write is also reported. On pebbledb it additionally
// reports diskreads/write (block-cache misses ≈ physical reads per write) and
// blocks/write (hits+misses ≈ logical traversal depth) — the empirical basis
// for the write-depth gas params, where bp32's shallow tree should show far
// fewer reads than IAVL's deep COW path. Pruning and the drop-in-memory-tree
// reload happen outside the timer (a real node prunes out-of-band and starts
// each block from committed state, lazily loading what its txs touch — which
// the timed Set path models).
func BenchmarkDiskBlockWrite(b *testing.B) {
	n := uint64(*diskKeys)
	bs := *diskBlock
	for _, f := range selectedFactories() {
		fx := ensureDiskFixture(b, f, n)
		b.Run(fmt.Sprintf("%s/%s/block-%d", f.name, humanCount(n), bs), func(b *testing.B) {
			b.ReportAllocs()
			rng := mrand.New(mrand.NewSource(2))
			next := uint64(fx.tree.Size()) // fresh-insert index, past all existing keys
			// diskreads/write (misses ≈ physical reads) and blocks/write
			// (hits+misses ≈ logical depth) over the TIMED Set+SaveVersion
			// region; snap()/fold() bracket it so the untimed prune/reload
			// between blocks are excluded.
			rm := newReadMeter(fx.db)
			b.ResetTimer()
			for i := 0; i < b.N; i++ { // one iteration == one block
				rm.snap()
				for j := 0; j < bs; j++ {
					k := make([]byte, diskKeyLen) // fresh per Set (IAVL retains key ref)
					v := make([]byte, diskValLen)
					if rng.Float64() < *diskUpdateFrac {
						putDiskKey(k, uint64(rng.Int63n(int64(n)))) // update existing
					} else {
						putDiskKey(k, next) // insert new
						next++
					}
					putDiskVal(v, next+uint64(j))
					if _, err := fx.tree.Set(k, v); err != nil {
						b.Fatalf("Set: %v", err)
					}
				}
				_, ver, err := fx.tree.SaveVersion()
				if err != nil {
					b.Fatalf("SaveVersion: %v", err)
				}
				rm.fold()
				b.StopTimer()
				if ver > historySize {
					if err := fx.tree.DeleteVersionsTo(ver - historySize); err != nil {
						b.Fatalf("prune: %v", err)
					}
				}
				if _, err := fx.tree.Load(); err != nil { // drop in-mem tree; LRU stays warm
					b.Fatalf("reload: %v", err)
				}
				b.StartTimer()
			}
			b.ReportMetric(float64(bs), "writes/block")
			if b.N > 0 {
				w := float64(b.N * bs)
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/w, "ns/write")
				rm.report(b, w, "diskreads/write", "blocks/write")
			}
		})
		fx.close()
	}
}

// TestDiskPopulate measures wall-clock time to populate each tree backend
// (iavl, bptree) to -disk-keys from empty, separately, into its own fresh
// pebbledb directory. Gated on -disk-dir so it never runs during a normal
// `go test`. Example:
//
//	go test ./tm2/pkg/bptree/benchmarks/ -run=TestDiskPopulate -v \
//	  -disk-dir=/data/pop -disk-keys=10000000 -timeout=2h
func TestDiskPopulate(t *testing.T) {
	if *diskDir == "" {
		t.Skip("set -disk-dir (and -disk-keys) to run the disk populate")
	}
	n := uint64(*diskKeys)
	require.NoError(t, os.MkdirAll(*diskDir, 0o755))
	for _, f := range selectedFactories() {
		// Build into the exact path the disk benchmarks reuse (<dir>/<name>.db),
		// so no rename is needed afterward. Resumable: if already at >= n keys,
		// skip; otherwise continue from the current size.
		name := fmt.Sprintf("%s-disk", f.name)
		pdb, err := openDiskDB(name, *diskDir)
		require.NoError(t, err)
		tree := f.newTree(pdb, *diskNodeCache)
		if _, err := tree.Load(); err != nil {
			t.Fatalf("%s load: %v", f.name, err)
		}
		have := uint64(tree.Size())
		if have < n {
			start := time.Now()
			buildDiskFixture(t, tree, have, n, uint64(*diskBuildBatch), f.name, true)
			elapsed := time.Since(start)
			t.Logf(">>> POPULATE %-6s: %d -> %d keys in %s (%.0f keys/sec)",
				f.name, have, n, elapsed.Round(time.Millisecond), float64(n-have)/elapsed.Seconds())
		} else {
			t.Logf(">>> %-6s already populated (size=%d), skipping", f.name, have)
		}
		size := tree.Size()
		tree.Close()
		pdb.Close()
		mb := dirSizeMB(filepath.Join(*diskDir, name+".db"))
		t.Logf(">>> %-6s: size=%d, disk=%.0f MB (%.0f B/key)", f.name, size, mb, mb*1024*1024/float64(n))
	}
}
