package benchmarks

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	mrand "math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/stretchr/testify/require"

	bptree "github.com/gnolang/gno/tm2/pkg/bptree"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/iavl"
)

var benchBackend = flag.String("backend", "memdb", "DB backend for benchmarks: memdb, goleveldb")

const (
	historySize = 20
	keyLen      = 16
	dataLen     = 40

	// pregenPool is the maximum number of pre-generated random keys/values.
	// Benchmarks cycle through this pool with i%pregenPool to avoid
	// allocating b.N items (which can be 10-30M for fast operations).
	pregenPool = 10_000
)

// -----------------------------------------------------------------------
// TreeBench interface — common abstraction over IAVL and bptree
// -----------------------------------------------------------------------

type TreeBench interface {
	Set(key, value []byte) (bool, error)
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Remove(key []byte) ([]byte, bool, error)
	SaveVersion() ([]byte, int64, error)
	LoadVersion(version int64) (int64, error)
	DeleteVersionsTo(toVersion int64) error
	Iterator(start, end []byte, ascending bool) (dbm.Iterator, error)
	GetMembershipProof(key []byte) (*ics23.CommitmentProof, error)
	GetNonMembershipProof(key []byte) (*ics23.CommitmentProof, error)
	Hash() []byte
	WorkingHash() []byte
	Size() int64
	Height() int8
	Version() int64
	Close() error
}

// -----------------------------------------------------------------------
// IAVL wrapper
// -----------------------------------------------------------------------

type iavlTree struct {
	t *iavl.MutableTree
}

func newIAVLTree(db dbm.DB, cacheSize int) *iavlTree {
	return &iavlTree{t: iavl.NewMutableTree(db, cacheSize, false, iavl.NewNopLogger())}
}

func (w *iavlTree) Set(k, v []byte) (bool, error)        { return w.t.Set(k, v) }
func (w *iavlTree) Get(k []byte) ([]byte, error)          { return w.t.Get(k) }
func (w *iavlTree) Has(k []byte) (bool, error)            { return w.t.Has(k) }
func (w *iavlTree) Remove(k []byte) ([]byte, bool, error) { return w.t.Remove(k) }
func (w *iavlTree) SaveVersion() ([]byte, int64, error)   { return w.t.SaveVersion() }
func (w *iavlTree) LoadVersion(v int64) (int64, error)    { return w.t.LoadVersion(v) }
func (w *iavlTree) DeleteVersionsTo(v int64) error        { return w.t.DeleteVersionsTo(v) }
func (w *iavlTree) Iterator(start, end []byte, asc bool) (dbm.Iterator, error) {
	return w.t.Iterator(start, end, asc)
}
func (w *iavlTree) GetMembershipProof(k []byte) (*ics23.CommitmentProof, error) {
	return w.t.GetMembershipProof(k)
}
func (w *iavlTree) GetNonMembershipProof(k []byte) (*ics23.CommitmentProof, error) {
	return w.t.GetNonMembershipProof(k)
}
func (w *iavlTree) Hash() []byte        { return w.t.Hash() }
func (w *iavlTree) WorkingHash() []byte { return w.t.WorkingHash() }
func (w *iavlTree) Size() int64         { return w.t.Size() }
func (w *iavlTree) Height() int8        { return w.t.Height() }
func (w *iavlTree) Version() int64      { return w.t.Version() }
func (w *iavlTree) Close() error        { return w.t.Close() }

// -----------------------------------------------------------------------
// bptree wrapper
// -----------------------------------------------------------------------

type bptreeTree struct {
	t *bptree.MutableTree
}

func newBptreeTree(db dbm.DB, cacheSize int) *bptreeTree {
	return &bptreeTree{t: bptree.NewMutableTreeWithDB(db, cacheSize, bptree.NewNopLogger())}
}

