//go:build cgo

package benchstore

// Storage benchmarks for comparing DB backends (PebbleDB vs LMDB vs MDBX).
//
// Usage:
//
//	go test ./gnovm/cmd/benchstore/ -bench=. -benchmem -timeout=30m -db=pebbledb
//	go test ./gnovm/cmd/benchstore/ -bench=. -benchmem -timeout=30m -db=lmdb
//	go test ./gnovm/cmd/benchstore/ -bench=. -benchmem -timeout=30m -db=mdbx
//
// PebbleDB options:
//
//	-cache-mb=1024 -memtable-mb=128 -compactions=4
//
// Cache sweep (PebbleDB only):
//
//	go test ./gnovm/cmd/benchstore/ -bench=GetCacheSweep -timeout=6h \
//	    -db=pebbledb -sweep-keys=100000000 -cache-sweep=500,1024,2048,4096,8192

import (
	"encoding/binary"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/cockroachdb/pebble"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/lmdbdb"
	"github.com/gnolang/gno/tm2/pkg/db/mdbxdb"
	"github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

var (
	flagDB          = flag.String("db", "", "Database backend: pebbledb, lmdb, or mdbx (required)")
	flagCacheMB     = flag.Int("cache-mb", 0, "PebbleDB block cache size in MB (0 = use default 500MB)")
	flagMemtableMB  = flag.Int("memtable-mb", 0, "PebbleDB memtable size in MB (0 = use default 64MB)")
	flagCompactions = flag.Int("compactions", 0, "PebbleDB max concurrent compactions (0 = use default 3)")
	flagMaxKeys     = flag.Int("max-keys", 0, "Skip DB sizes above this many keys (0 = no limit)")
	flagCacheSweep  = flag.String("cache-sweep", "", "Comma-separated cache sizes in MB for GetCacheSweep (e.g. 500,1024,2048,4096,8192)")
	flagSweepKeys   = flag.Int("sweep-keys", 100_000_000, "Number of keys for GetCacheSweep benchmark")
)

var keySizes = []int{1_000, 10_000, 100_000, 1_000_000, 10_000_000, 100_000_000, 500_000_000, 750_000_000, 1_000_000_000}

func requireDB(b *testing.B) {
	b.Helper()
	if *flagDB == "" {
		b.Skip("use -db=pebbledb, -db=lmdb, or -db=mdbx to run")
	}
}

// ----------------------------------------
// Unified DB environment

type benchEnv struct {
	db  dbm.DB
	dir string
	n   int
}

func newBenchEnv(b *testing.B, n int, valSize int) *benchEnv {
	b.Helper()
	dir, err := os.MkdirTemp("", "gno-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	db, err := openDB("bench", dir)
	if err != nil {
		os.RemoveAll(dir)
		b.Fatal(err)
	}

	// Populate with varying values to avoid compression artifacts.
	val := make([]byte, valSize)
	prng := rand.New(rand.NewSource(0))
	batch := db.NewBatch()
	for i := 0; i < n; i++ {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(i))
		prng.Read(val)
		batch.Set(key, val)
		if (i+1)%10000 == 0 {
			batch.Write()
			batch.Close()
			batch = db.NewBatch()
			printProgress("populate", i+1, n)
		}
	}
	batch.Write()
	batch.Close()
	printProgress("populate", n, n)

	// Warmup: PebbleDB gets proportional random reads to fill block cache.
	// LMDB/MDBX use OS page cache via mmap — no explicit warmup needed.
	if *flagDB == "pebbledb" {
		cacheMB := 500 // default
		if *flagCacheMB > 0 {
			cacheMB = *flagCacheMB
		}
		warmupReads := int(int64(cacheMB) << 20 / 4096)
		if warmupReads > n {
			warmupReads = n
		}
		rng := rand.New(rand.NewSource(99))
		for i := 0; i < warmupReads; i++ {
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(rng.Intn(n)))
			db.Get(key)
			if (i+1)%10000 == 0 {
				printProgress("warmup", i+1, warmupReads)
			}
		}
		printProgress("warmup", warmupReads, warmupReads)
	}

	return &benchEnv{db: db, dir: dir, n: n}
}

