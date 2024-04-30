package benchmarks

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/_all"
	"github.com/gnolang/gno/tm2/pkg/iavl"
)

const historySize = 20

func randBytes(length int) []byte {
	key := make([]byte, length)
	// math.rand.Read always returns err=nil
	// we do not need cryptographic randomness for this test:
	rand.Read(key)
	return key
}

func prepareTree(b *testing.B, db db.DB, size, keyLen, dataLen int) (*iavl.MutableTree, [][]byte) {
	b.Helper()

	t := iavl.NewMutableTree(db, size)
	keys := make([][]byte, size)

	for i := 0; i < size; i++ {
		key := randBytes(keyLen)
		t.Set(key, randBytes(dataLen))
		keys[i] = key
	}
	commitTree(b, t)
	runtime.GC()
	return t, keys
}

// commit tree saves a new version and deletes and old one...
func commitTree(b *testing.B, t *iavl.MutableTree) {
	b.Helper()

	t.Hash()
	_, version, err := t.SaveVersion()
	if err != nil {
		b.Errorf("Can't save: %v", err)
	}
	if version > historySize {
		err = t.DeleteVersion(version - historySize)
		if err != nil {
			b.Errorf("Can't delete: %v", err)
		}
	}
}

func runQueries(b *testing.B, t *iavl.MutableTree, keyLen int) {
	b.Helper()

	for i := 0; i < b.N; i++ {
		q := randBytes(keyLen)
		t.Get(q)
	}
}

func runKnownQueries(b *testing.B, t *iavl.MutableTree, keys [][]byte) {
	b.Helper()

	l := int32(len(keys))
	for i := 0; i < b.N; i++ {
		q := keys[rand.Int31n(l)]
		t.Get(q)
	}
}

func runUpdate(b *testing.B, t *iavl.MutableTree, dataLen, blockSize int, keys [][]byte) *iavl.MutableTree {
	b.Helper()

	l := int32(len(keys))
	for i := 1; i <= b.N; i++ {
		key := keys[rand.Int31n(l)]
		t.Set(key, randBytes(dataLen))
		if i%blockSize == 0 {
			commitTree(b, t)
		}
	}
	return t
}

// runBlock measures time for an entire block, not just one tx
func runBlock(b *testing.B, t *iavl.MutableTree, keyLen, dataLen, blockSize int, keys [][]byte) *iavl.MutableTree {
	b.Helper()

	l := int32(len(keys))

	// XXX: This was adapted to work with VersionedTree but needs to be re-thought.

	lastCommit := t
	realTree := t
	// check := t

	for i := 0; i < b.N; i++ {
		for j := 0; j < blockSize; j++ {
			// 50% insert, 50% update
			var key []byte
			if i%2 == 0 {
				key = keys[rand.Int31n(l)]
			} else {
				key = randBytes(keyLen)
			}
			data := randBytes(dataLen)

			// perform query and write on check and then real
			// check.Get(key)
			// check.Set(key, data)
			realTree.Get(key)
			realTree.Set(key, data)
		}

		// at the end of a block, move it all along....
		commitTree(b, realTree)
		lastCommit = realTree
	}

	return lastCommit
}

func BenchmarkRandomBytes(b *testing.B) {
	benchmarks := []struct {
		length int
	}{
		{4}, {16}, {32}, {100}, {1000},
	}
	for _, bench := range benchmarks {
		name := fmt.Sprintf("random-%d", bench.length)
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				randBytes(bench.length)
			}
			runtime.GC()
		})
	}
}

type benchmark struct {
	dbType              db.BackendType
	initSize, blockSize int
	keyLen, dataLen     int
}

func backendList() []db.BackendType {
	return slices.DeleteFunc(db.BackendList(), func(s db.BackendType) bool {
		// fsdb doesn't support batch ops, and it's slow anyways, so let's skip.
		return s == db.FSDBBackend
	})
}

func BenchmarkSmall(b *testing.B) {
	ls := backendList()
	bs := make([]benchmark, 0, len(ls))
	for _, backend := range ls {
		bs = append(bs, benchmark{backend, 1_000, 100, 16, 40})
	}
	runBenchmarks(b, bs)
}