func (w *bptreeTree) Set(k, v []byte) (bool, error)        { return w.t.Set(k, v) }
func (w *bptreeTree) Get(k []byte) ([]byte, error)          { return w.t.Get(k) }
func (w *bptreeTree) Has(k []byte) (bool, error)            { return w.t.Has(k) }
func (w *bptreeTree) Remove(k []byte) ([]byte, bool, error) { return w.t.Remove(k) }
func (w *bptreeTree) SaveVersion() ([]byte, int64, error)   { return w.t.SaveVersion() }
func (w *bptreeTree) LoadVersion(v int64) (int64, error)    { return w.t.LoadVersion(v) }
func (w *bptreeTree) DeleteVersionsTo(v int64) error        { return w.t.DeleteVersionsTo(v) }
func (w *bptreeTree) Iterator(start, end []byte, asc bool) (dbm.Iterator, error) {
	itr, err := w.t.Iterator(start, end, asc)
	return itr, err
}
func (w *bptreeTree) GetMembershipProof(k []byte) (*ics23.CommitmentProof, error) {
	return w.t.GetMembershipProof(k)
}
func (w *bptreeTree) GetNonMembershipProof(k []byte) (*ics23.CommitmentProof, error) {
	return w.t.GetNonMembershipProof(k)
}
func (w *bptreeTree) Hash() []byte        { return w.t.Hash() }
func (w *bptreeTree) WorkingHash() []byte { return w.t.WorkingHash() }
func (w *bptreeTree) Size() int64         { return w.t.Size() }
func (w *bptreeTree) Height() int8        { return w.t.Height() }
func (w *bptreeTree) Version() int64      { return w.t.Version() }
func (w *bptreeTree) Close() error        { return w.t.Close() }

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

func randBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b) //nolint:errcheck
	return b
}

func pregenKeys(n int, kLen int) [][]byte {
	keys := make([][]byte, n)
	for i := range keys {
		keys[i] = randBytes(kLen)
	}
	return keys
}

func pregenVals(n int, dLen int) [][]byte {
	vals := make([][]byte, n)
	for i := range vals {
		vals[i] = randBytes(dLen)
	}
	return vals
}

func prepareTree(b *testing.B, tree TreeBench, size, kLen, dLen int) [][]byte {
	b.Helper()
	keys := make([][]byte, size)
	for i := 0; i < size; i++ {
		k := randBytes(kLen)
		_, err := tree.Set(k, randBytes(dLen))
		require.NoError(b, err)
		keys[i] = k
	}
	commitTree(b, tree)
	runtime.GC()
	return keys
}

func commitTree(b *testing.B, tree TreeBench) {
	b.Helper()
	tree.WorkingHash()
	_, version, err := tree.SaveVersion()
	require.NoError(b, err)
	if version > historySize {
		require.NoError(b, tree.DeleteVersionsTo(version-historySize))
	}
}

func memUseMB() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / (1024 * 1024)
}

func dirSizeMB(dir string) float64 {
	var total int64
	filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		total += info.Size()
		return nil
	})
	return float64(total) / (1024 * 1024)
}

type dbInfo struct {
	db      dbm.DB
	dir     string
	cleanup func()
}

func makeDB(b *testing.B, backend string) dbInfo {
	b.Helper()
	switch backend {
	case "memdb":
		db := memdb.NewMemDB()
		return dbInfo{db: db, cleanup: func() { db.Close() }}
	case "goleveldb":
		dir := b.TempDir()
		db, err := goleveldb.NewGoLevelDB("bench", dir)
		require.NoError(b, err)
		return dbInfo{db: db, dir: dir, cleanup: func() { db.Close() }}
	default:
		b.Fatalf("unknown backend: %s", backend)
		return dbInfo{}
	}
}

func cacheForSize(size int) int {
	if size <= 100_000 {
		return 500
	}
	return 10_000
}

type treeFactory struct {
	name    string
	newTree func(db dbm.DB, cache int) TreeBench
}

var factories = []treeFactory{
	{"iavl", func(db dbm.DB, c int) TreeBench { return newIAVLTree(db, c) }},
	{"bptree", func(db dbm.DB, c int) TreeBench { return newBptreeTree(db, c) }},
}

// buildTree constructs a tree once outside b.Run so it is NOT rebuilt on
// every b.N calibration round. This is safe for read-only benchmarks and
// benchmarks that only update existing keys (no structural changes).
type builtTree struct {
	tree    TreeBench
	keys    [][]byte
	cleanup func()
}