func (env *benchEnv) Close() {
	env.db.Close()
	fmt.Fprintf(os.Stderr, "  db size: %s\n", dirSize(env.dir))
	os.RemoveAll(env.dir)
}

func openDB(name, dir string) (dbm.DB, error) {
	switch *flagDB {
	case "pebbledb":
		return pebbledb.NewPebbleDBWithOpts(name, dir, benchPebbleOpts())
	case "lmdb":
		// MapSize: 1TB default, enough for any benchmark.
		return lmdbdb.NewLMDBWithOptions(name, dir, lmdbdb.DefaultMapSize, 0)
	case "mdbx":
		return mdbxdb.NewMDBXWithOptions(name, dir, mdbxdb.DefaultMapSize, 0)
	default:
		return nil, fmt.Errorf("unknown -db=%q; use pebbledb, lmdb, or mdbx", *flagDB)
	}
}

func benchPebbleOpts() *pebble.Options {
	opts := pebbledb.DefaultPebbleOptions()
	if *flagCacheMB > 0 {
		opts.Cache = pebble.NewCache(int64(*flagCacheMB) << 20)
	}
	if *flagMemtableMB > 0 {
		opts.MemTableSize = uint64(*flagMemtableMB) << 20
	}
	if *flagCompactions > 0 {
		n := *flagCompactions
		opts.MaxConcurrentCompactions = func() int { return n }
	}
	return opts
}

// ----------------------------------------
// Helpers

func printProgress(label string, done, total int) {
	const width = 30
	filled := width * done / total
	fmt.Fprintf(os.Stderr, "\r  %s [%s%s] %d/%d",
		label,
		string(repeat('#', filled)),
		string(repeat(' ', width-filled)),
		done, total)
	if done == total {
		fmt.Fprint(os.Stderr, "\n")
	}
}

func repeat(ch byte, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = ch
	}
	return b
}

// noopLogger suppresses PebbleDB WAL replay log spam.
type noopLogger struct{}

func (noopLogger) Infof(format string, args ...interface{})  {}
func (noopLogger) Fatalf(format string, args ...interface{}) { panic(fmt.Sprintf(format, args...)) }

func dirSize(path string) string {
	var total int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			// Use Blocks * 512 for actual disk usage (handles sparse files).
			if st, ok := info.Sys().(*syscall.Stat_t); ok {
				total += st.Blocks * 512
			} else {
				total += info.Size()
			}
		}
		return nil
	})
	switch {
	case total >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(total)/(1<<30))
	case total >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(total)/(1<<20))
	default:
		return fmt.Sprintf("%.1f KB", float64(total)/(1<<10))
	}
}

// ----------------------------------------
// Benchmarks

func BenchmarkStoreGet(b *testing.B) {
	requireDB(b)
	for _, n := range keySizes {
		n := n
		if *flagMaxKeys > 0 && n > *flagMaxKeys {
			continue
		}
		var env *benchEnv
		b.Run(fmt.Sprintf("keys=%d", n), func(b *testing.B) {
			if env == nil {
				env = newBenchEnv(b, n, 256)
			}
			rng := rand.New(rand.NewSource(42))
			var sink []byte
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := make([]byte, 8)
				binary.BigEndian.PutUint64(key, uint64(rng.Intn(n)))
				sink, _ = env.db.Get(key)
			}
			runtime.KeepAlive(sink)
		})
		if env != nil {
			env.Close()
		}
	}
}

var batchSizes = []int{10, 100, 1000}

func BenchmarkStoreSetOverwrite(b *testing.B) {
	requireDB(b)
	for _, n := range keySizes {
		n := n
		if *flagMaxKeys > 0 && n > *flagMaxKeys {
			continue
		}
		var env *benchEnv
		for _, bs := range batchSizes {
			bs := bs
			b.Run(fmt.Sprintf("keys=%d/batch=%d", n, bs), func(b *testing.B) {
				if env == nil {
					env = newBenchEnv(b, n, 256)
				}
				rng := rand.New(rand.NewSource(42))
				val := make([]byte, 256)
				rng.Read(val)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					batch := env.db.NewBatch()
					for j := 0; j < bs; j++ {
						key := make([]byte, 8)
						binary.BigEndian.PutUint64(key, uint64(rng.Intn(n)))
						batch.Set(key, val)
					}
					batch.Write()
					batch.Close()
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(int64(b.N)*int64(bs)), "ns/key")
			})
		}
		if env != nil {
			env.Close()
		}
	}
}

