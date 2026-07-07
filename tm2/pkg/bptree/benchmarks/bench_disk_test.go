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
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	bptree "github.com/gnolang/gno/tm2/pkg/bptree"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/boltdb"    // -disk-backend=boltdb
	_ "github.com/gnolang/gno/tm2/pkg/db/goleveldb" // -disk-backend=goleveldb
	"github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

var (
	diskDir           = flag.String("disk-dir", "", "persistent dir for disk fixtures; empty = ephemeral TempDir (fixture rebuilt each run)")
	diskKeys          = flag.Int64("disk-keys", 1_000_000, "fixture size N in keys (set 100000000 for the realistic disk-bound comparison)")
	diskBlock         = flag.Int("disk-block", 1000, "writes per block (SaveVersion cadence) for the block-write benchmark")
	diskNodeCache     = flag.Int("disk-node-cache", 10000, "in-process node LRU cache size, in nodes (production-realistic)")
	diskUpdateFrac    = flag.Float64("disk-update-frac", 0.5, "fraction of block writes that update existing keys (rest insert new keys)")
	diskBuildBatch    = flag.Int64("disk-build-batch", 25_000, "keys per SaveVersion while building the fixture")
	diskWarmupOps     = flag.Int("disk-warmup-ops", 0, "untimed ops before measurement to warm the node LRU (random gets for the Get benches; whole blocks for BlockWrite). A fresh tree starts cold and reported counts average over every iteration, so size this to several times the node cache: ~50000 for the default 10K cache, ~1500000 for a 330K cache")
	diskReloadEvery   = flag.Int("disk-reload-every", 100_000, "reload latest every N read ops to bound resident memory (the node LRU stays warm across reloads)")
	diskCommittedRead = flag.Bool("disk-committed-read", false, "read the Get/GetMiss benches through a committed snapshot held at the latest version (the ABCI-query path that the bptree fast index serves) instead of the working-tree Get, which is index-free")
	diskFactory       = flag.String("disk-factory", "", "limit disk populate/benchmarks to one backend: iavl|bptree|bptree-fast (empty = all). Lets processes populate in parallel into one -disk-dir. bptree-fast reuses the bptree fixture (run it after bptree, not concurrently).")
	diskVerbose       = flag.Bool("disk-verbose", false, "stream live populate progress to stderr: keys/sec + time split across set/save/prune/reload")
	diskVerboseEvery  = flag.Duration("disk-verbose-every", time.Minute, "reporting interval for -disk-verbose")
	diskBackend       = flag.String("disk-backend", "pebbledb", "db backend for fixtures: pebbledb (tuned 500MB cache+bloom), goleveldb, boltdb; lmdbdb/mdbxdb need a cgo build")
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
	panic(fmt.Sprintf("unknown -disk-factory %q (want iavl|bptree|bptree-fast)", *diskFactory))
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
	db    *countingDB // counts the tree's DB reads/writes; see readMeter
	n     uint64
	close func()
}

// readMeter accumulates the tree's DB-operation counts (node reads via Get,
// node writes via Set/Delete) across timed segments, so a benchmark can report
// reads and writes per op. Unlike pebble's global block-cache counter, these
// are reproducible, backend-agnostic, and unaffected by pebble's background
// compaction (which is below the dbm.DB interface). snap() opens a fresh
// segment — call it around untimed work (prune/reload) to exclude it; fold()
// accumulates the delta since the last snap/fold and advances the baseline.
//
// Caveat: reads are node-LRU *misses* (Get fires only on a cache miss), so the
// per-op counts depend on -disk-node-cache and the access pattern, not on tree
// shape alone. Read them as a fair *relative* iavl-vs-bp32 indicator at full
// fixture scale and high -benchtime (short runs under-report while the warm LRU
// still covers the working set), not as an absolute structural depth — to set a
// fixed gas param, calibrate against production cache size / a cold-cache run.
type segment struct {
	ops, reads, writes int64
	ns                 int64 // timed wall ns (b.Elapsed deltas; StopTimer spans excluded)
}

type readMeter struct {
	db            *countingDB
	r0, w0        int64 // baseline at the current segment's start
	prevNS        int64 // b.Elapsed() at the current segment's start
	reads, writes int64 // accumulated over folded segments
	segs          []segment
}

// newReadMeter must be created immediately before b.ResetTimer(): the elapsed
// baseline starts at 0, which ResetTimer makes true.
func newReadMeter(db *countingDB) *readMeter {
	rm := &readMeter{db: db}
	rm.r0, rm.w0 = db.stats()
	return rm
}