func buildTree(b *testing.B, f treeFactory, sz int) builtTree {
	b.Helper()
	di := makeDB(b, *benchBackend)
	tree := f.newTree(di.db, cacheForSize(sz))
	keys := prepareTree(b, tree, sz, keyLen, dataLen)
	return builtTree{
		tree: tree, keys: keys,
		cleanup: func() { tree.Close(); di.cleanup() },
	}
}

// -----------------------------------------------------------------------
// Single Operations
// -----------------------------------------------------------------------

func BenchmarkGetHit(b *testing.B) {
	sizes := []int{1_000, 10_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			l := int32(len(bt.keys))
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					bt.tree.Get(bt.keys[mrand.Int31n(l)])
				}
			})
			bt.cleanup()
		}
	}
}

func BenchmarkGetMiss(b *testing.B) {
	missKeys := pregenKeys(pregenPool, keyLen)
	sizes := []int{1_000, 10_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					bt.tree.Get(missKeys[i%pregenPool])
				}
			})
			bt.cleanup()
		}
	}
}

func BenchmarkHas(b *testing.B) {
	sizes := []int{1_000, 10_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			l := int32(len(bt.keys))
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					bt.tree.Has(bt.keys[mrand.Int31n(l)])
				}
			})
			bt.cleanup()
		}
	}
}

func BenchmarkSetInsert(b *testing.B) {
	sizes := []int{1_000, 10_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			f := f
			sz := sz
			b.Run(name, func(b *testing.B) {
				b.StopTimer()
				di := makeDB(b, *benchBackend)
				defer di.cleanup()
				tree := f.newTree(di.db, cacheForSize(sz))
				_ = prepareTree(b, tree, sz, keyLen, dataLen)
				defer tree.Close()
				// Use inline randBytes: ~200ns overhead is <5% of Set (~2-8us).
				// Avoids allocating b.N keys+vals (can be 1M+ items = 56MB+).
				b.ReportAllocs()
				b.StartTimer()
				for i := 0; i < b.N; i++ {
					tree.Set(randBytes(keyLen), randBytes(dataLen))
				}
			})
		}
	}
}

func BenchmarkSetUpdate(b *testing.B) {
	vals := pregenVals(pregenPool, dataLen)
	sizes := []int{1_000, 10_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			l := int32(len(bt.keys))
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					bt.tree.Set(bt.keys[mrand.Int31n(l)], vals[i%pregenPool])
				}
			})
			bt.cleanup()
		}
	}
}

func BenchmarkRemove(b *testing.B) {
	const batchSize = 100 // re-insert batch after this many removes (1 timer toggle per batch)
	sizes := []int{1_000, 10_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			f := f
			sz := sz
			b.Run(name, func(b *testing.B) {
				b.StopTimer()
				di := makeDB(b, *benchBackend)
				defer di.cleanup()
				tree := f.newTree(di.db, cacheForSize(sz))
				keys := prepareTree(b, tree, sz, keyLen, dataLen)
				defer tree.Close()
				// Select a fixed batch of keys to remove/re-insert repeatedly.
				// Each Remove is a guaranteed hit. Tree shrinks by at most
				// batchSize/sz (0.1-10%) before being restored. Timer overhead
				// is 1 StopTimer/StartTimer per batchSize ops (~0.005%).
				perm := mrand.Perm(sz)
				removeSet := make([][]byte, batchSize)
				for i := range removeSet {
					removeSet[i] = keys[perm[i]]
				}
				reinsertVal := randBytes(dataLen)
				b.ReportAllocs()
				b.StartTimer()
				for i := 0; i < b.N; i++ {
					idx := i % batchSize
					tree.Remove(removeSet[idx])
					if idx == batchSize-1 {
						b.StopTimer()
						for _, k := range removeSet {
							tree.Set(k, reinsertVal)
						}
						b.StartTimer()
					}
				}
			})
		}
	}
}

// -----------------------------------------------------------------------
// Iteration
// -----------------------------------------------------------------------

func BenchmarkIterationFull(b *testing.B) {
	sizes := []int{1_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					itr, _ := bt.tree.Iterator(nil, nil, true)
					for ; itr.Valid(); itr.Next() {
						_ = itr.Key()
						_ = itr.Value()
					}
					itr.Close()
				}
			})
			bt.cleanup()
		}
	}
}