func BenchmarkStoreSetInsert(b *testing.B) {
	requireDB(b)
	for _, n := range keySizes {
		n := n
		if *flagMaxKeys > 0 && n > *flagMaxKeys {
			continue
		}
		var env *benchEnv
		for _, bs := range batchSizes {
			bs := bs
			b.Run(fmt.Sprintf("keys=%d/batch=%d", n, bs), func(b *testing.B) {
				if env == nil {
					env = newBenchEnv(b, n, 256)
				}
				val := make([]byte, 256)
				//nolint:staticcheck // math/rand.Read deterministic is fine for bench payload
				rand.Read(val)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					batch := env.db.NewBatch()
					for j := 0; j < bs; j++ {
						key := make([]byte, 8)
						binary.BigEndian.PutUint64(key, uint64(n+i*bs+j))
						batch.Set(key, val)
					}
					batch.Write()
					batch.Close()
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(int64(b.N)*int64(bs)), "ns/key")
			})
		}
		if env != nil {
			env.Close()
		}
	}
}

func BenchmarkStoreDeleteAndInsert(b *testing.B) {
	requireDB(b)
	for _, n := range keySizes {
		n := n
		if *flagMaxKeys > 0 && n > *flagMaxKeys {
			continue
		}
		var env *benchEnv
		for _, bs := range batchSizes {
			bs := bs
			b.Run(fmt.Sprintf("keys=%d/batch=%d", n, bs), func(b *testing.B) {
				if env == nil {
					env = newBenchEnv(b, n, 256)
				}
				rng := rand.New(rand.NewSource(42))
				val := make([]byte, 256)
				rng.Read(val)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					batch := env.db.NewBatch()
					for j := 0; j < bs; j++ {
						// Delete existing key, re-insert at same key to keep DB stable.
						key := make([]byte, 8)
						binary.BigEndian.PutUint64(key, uint64(rng.Intn(n)))
						batch.Delete(key)
						batch.Set(key, val)
					}
					batch.Write()
					batch.Close()
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(int64(b.N)*int64(bs)), "ns/key")
			})
		}
		if env != nil {
			env.Close()
		}
	}
}

// ----------------------------------------
// Iterator benchmarks
//
// Measures the cost of iterator operations to calibrate IterNextCostFlat
// in tm2/pkg/store/types/gas.go. Two separate measurements:
//
//   - BenchmarkIterNext: per-step cost of Iterator.Next()+Value() after
//     the iterator is positioned. This is what IterNextCostFlat models.
//   - BenchmarkIterSeek: cost of opening a fresh iterator (tree walk to
//     the first leaf). Useful for deciding whether seek and step should
//     use a single constant or split.
//
// Usage:
//
//	go test ./gnovm/cmd/benchstore/ -bench=IterNext -timeout=30m -db=lmdb
//	go test ./gnovm/cmd/benchstore/ -bench=IterSeek -timeout=30m -db=lmdb

var iterKeySizes = []int{10_000, 100_000, 1_000_000, 10_000_000, 100_000_000}

func BenchmarkIterNext(b *testing.B) {
	requireDB(b)
	for _, n := range iterKeySizes {
		n := n
		if *flagMaxKeys > 0 && n > *flagMaxKeys {
			continue
		}
		var env *benchEnv
		b.Run(fmt.Sprintf("keys=%d", n), func(b *testing.B) {
			if env == nil {
				env = newBenchEnv(b, n, 256)
			}
			iter, _ := env.db.Iterator(nil, nil)
			var sinkK, sinkV []byte
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if !iter.Valid() {
					iter.Close()
					iter, _ = env.db.Iterator(nil, nil)
				}
				sinkK = iter.Key()
				sinkV = iter.Value()
				iter.Next()
			}
			b.StopTimer()
			iter.Close()
			runtime.KeepAlive(sinkK)
			runtime.KeepAlive(sinkV)
		})
		if env != nil {
			env.Close()
		}
	}
}

