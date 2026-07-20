package bptree

import (
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// Benchmarks isolating the read-path cost of non-memoizing getChild.
// Run the same benchmarks on HEAD (with memoization) and after the change
// (without), then compare with benchstat. The tree fits the cache, so every
// getChild is a cache hit — this measures cache-lookup vs. pointer-follow, the
// degradation we trade for the bounded-memory win.

const benchTreeSize = 100_000

// buildBenchTree builds a tree of benchTreeSize sequential keys, commits one
// version, and returns the mutable tree plus its committed snapshot. The cache
// is sized to hold the whole tree so reads stay cache-resident.
func buildBenchTree(b *testing.B) (*MutableTree, *ImmutableTree) {
	b.Helper()
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 200_000, NewNopLogger())
	for i := range benchTreeSize {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			b.Fatal(err)
		}
	}
	_, version, err := tree.SaveVersion()
	if err != nil {
		b.Fatal(err)
	}
	imm, err := tree.GetImmutable(version)
	if err != nil {
		b.Fatal(err)
	}
	return tree, imm
}

// BenchmarkGet — random point reads on a committed snapshot. On HEAD the
// accessed subgraph is memoized on the snapshot after warmup; without
// memoization every getChild re-fetches from the cache. This is the worst case
// for the change.
func BenchmarkGet(b *testing.B) {
	_, imm := buildBenchTree(b)
	rng := rand.New(rand.NewSource(1))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := imm.Get(i2b(rng.Intn(benchTreeSize))); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIterate — full ascending scan. Every leaf transition re-descends via
// getChild.
func BenchmarkIterate(b *testing.B) {
	_, imm := buildBenchTree(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it, err := imm.Iterator(nil, nil, true)
		if err != nil {
			b.Fatal(err)
		}
		n := 0
		for ; it.Valid(); it.Next() {
			_ = it.Key()
			n++
		}
		it.Close()
		if n != benchTreeSize {
			b.Fatalf("scanned %d keys, want %d", n, benchTreeSize)
		}
	}
}

// BenchmarkProof — random membership proofs (root-to-leaf descent + sibling
// paths).
func BenchmarkProof(b *testing.B) {
	_, imm := buildBenchTree(b)
	rng := rand.New(rand.NewSource(1))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := imm.GetMembershipProof(i2b(rng.Intn(benchTreeSize))); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBlockCommit — write-heavy block: update keysPerBlock random keys and
// SaveVersion. The dirty path is memoized via setChild regardless of the
// change, so this should be largely insensitive to it (control benchmark).
func BenchmarkBlockCommit(b *testing.B) {
	const keysPerBlock = 1000
	tree, _ := buildBenchTree(b)
	rng := rand.New(rand.NewSource(1))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for range keysPerBlock {
			k := i2b(rng.Intn(benchTreeSize))
			if _, err := tree.Set(k, k); err != nil {
				b.Fatal(err)
			}
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			b.Fatal(err)
		}
	}
}