func BenchmarkIterationDescending(b *testing.B) {
	sizes := []int{1_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					itr, _ := bt.tree.Iterator(nil, nil, false)
					for ; itr.Valid(); itr.Next() {
						_ = itr.Key()
						_ = itr.Value()
					}
					itr.Close()
				}
			})
			bt.cleanup()
		}
	}
}

func BenchmarkIterationRange(b *testing.B) {
	sizes := []int{1_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			sorted := make([][]byte, len(bt.keys))
			copy(sorted, bt.keys)
			sort.Slice(sorted, func(i, j int) bool {
				return bytes.Compare(sorted[i], sorted[j]) < 0
			})
			rangeSize := sz / 100
			if rangeSize < 1 {
				rangeSize = 1
			}
			startIdx := sz / 2
			endIdx := startIdx + rangeSize
			if endIdx >= sz {
				endIdx = sz - 1
			}
			start := sorted[startIdx]
			end := sorted[endIdx]
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					itr, _ := bt.tree.Iterator(start, end, true)
					for ; itr.Valid(); itr.Next() {
						_ = itr.Key()
						_ = itr.Value()
					}
					itr.Close()
				}
			})
			bt.cleanup()
		}
	}
}

// -----------------------------------------------------------------------
// Block workload
// -----------------------------------------------------------------------

func BenchmarkBlock(b *testing.B) {
	blockSizes := []int{100, 500}
	sz := 100_000
	for _, bs := range blockSizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/block-%d", f.name, bs)
			f := f
			bs := bs
			b.Run(name, func(b *testing.B) {
				b.StopTimer()
				di := makeDB(b, *benchBackend)
				defer di.cleanup()
				tree := f.newTree(di.db, cacheForSize(sz))
				keys := prepareTree(b, tree, sz, keyLen, dataLen)
				defer tree.Close()
				l := int32(len(keys))
				b.ReportAllocs()
				b.StartTimer()
				for i := 0; i < b.N; i++ {
					for j := 0; j < bs; j++ {
						if j%2 == 0 {
							tree.Set(randBytes(keyLen), randBytes(dataLen))
						} else {
							tree.Set(keys[mrand.Int31n(l)], randBytes(dataLen))
						}
					}
					commitTree(b, tree)
				}
			})
		}
	}
}

// -----------------------------------------------------------------------
// Versioning
// -----------------------------------------------------------------------

func BenchmarkSaveVersion(b *testing.B) {
	sizes := []int{1_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			f := f
			sz := sz
			b.Run(name, func(b *testing.B) {
				b.StopTimer()
				di := makeDB(b, *benchBackend)
				defer di.cleanup()
				tree := f.newTree(di.db, cacheForSize(sz))
				_ = prepareTree(b, tree, sz, keyLen, dataLen)
				defer tree.Close()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					for j := 0; j < 100; j++ {
						tree.Set(randBytes(keyLen), randBytes(dataLen))
					}
					b.StartTimer()
					tree.SaveVersion()
				}
			})
		}
	}
}

func BenchmarkLoadVersion(b *testing.B) {
	sizes := []int{1_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			f := f
			sz := sz
			b.Run(name, func(b *testing.B) {
				b.StopTimer()
				di := makeDB(b, "goleveldb")
				defer di.cleanup()
				tree := f.newTree(di.db, cacheForSize(sz))
				_ = prepareTree(b, tree, sz, keyLen, dataLen)
				ver := tree.Version()
				tree.Close()
				b.ReportAllocs()
				b.StartTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					t2 := f.newTree(di.db, cacheForSize(sz))
					b.StartTimer()
					t2.LoadVersion(ver)
					b.StopTimer()
					t2.Close()
				}
			})
		}
	}
}

func BenchmarkMultiVersionCreate(b *testing.B) {
	versions := []int{10, 100}
	baseSz := 10_000
	for _, nv := range versions {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%d-versions", f.name, nv)
			f := f
			nv := nv
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					di := makeDB(b, *benchBackend)
					tree := f.newTree(di.db, cacheForSize(baseSz))
					_ = prepareTree(b, tree, baseSz, keyLen, dataLen)
					b.StartTimer()
					for v := 0; v < nv; v++ {
						for j := 0; j < 50; j++ {
							tree.Set(randBytes(keyLen), randBytes(dataLen))
						}
						tree.SaveVersion()
					}
					b.StopTimer()
					tree.Close()
					di.cleanup()
				}
			})
		}
	}
}

