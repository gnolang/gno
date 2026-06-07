package benchmarks

// Large-scale, disk-bound benchmark comparison (IAVL vs B+32) on pebbledb.
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

	"github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

var (
	diskDir         = flag.String("disk-dir", "", "persistent dir for disk fixtures; empty = ephemeral TempDir (fixture rebuilt each run)")
	diskKeys        = flag.Int64("disk-keys", 1_000_000, "fixture size N in keys (set 100000000 for the realistic disk-bound comparison)")
	diskBlock       = flag.Int("disk-block", 1000, "writes per block (SaveVersion cadence) for the block-write benchmark")
	diskNodeCache   = flag.Int("disk-node-cache", 10000, "in-process node LRU cache size, in nodes (production-realistic)")
	diskUpdateFrac  = flag.Float64("disk-update-frac", 0.5, "fraction of block writes that update existing keys (rest insert new keys)")
	diskBuildBatch  = flag.Int64("disk-build-batch", 100_000, "keys per SaveVersion while building the fixture")
	diskReloadEvery = flag.Int("disk-reload-every", 100_000, "reload latest every N read ops to bound resident memory (the node LRU stays warm across reloads)")
	diskFactory     = flag.String("disk-factory", "", "limit disk populate/benchmarks to one backend: iavl|bptree (empty = both). Lets two processes populate in parallel into one -disk-dir.")
)

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
	n     uint64
	close func()
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
	for i := from; i < to; {
		end := i + batch
		if end > to {
			end = to
		}
		for ; i < end; i++ {
			k := make([]byte, diskKeyLen)
			v := make([]byte, diskValLen)
			putDiskKey(k, i)
			putDiskVal(v, i)
			if _, err := tree.Set(k, v); err != nil {
				tb.Fatalf("%s build Set: %v", label, err)
			}
		}
		_, ver, err := tree.SaveVersion()
		if err != nil {
			tb.Fatalf("%s build SaveVersion: %v", label, err)
		}
		if ver > historySize {
			if err := tree.DeleteVersionsTo(ver - historySize); err != nil {
				tb.Fatalf("%s build prune: %v", label, err)
			}
		}
		if _, err := tree.Load(); err != nil { // drop in-mem tree; node LRU stays warm
			tb.Fatalf("%s build reload: %v", label, err)
		}
		if logProgress && (i%(batch*10) == 0 || i == to) {
			tb.Logf("  %s: %d/%d keys", label, i, to)
		}
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
	pdb, err := pebbledb.NewPebbleDBWithOpts(name, dir, pebbledb.DefaultPebbleOptions())
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
	return diskFixture{tree: tree, n: n, close: closeFn}
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
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i > 0 && *diskReloadEvery > 0 && i%*diskReloadEvery == 0 {
					b.StopTimer()
					_, _ = fx.tree.Load() // bound memory; node LRU stays warm
					b.StartTimer()
				}
				putDiskKey(key[:], uint64(rng.Int63n(int64(n))))
				if _, err := fx.tree.Get(key[:]); err != nil {
					b.Fatalf("Get: %v", err)
				}
			}
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
// per-block latency; ns/write is also reported. Pruning and the
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
			rng := mrand.New(mrand.NewSource(2))
			next := uint64(fx.tree.Size()) // fresh-insert index, past all existing keys
			b.ResetTimer()
			for i := 0; i < b.N; i++ { // one iteration == one block
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
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*bs), "ns/write")
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
		pdb, err := pebbledb.NewPebbleDBWithOpts(name, *diskDir, pebbledb.DefaultPebbleOptions())
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