func BenchmarkIterSeek(b *testing.B) {
	requireDB(b)
	for _, n := range iterKeySizes {
		n := n
		if *flagMaxKeys > 0 && n > *flagMaxKeys {
			continue
		}
		var env *benchEnv
		b.Run(fmt.Sprintf("keys=%d", n), func(b *testing.B) {
			if env == nil {
				env = newBenchEnv(b, n, 256)
			}
			rng := rand.New(rand.NewSource(42))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				start := make([]byte, 8)
				binary.BigEndian.PutUint64(start, uint64(rng.Intn(n)))
				iter, _ := env.db.Iterator(start, nil)
				// Force positioning; without Valid()/Value() the backend may lazy-seek.
				if iter.Valid() {
					_ = iter.Value()
				}
				iter.Close()
			}
		})
		if env != nil {
			env.Close()
		}
	}
}

// ----------------------------------------
// PebbleDB cache sweep (PebbleDB only)

func BenchmarkStoreGetCacheSweep(b *testing.B) {
	if *flagDB != "pebbledb" {
		b.Skip("cache sweep is PebbleDB-only")
	}
	if *flagCacheSweep == "" {
		b.Skip("use -cache-sweep=500,1024,... to run")
	}
	sweep := strings.Split(*flagCacheSweep, ",")
	cacheSizes := make([]int, 0, len(sweep))
	for _, s := range sweep {
		mb, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			b.Fatalf("bad cache size %q: %v", s, err)
		}
		cacheSizes = append(cacheSizes, mb)
	}

	n := *flagSweepKeys

	// Populate once.
	dir, err := os.MkdirTemp("", "gno-cache-sweep-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	func() {
		db, err := pebbledb.NewPebbleDBWithOpts("bench", dir, benchPebbleOpts())
		if err != nil {
			b.Fatal(err)
		}
		defer db.Close()

		val := make([]byte, 256)
		prng := rand.New(rand.NewSource(0))
		batch := db.NewBatch()
		for i := 0; i < n; i++ {
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(i))
			prng.Read(val)
			batch.Set(key, val)
			if (i+1)%10000 == 0 {
				batch.Write()
				batch.Close()
				batch = db.NewBatch()
				printProgress("populate", i+1, n)
			}
		}
		batch.Write()
		batch.Close()
		printProgress("populate", n, n)
	}()

	fmt.Fprintf(os.Stderr, "  db size: %s\n", dirSize(dir))

	for _, mb := range cacheSizes {
		mb := mb

		opts := pebbledb.DefaultPebbleOptions()
		cache := pebble.NewCache(int64(mb) << 20)
		opts.Cache = cache
		opts.Logger = noopLogger{}

		db, err := pebbledb.NewPebbleDBWithOpts("bench", dir, opts)
		if err != nil {
			b.Fatal(err)
		}

		warmupReads := int(int64(mb) << 20 / 4096)
		if warmupReads > n {
			warmupReads = n
		}
		rng := rand.New(rand.NewSource(99))
		for i := 0; i < warmupReads; i++ {
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(rng.Intn(n)))
			db.Get(key)
			if (i+1)%10000 == 0 {
				printProgress(fmt.Sprintf("warmup %dMB", mb), i+1, warmupReads)
			}
		}
		printProgress(fmt.Sprintf("warmup %dMB", mb), warmupReads, warmupReads)

		b.Run(fmt.Sprintf("cache=%dMB/keys=%d", mb, n), func(b *testing.B) {
			rng := rand.New(rand.NewSource(42))
			var sink []byte
			for i := 0; i < b.N; i++ {
				key := make([]byte, 8)
				binary.BigEndian.PutUint64(key, uint64(rng.Intn(n)))
				sink, _ = db.Get(key)
			}
			runtime.KeepAlive(sink)
		})

		db.Close()
		cache.Unref()
	}
}