// -----------------------------------------------------------------------
// Pruning
// -----------------------------------------------------------------------

func BenchmarkPrune(b *testing.B) {
	baseSz := 10_000
	for _, f := range factories {
		name := f.name
		f := f
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				di := makeDB(b, *benchBackend)
				tree := f.newTree(di.db, cacheForSize(baseSz))
				_ = prepareTree(b, tree, baseSz, keyLen, dataLen)
				for v := 0; v < 100; v++ {
					for j := 0; j < 100; j++ {
						tree.Set(randBytes(keyLen), randBytes(dataLen))
					}
					tree.SaveVersion()
				}
				b.StartTimer()
				tree.DeleteVersionsTo(50)
				b.StopTimer()
				tree.Close()
				di.cleanup()
			}
		})
	}
}

// -----------------------------------------------------------------------
// Proofs
// -----------------------------------------------------------------------

func BenchmarkMembershipProof(b *testing.B) {
	sizes := []int{1_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			l := int32(len(bt.keys))
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				var lastSize int
				for i := 0; i < b.N; i++ {
					proof, _ := bt.tree.GetMembershipProof(bt.keys[mrand.Int31n(l)])
					lastSize = proof.Size()
				}
				b.ReportMetric(float64(lastSize), "proof-bytes")
			})
			bt.cleanup()
		}
	}
}

func BenchmarkNonMembershipProof(b *testing.B) {
	missKeys := pregenKeys(pregenPool, keyLen)
	sizes := []int{1_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				var lastSize int
				for i := 0; i < b.N; i++ {
					proof, _ := bt.tree.GetNonMembershipProof(missKeys[i%pregenPool])
					lastSize = proof.Size()
				}
				b.ReportMetric(float64(lastSize), "proof-bytes")
			})
			bt.cleanup()
		}
	}
}

// -----------------------------------------------------------------------
// WorkingHash
// -----------------------------------------------------------------------

func BenchmarkWorkingHash(b *testing.B) {
	// Measures WorkingHash after 10 mutations (the realistic usage pattern).
	// Mutations are included in timed work to avoid b.StopTimer/b.StartTimer
	// per iteration, which causes b.N explosion and timeouts.
	sizes := []int{1_000, 10_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			f := f
			sz := sz
			b.Run(name, func(b *testing.B) {
				b.StopTimer()
				di := makeDB(b, *benchBackend)
				defer di.cleanup()
				tree := f.newTree(di.db, cacheForSize(sz))
				_ = prepareTree(b, tree, sz, keyLen, dataLen)
				defer tree.Close()
				b.ReportAllocs()
				b.StartTimer()
				for i := 0; i < b.N; i++ {
					for j := 0; j < 10; j++ {
						tree.Set(randBytes(keyLen), randBytes(dataLen))
					}
					tree.WorkingHash()
				}
			})
		}
	}
}

// -----------------------------------------------------------------------
// Disk Space
// -----------------------------------------------------------------------

func BenchmarkDiskSpace(b *testing.B) {
	sizes := []int{1_000, 10_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			f := f
			sz := sz
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					di := makeDB(b, "goleveldb")
					tree := f.newTree(di.db, cacheForSize(sz))
					_ = prepareTree(b, tree, sz, keyLen, dataLen)
					tree.Close()
					di.cleanup()
					mb := dirSizeMB(di.dir)
					b.ReportMetric(mb, "MB")
					b.ReportMetric(mb*1024*1024/float64(sz), "bytes/key")
				}
			})
		}
	}
}

func BenchmarkDiskSpaceMultiVersion(b *testing.B) {
	versionCounts := []int{10, 100}
	baseSz := 10_000
	for _, nv := range versionCounts {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%d-versions", f.name, nv)
			f := f
			nv := nv
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					di := makeDB(b, "goleveldb")
					tree := f.newTree(di.db, cacheForSize(baseSz))
					_ = prepareTree(b, tree, baseSz, keyLen, dataLen)
					for v := 0; v < nv; v++ {
						for j := 0; j < 50; j++ {
							tree.Set(randBytes(keyLen), randBytes(dataLen))
						}
						tree.SaveVersion()
					}
					tree.Close()
					di.cleanup()
					mb := dirSizeMB(di.dir)
					b.ReportMetric(mb, "MB")
				}
			})
		}
	}
}