// snap opens a fresh segment; call after untimed work so it is excluded.
func (rm *readMeter) snap(b *testing.B) {
	b.Helper()
	rm.r0, rm.w0 = rm.db.stats()
	rm.prevNS = b.Elapsed().Nanoseconds()
}

// fold closes the current segment covering ops benchmark ops. b.Elapsed()
// freezes across StopTimer spans, so untimed reload/prune gaps between a
// fold and the following snap contribute nothing to segment wall time.
func (rm *readMeter) fold(b *testing.B, ops int64) {
	b.Helper()
	r, w := rm.db.stats()
	el := b.Elapsed().Nanoseconds()
	if ops > 0 {
		rm.segs = append(rm.segs, segment{ops: ops, reads: r - rm.r0, writes: w - rm.w0, ns: el - rm.prevNS})
	}
	rm.reads += r - rm.r0
	rm.writes += w - rm.w0
	rm.r0, rm.w0, rm.prevNS = r, w, el
}

// report emits reads and writes per denom (skipping a metric named "").
func (rm *readMeter) report(b *testing.B, denom float64, readMetric, writeMetric string) {
	b.Helper()
	if denom <= 0 {
		return
	}
	if readMetric != "" {
		b.ReportMetric(float64(rm.reads)/denom, readMetric)
	}
	if writeMetric != "" {
		b.ReportMetric(float64(rm.writes)/denom, writeMetric)
	}
}

const tailMaxWindows = 8

type window struct {
	ops, reads, writes int64
	medNSPerOp         float64 // median over the window's segments of seg.ns/seg.ops
}

// windows groups the folded segments into at most maxW consecutive windows of
// K = ceil(S/maxW) segments, grouped from the END so every window except
// possibly the first is full — the tail windows the convergence verdict reads
// are never runts. A short final segment (under half the previous one) is
// merged into its predecessor first. The median makes one compaction-stalled
// block unable to flip a write window; for single-segment windows it
// degenerates to the segment mean.
func (rm *readMeter) windows(maxW int) []window {
	segs := append([]segment(nil), rm.segs...)
	if n := len(segs); n >= 2 && segs[n-1].ops*2 < segs[n-2].ops {
		last := segs[n-1]
		segs[n-2].ops += last.ops
		segs[n-2].reads += last.reads
		segs[n-2].writes += last.writes
		segs[n-2].ns += last.ns
		segs = segs[:n-1]
	}
	if len(segs) == 0 {
		return nil
	}
	k := (len(segs) + maxW - 1) / maxW
	var ws []window
	for hi := len(segs); hi > 0; hi -= k {
		lo := max(hi-k, 0)
		var w window
		rates := make([]float64, 0, hi-lo)
		for _, sg := range segs[lo:hi] {
			w.ops += sg.ops
			w.reads += sg.reads
			w.writes += sg.writes
			rates = append(rates, float64(sg.ns)/float64(sg.ops))
		}
		slices.Sort(rates)
		if m := len(rates); m%2 == 1 {
			w.medNSPerOp = rates[m/2]
		} else {
			w.medNSPerOp = (rates[m/2-1] + rates[m/2]) / 2
		}
		ws = append([]window{w}, ws...)
	}
	return ws
}

// pctDelta returns 100*(cur-prev)/prev. Both-zero is 0 (converged at zero,
// e.g. bptree miss reads); zero-to-nonzero has no finite delta.
func pctDelta(cur, prev float64) (float64, bool) {
	if prev == 0 {
		if cur == 0 {
			return 0, true
		}
		return 0, false
	}
	return 100 * (cur - prev) / prev, true
}