// ----------------------------------------
// Value size sweep
//
// Measures Get and SetOverwrite (batch=1000) at different value sizes.
// Key count is adaptive to stay within ~10GB of disk per test:
//   100B  → 10M keys  (~1 GB)
//   1KB   → 10M keys  (~10 GB)
//   10KB  → 1M keys   (~10 GB)
//   100KB → 100K keys (~10 GB)
//
// Usage:
//
//	go test ./gnovm/cmd/benchstore/ -bench=ValueSizeGet -timeout=2h -db=lmdb
//	go test ./gnovm/cmd/benchstore/ -bench=ValueSizeGet -timeout=2h -db=mdbx
//	go test ./gnovm/cmd/benchstore/ -bench=ValueSizeSet -timeout=2h -db=lmdb
//	go test ./gnovm/cmd/benchstore/ -bench=ValueSizeSet -timeout=2h -db=mdbx

var valueSizeCases = []struct {
	valSize int
	numKeys int
}{
	{100, 100_000_000},
	{1_000, 100_000_000},
	{10_000, 10_000_000},
	{100_000, 1_000_000},
}

func BenchmarkValueSizeGet(b *testing.B) {
	requireDB(b)
	for _, tc := range valueSizeCases {
		tc := tc
		var env *benchEnv
		b.Run(fmt.Sprintf("val=%dB/keys=%d", tc.valSize, tc.numKeys), func(b *testing.B) {
			if env == nil {
				env = newBenchEnvWithValSize(b, tc.numKeys, tc.valSize)
			}
			rng := rand.New(rand.NewSource(42))
			var sink []byte
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := make([]byte, 8)
				binary.BigEndian.PutUint64(key, uint64(rng.Intn(tc.numKeys)))
				sink, _ = env.db.Get(key)
			}
			runtime.KeepAlive(sink)
		})
		if env != nil {
			env.Close()
		}
	}
}

func BenchmarkValueSizeSet(b *testing.B) {
	requireDB(b)
	for _, tc := range valueSizeCases {
		tc := tc
		var env *benchEnv
		b.Run(fmt.Sprintf("val=%dB/keys=%d", tc.valSize, tc.numKeys), func(b *testing.B) {
			if env == nil {
				env = newBenchEnvWithValSize(b, tc.numKeys, tc.valSize)
			}
			rng := rand.New(rand.NewSource(42))
			val := make([]byte, tc.valSize)
			rng.Read(val)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				batch := env.db.NewBatch()
				for j := 0; j < 1000; j++ {
					key := make([]byte, 8)
					binary.BigEndian.PutUint64(key, uint64(rng.Intn(tc.numKeys)))
					batch.Set(key, val)
				}
				batch.Write()
				batch.Close()
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(int64(b.N)*1000), "ns/key")
		})
		if env != nil {
			env.Close()
		}
	}
}

// newBenchEnvWithValSize is like newBenchEnv but with configurable value size.
func newBenchEnvWithValSize(b *testing.B, n int, valSize int) *benchEnv {
	b.Helper()
	dir, err := os.MkdirTemp("", "gno-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	db, err := openDB("bench", dir)
	if err != nil {
		os.RemoveAll(dir)
		b.Fatal(err)
	}

	val := make([]byte, valSize)
	prng := rand.New(rand.NewSource(0))
	batch := db.NewBatch()
	for i := 0; i < n; i++ {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(i))
		prng.Read(val)
		batch.Set(key, val)
		if (i+1)%10000 == 0 {
			batch.Write()
			batch.Close()
			batch = db.NewBatch()
			printProgress("populate", i+1, n)
		}
	}
	batch.Write()
	batch.Close()
	printProgress("populate", n, n)

	// PebbleDB warmup.
	if *flagDB == "pebbledb" {
		cacheMB := 500
		if *flagCacheMB > 0 {
			cacheMB = *flagCacheMB
		}
		warmupReads := int(int64(cacheMB) << 20 / 4096)
		if warmupReads > n {
			warmupReads = n
		}
		rng := rand.New(rand.NewSource(99))
		for i := 0; i < warmupReads; i++ {
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(rng.Intn(n)))
			db.Get(key)
			if (i+1)%10000 == 0 {
				printProgress("warmup", i+1, warmupReads)
			}
		}
		printProgress("warmup", warmupReads, warmupReads)
	}

	return &benchEnv{db: db, dir: dir, n: n}
}