// -----------------------------------------------------------------------
// Memory
// -----------------------------------------------------------------------

func BenchmarkMemory(b *testing.B) {
	sizes := []int{1_000, 10_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			f := f
			sz := sz
			b.Run(name, func(b *testing.B) {
				// Tree construction IS the timed work so b.N stays at 1-2.
				// Previous version timed only Has() (~100ns) causing b.N to
				// explode while each iteration rebuilt a full tree (seconds).
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					di := makeDB(b, *benchBackend)
					tree := f.newTree(di.db, cacheForSize(sz))
					runtime.GC()
					before := memUseMB()
					_ = prepareTree(b, tree, sz, keyLen, dataLen)
					runtime.GC()
					after := memUseMB()
					used := after - before
					b.ReportMetric(used, "MB")
					b.ReportMetric(used*1024*1024/float64(sz), "bytes/key")
					tree.Close()
					di.cleanup()
				}
			})
		}
	}
}

// -----------------------------------------------------------------------
// Scaling
// -----------------------------------------------------------------------

func BenchmarkScalingGet(b *testing.B) {
	sizes := []int{1_000, 10_000, 100_000, 1_000_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			l := int32(len(bt.keys))
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				b.ReportMetric(float64(bt.tree.Height()), "height")
				for i := 0; i < b.N; i++ {
					bt.tree.Get(bt.keys[mrand.Int31n(l)])
				}
			})
			bt.cleanup()
		}
	}
}

func BenchmarkScalingSet(b *testing.B) {
	vals := pregenVals(pregenPool, dataLen)
	sizes := []int{1_000, 10_000, 100_000, 1_000_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			bt := buildTree(b, f, sz)
			l := int32(len(bt.keys))
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					bt.tree.Set(bt.keys[mrand.Int31n(l)], vals[i%pregenPool])
				}
			})
			bt.cleanup()
		}
	}
}

func BenchmarkScalingSaveVersion(b *testing.B) {
	sizes := []int{1_000, 10_000, 100_000}
	for _, sz := range sizes {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%dk", f.name, sz/1000)
			f := f
			sz := sz
			b.Run(name, func(b *testing.B) {
				b.StopTimer()
				di := makeDB(b, *benchBackend)
				defer di.cleanup()
				tree := f.newTree(di.db, cacheForSize(sz))
				_ = prepareTree(b, tree, sz, keyLen, dataLen)
				defer tree.Close()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					for j := 0; j < 100; j++ {
						tree.Set(randBytes(keyLen), randBytes(dataLen))
					}
					b.StartTimer()
					tree.SaveVersion()
				}
			})
		}
	}
}

// -----------------------------------------------------------------------
// Backend Comparison
// -----------------------------------------------------------------------

func BenchmarkBackends(b *testing.B) {
	backends := []string{"memdb", "goleveldb"}
	sz := 100_000
	for _, be := range backends {
		for _, f := range factories {
			name := fmt.Sprintf("%s/%s", f.name, be)
			f := f
			be := be
			b.Run(name, func(b *testing.B) {
				b.StopTimer()
				di := makeDB(b, be)
				defer di.cleanup()
				tree := f.newTree(di.db, cacheForSize(sz))
				keys := prepareTree(b, tree, sz, keyLen, dataLen)
				defer tree.Close()
				l := int32(len(keys))
				b.ReportAllocs()
				b.StartTimer()
				for i := 0; i < b.N; i++ {
					r := mrand.Float32()
					switch {
					case r < 0.70:
						tree.Get(keys[mrand.Int31n(l)])
					case r < 0.90:
						tree.Set(keys[mrand.Int31n(l)], randBytes(dataLen))
					default:
						tree.Set(randBytes(keyLen), randBytes(dataLen))
					}
					if (i+1)%500 == 0 {
						commitTree(b, tree)
					}
				}
			})
		}
	}
}