// reportTail emits steady-state estimates from the FINAL window plus
// convergence deltas vs the penultimate window, and logs the full per-window
// series so a pasted run shows ramp-then-flat (or still-ramping) at a glance.
// Steady iff the conv-* deltas are small AND the logged tail isn't a monotone
// same-sign slide. With fewer than 2 windows only the window count is emitted:
// such a run carries no convergence evidence.
func (rm *readMeter) reportTail(b *testing.B, denom string, emitWrites bool) {
	b.Helper()
	ws := rm.windows(tailMaxWindows)
	b.ReportMetric(float64(len(ws)), "windows")
	if len(ws) == 0 {
		return
	}
	nsSeries := make([]string, len(ws))
	readSeries := make([]string, len(ws))
	for i, w := range ws {
		nsSeries[i] = fmt.Sprintf("%.0f", w.medNSPerOp)
		readSeries[i] = fmt.Sprintf("%.3f", float64(w.reads)/float64(w.ops))
	}
	b.Logf("windows (%d ops each, oldest->newest): ns/%s: %s | reads/%s: %s",
		ws[len(ws)-1].ops, denom, strings.Join(nsSeries, " "), denom, strings.Join(readSeries, " "))
	if len(ws) < 2 {
		return
	}
	last, prev := ws[len(ws)-1], ws[len(ws)-2]
	b.ReportMetric(last.medNSPerOp, "tail-ns/"+denom)
	if d, ok := pctDelta(last.medNSPerOp, prev.medNSPerOp); ok {
		b.ReportMetric(d, "conv-ns-%")
	}
	lastReads := float64(last.reads) / float64(last.ops)
	b.ReportMetric(lastReads, "tail-reads/"+denom)
	if d, ok := pctDelta(lastReads, float64(prev.reads)/float64(prev.ops)); ok {
		b.ReportMetric(d, "conv-reads-%")
	}
	if emitWrites {
		lastWrites := float64(last.writes) / float64(last.ops)
		b.ReportMetric(lastWrites, "tail-writes/"+denom)
		if d, ok := pctDelta(lastWrites, float64(prev.writes)/float64(prev.ops)); ok {
			b.ReportMetric(d, "conv-writes-%")
		}
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
		end := min(i+batch, to)
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
	// Distinct sub-DB per factory so iavl and bptree don't share a directory
	// (bptree-fast deliberately reuses bptree's via fixtureName).
	name := fmt.Sprintf("%s-disk", f.fixtureName())
	pdb, err := openDiskDB(name, dir)
	require.NoError(b, err)

	// Wrap so the tree's node reads/writes are counted for the per-op metrics
	// (the build below also goes through it, but readMeter only samples the
	// timed benchmark region, so build counts are excluded).
	cdb := newCountingDB(pdb)
	tree := f.newTree(cdb, *diskNodeCache)
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
	return diskFixture{tree: tree, db: cdb, n: n, close: closeFn}
}

// pointReadFn returns the per-op read function for the Get benches plus a closer.
// By default it's the working-tree Get (index-free for bptree); with
// -disk-committed-read it's a point read through a single committed snapshot held
// at the latest version — the ABCI-query path, where the bptree fast index
// engages (iavl's fast nodes engage on either path). The snapshot is opened ONCE
// and reused for the whole loop — deliberately not MutableTree.GetVersioned, which
// re-opens (GetImmutable) per call: a per-read open adds its root read to every op
// and would inflate reads/op by ~1 for every tree, hiding the index's win. Node
// memory stays bounded by the shared LRU since the snapshot pins only its root.
//
// Contract: a disk read bench must route EVERY lookup through the returned read fn
// (warmup included), never fx.tree.Get directly — otherwise -disk-committed-read is
// silently ignored for that bench and it measures the index-free working tree.
func pointReadFn(b *testing.B, fx diskFixture) (read func([]byte) ([]byte, error), closeFn func()) {
	b.Helper()
	if !*diskCommittedRead {
		return fx.tree.Get, func() {}
	}
	v := fx.tree.Version()
	get, c, err := fx.tree.CommittedReader(v)
	if err != nil {
		b.Fatalf("committed reader at v%d: %v", v, err)
	}
	b.Logf("reading via committed snapshot at version %d (ABCI-query path)", v)
	return get, func() { _ = c() }
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
			b.Logf("%s fixture: %d items, height %d", f.name, fx.tree.Size(), fx.tree.Height())
			read, closeReader := pointReadFn(b, fx)
			defer closeReader()
			rng := mrand.New(mrand.NewSource(1))
			var key [diskKeyLen]byte
			// Untimed warmup: move the LRU ramp's knee before the measured
			// region (the tail windows below verify whatever ramp remains).
			// Distinct seed: warm with a representative random set, not the
			// exact keys the timed loop will read. Gated on the real round —
			// go-bench's b.N=1 calibration pass needn't pay it.
			if b.N > 1 {
				warm := mrand.New(mrand.NewSource(11))
				for i := 0; i < *diskWarmupOps; i++ {
					if i > 0 && *diskReloadEvery > 0 && i%*diskReloadEvery == 0 {
						_, _ = fx.tree.Load()
					}
					putDiskKey(key[:], uint64(warm.Int63n(int64(n))))
					if _, err := read(key[:]); err != nil {
						b.Fatalf("warmup Get: %v", err)
					}
				}
			}
			// reads/op = node DB reads (Get) per point read. With
			// -disk-committed-read, IAVL and bptree-fast both ~1 (their fast
			// layer returns the value in one read); plain bptree ~height (full
			// node walk + out-of-line value fetch). Without the flag, reads route
			// through the index-free working tree for every tree. Fold around the
			// untimed reload so its reads aren't counted; each reload interval is
			// one convergence segment.
			rm := newReadMeter(fx.db)
			b.ResetTimer()
			segStart := 0
			for i := 0; i < b.N; i++ {
				if i > 0 && *diskReloadEvery > 0 && i%*diskReloadEvery == 0 {
					b.StopTimer()
					rm.fold(b, int64(i-segStart)) // close segment before the untimed reload
					segStart = i
					_, _ = fx.tree.Load() // bound memory; node LRU stays warm
					rm.snap(b)            // reopen segment after reload
					b.StartTimer()
				}
				putDiskKey(key[:], uint64(rng.Int63n(int64(n))))
				if _, err := read(key[:]); err != nil {
					b.Fatalf("Get: %v", err)
				}
			}
			rm.fold(b, int64(b.N-segStart))            // final segment
			rm.report(b, float64(b.N), "reads/op", "") // whole-run average (ramp-inclusive)
			rm.reportTail(b, "op", false)              // steady-state tail + convergence
			// After ResetTimer (which clears earlier ReportMetric extras).
			b.ReportMetric(float64(fx.tree.Height()), "height")
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
			b.Logf("%s fixture: %d items, height %d", f.name, fx.tree.Size(), fx.tree.Height())
			read, closeReader := pointReadFn(b, fx)
			defer closeReader()
			rng := mrand.New(mrand.NewSource(3))
			var key [diskKeyLen]byte
			if b.N > 1 {
				warm := mrand.New(mrand.NewSource(13))
				for i := 0; i < *diskWarmupOps; i++ {
					if i > 0 && *diskReloadEvery > 0 && i%*diskReloadEvery == 0 {
						_, _ = fx.tree.Load()
					}
					putDiskMissKey(key[:], uint64(warm.Int63n(int64(n))))
					if _, err := read(key[:]); err != nil {
						b.Fatalf("warmup Get: %v", err)
					}
				}
			}
			rm := newReadMeter(fx.db)
			b.ResetTimer()
			segStart := 0
			for i := 0; i < b.N; i++ {
				if i > 0 && *diskReloadEvery > 0 && i%*diskReloadEvery == 0 {
					b.StopTimer()
					rm.fold(b, int64(i-segStart))
					segStart = i
					_, _ = fx.tree.Load()
					rm.snap(b)
					b.StartTimer()
				}
				putDiskMissKey(key[:], uint64(rng.Int63n(int64(n))))
				if _, err := read(key[:]); err != nil {
					b.Fatalf("Get: %v", err)
				}
			}
			rm.fold(b, int64(b.N-segStart))
			rm.report(b, float64(b.N), "reads/op", "")
			rm.reportTail(b, "op", false)
		})
		fx.close()
	}
}