func BenchmarkMedium(b *testing.B) {
	ls := backendList()
	bs := make([]benchmark, 0, len(ls))
	for _, backend := range ls {
		bs = append(bs, benchmark{backend, 100_000, 100, 16, 40})
	}
	runBenchmarks(b, bs)
}

func BenchmarkLarge(b *testing.B) {
	ls := backendList()
	bs := make([]benchmark, 0, len(ls))
	for _, backend := range ls {
		bs = append(bs, benchmark{backend, 1_000_000, 100, 16, 40})
	}
	runBenchmarks(b, bs)
}

func BenchmarkLevelDBBatchSizes(b *testing.B) {
	benchmarks := []benchmark{
		{db.GoLevelDBBackend, 100000, 5, 16, 40},
		{db.GoLevelDBBackend, 100000, 25, 16, 40},
		{db.GoLevelDBBackend, 100000, 100, 16, 40},
		{db.GoLevelDBBackend, 100000, 400, 16, 40},
		{db.GoLevelDBBackend, 100000, 2000, 16, 40},
	}
	runBenchmarks(b, benchmarks)
}

// BenchmarkLevelDBLargeData is intended to push disk limits
// in the leveldb, to make sure not everything is cached
func BenchmarkLevelDBLargeData(b *testing.B) {
	benchmarks := []benchmark{
		{db.GoLevelDBBackend, 50000, 100, 32, 100},
		{db.GoLevelDBBackend, 50000, 100, 32, 1000},
		{db.GoLevelDBBackend, 50000, 100, 32, 10000},
		{db.GoLevelDBBackend, 50000, 100, 32, 100000},
	}
	runBenchmarks(b, benchmarks)
}

func runBenchmarks(b *testing.B, benchmarks []benchmark) {
	b.Helper()

	for _, bb := range benchmarks {
		prefix := fmt.Sprintf("%s-%d-%d-%d-%d", bb.dbType, bb.initSize,
			bb.blockSize, bb.keyLen, bb.dataLen)

		// prepare a dir for the db and cleanup afterwards
		dirName := fmt.Sprintf("./%s-db", prefix)
		defer func() {
			err := os.RemoveAll(dirName)
			if err != nil {
				b.Errorf("%+v\n", err)
			}
		}()

		// note that "" leads to nil backing db!
		var d db.DB
		var err error
		if bb.dbType != "nodb" {
			d, err = db.NewDB("test", bb.dbType, dirName)
			require.NoError(b, err)
			defer d.Close()
		}
		b.Run(prefix, func(sub *testing.B) {
			runSuite(sub, d, bb.initSize, bb.blockSize, bb.keyLen, bb.dataLen)
		})
	}
}

// returns number of MB in use
func memUseMB() float64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	asize := mem.Alloc
	mb := float64(asize) / 1000000
	return mb
}

func runSuite(b *testing.B, d db.DB, initSize, blockSize, keyLen, dataLen int) {
	b.Helper()

	// measure mem usage
	runtime.GC()
	init := memUseMB()

	t, keys := prepareTree(b, d, initSize, keyLen, dataLen)
	used := memUseMB() - init
	fmt.Printf("Init Tree took %0.2f MB\n", used)

	b.ResetTimer()

	b.Run("query-miss", func(sub *testing.B) {
		runQueries(sub, t, keyLen)
	})
	b.Run("query-hits", func(sub *testing.B) {
		runKnownQueries(sub, t, keys)
	})
	b.Run("update", func(sub *testing.B) {
		t = runUpdate(sub, t, dataLen, blockSize, keys)
	})
	b.Run("block", func(sub *testing.B) {
		t = runBlock(sub, t, keyLen, dataLen, blockSize, keys)
	})

	// both of these edit size of the tree too much
	// need to run with their own tree
	// t = nil // for gc
	// b.Run("insert", func(sub *testing.B) {
	// 	it, _ := prepareTree(d, initSize, keyLen, dataLen)
	// 	sub.ResetTimer()
	// 	runInsert(sub, it, keyLen, dataLen, blockSize)
	// })
	// b.Run("delete", func(sub *testing.B) {
	// 	dt, dkeys := prepareTree(d, initSize+sub.N, keyLen, dataLen)
	// 	sub.ResetTimer()
	// 	runDelete(sub, dt, blockSize, dkeys)
	// })
}