// BenchmarkDiskBlockWrite measures the cost of committing a block: -disk-block
// writes (a configurable mix of updates to existing keys and new inserts)
// followed by SaveVersion, against the large on-disk fixture. ns/op is the
// per-block latency; ns/write is also reported, along with reads/write and
// writes/write — the tree's node DB reads/stores per write (deterministic,
// backend-agnostic), which are the empirical basis for the write-depth gas
// params: bp32's shallow tree should show far fewer of both than IAVL's deep
// COW path, and the fast-node index can't help the write path. Pruning and the
// drop-in-memory-tree reload happen outside the timer (a real node prunes
// out-of-band and starts each block from committed state, lazily loading what
// its txs touch — which the timed Set path models).
func BenchmarkDiskBlockWrite(b *testing.B) {
	n := uint64(*diskKeys)
	bs := *diskBlock
	for _, f := range selectedFactories() {
		fx := ensureDiskFixture(b, f, n)
		b.Run(fmt.Sprintf("%s/%s/block-%d", f.name, humanCount(n), bs), func(b *testing.B) {
			b.ReportAllocs()
			b.Logf("%s fixture: %d items, height %d", f.name, fx.tree.Size(), fx.tree.Height())
			rng := mrand.New(mrand.NewSource(2))
			next := uint64(fx.tree.Size()) // fresh-insert index, past all existing keys
			// Untimed warmup blocks (rounded up from -disk-warmup-ops writes):
			// warm the node LRU so the metered region reflects steady state.
			if *diskWarmupOps > 0 && b.N > 1 {
				warm := mrand.New(mrand.NewSource(12))
				for wb := (*diskWarmupOps + bs - 1) / bs; wb > 0; wb-- {
					for j := range bs {
						k := make([]byte, diskKeyLen)
						v := make([]byte, diskValLen)
						if warm.Float64() < *diskUpdateFrac {
							putDiskKey(k, uint64(warm.Int63n(int64(n))))
						} else {
							putDiskKey(k, next)
							next++
						}
						putDiskVal(v, next+uint64(j))
						if _, err := fx.tree.Set(k, v); err != nil {
							b.Fatalf("warmup Set: %v", err)
						}
					}
					_, ver, err := fx.tree.SaveVersion()
					if err != nil {
						b.Fatalf("warmup SaveVersion: %v", err)
					}
					if ver > historySize {
						if err := fx.tree.DeleteVersionsTo(ver - historySize); err != nil {
							b.Fatalf("warmup prune: %v", err)
						}
					}
					if _, err := fx.tree.Load(); err != nil {
						b.Fatalf("warmup reload: %v", err)
					}
				}
			}
			// reads/write and writes/write = the tree's node DB reads/stores
			// per write, over the TIMED Set+SaveVersion region; snap()/fold()
			// bracket it so the untimed prune/reload between blocks are excluded.
			rm := newReadMeter(fx.db)
			b.ResetTimer()
			for i := 0; i < b.N; i++ { // one iteration == one block
				rm.snap(b)
				for j := range bs {
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
				rm.fold(b, int64(bs))
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
				rm.report(b, w, "reads/write", "writes/write") // whole-run averages
				rm.reportTail(b, "write", true)                // steady tail; median defeats compaction spikes
			}
			// Post-run size: this benchmark grew the fixture by the inserts in
			// b.N*bs writes; quantifies the per-run drift the start log warned of.
			b.Logf("%s fixture after run: %d items", f.name, fx.tree.Size())
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
		name := fmt.Sprintf("%s-disk", f.fixtureName())
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

// stderrLogger streams bptree Info/Warn/Error logs (e.g. fast-index rebuild
// progress) to stderr so a long backfill is observable. Debug is dropped.
type stderrLogger struct{}

func (stderrLogger) Info(msg string, kv ...any)  { fmt.Fprintf(os.Stderr, "  %s %v\n", msg, kv) }
func (stderrLogger) Warn(msg string, kv ...any)  { fmt.Fprintf(os.Stderr, "  WARN %s %v\n", msg, kv) }
func (stderrLogger) Error(msg string, kv ...any) { fmt.Fprintf(os.Stderr, "  ERROR %s %v\n", msg, kv) }
func (stderrLogger) Debug(string, ...any)        {}

// TestDiskFastIndexBackfill builds the inline 1-GET fast index onto an existing
// bptree disk fixture (one-time). Point -disk-dir at the fixture directory; it
// opens the "bptree-disk" sub-DB with the index enabled and Loads, which rebuilds the
// index from the latest tree — reading and re-storing every live value (slow at
// 100M, but one-time). Idempotent: a second run finds the stamp current and is a
// no-op. Progress streams to stderr.
//
//	go test ./tm2/pkg/bptree/benchmarks/ -run TestDiskFastIndexBackfill -v \
//	  -disk-dir=/data/bptree-bench -disk-node-cache=10000 -timeout=24h
func TestDiskFastIndexBackfill(t *testing.T) {
	if *diskDir == "" {
		t.Skip("set -disk-dir to the fixture directory")
	}
	pdb, err := openDiskDB("bptree-disk", *diskDir)
	require.NoError(t, err)
	defer pdb.Close()

	tree := bptree.NewMutableTreeWithDB(pdb, *diskNodeCache, stderrLogger{}, bptree.FastIndexOption(true))
	start := time.Now()
	v, err := tree.Load()
	require.NoError(t, err)
	t.Logf("fast index backfill complete: version=%d, tree size=%d, elapsed=%s",
		v, tree.Size(), time.Since(start))
}
