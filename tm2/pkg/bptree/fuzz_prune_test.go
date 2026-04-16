package bptree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// FuzzPruneConsistency drives randomized Set / Remove / SaveVersion /
// PruneVersionsTo operations against a bptree and a reference map.
//
// It is the regression harness for Finding #3 in POTENTIAL_IMPROVEMENTS.md.
// The earlier positional-descent prune algorithm could mis-identify
// subtrees when successive splits and merges restructured the new
// version's tree, causing nodes that were still live to be deleted.
// The current mark-and-sweep implementation replaced that algorithm; this
// fuzzer guards against any regression that would resurrect the bug.
// Any crash surfaces as either
//  - a `bptree: failed to load child node ...` panic from `getChild`, or
//  - a mirror-vs-tree hash/value mismatch after a SaveVersion that follows
//    a prune.
//
// Run locally with e.g.:
//
//	go test -run=^$ -fuzz=FuzzPruneConsistency -fuzztime=5m ./tm2/pkg/bptree/
//
// Crash seeds are auto-saved by `go test` under `testdata/fuzz/FuzzPruneConsistency/`
// and should be checked in.
func FuzzPruneConsistency(f *testing.F) {
	// Seed corpus — known workloads that exercise splits + merges at B=32.
	// The fuzzer will mutate these; the exact bytes are not load-bearing.
	f.Add([]byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	})
	f.Add(bytes.Repeat([]byte{0xaa, 0x55}, 32))
	f.Add(bytes.Repeat([]byte{0xff}, 64))
	// A pattern biased towards the known-bad split+prune sequence.
	f.Add([]byte{
		0x00, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70,
		0x80, 0x90, 0xa0, 0xb0, 0xc0, 0xd0, 0xe0, 0xf0,
		0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe, 0xba, 0xbe,
	})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 8 {
			t.Skip()
		}
		driveRandomOps(t, data)
	})
}

// driveRandomOps interprets `data` as an opcode stream and executes it
// against a fresh MutableTree and a mirror map, checking invariants.
func driveRandomOps(t *testing.T, data []byte) {
	t.Helper()

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 64, NewNopLogger())
	// `mirror` tracks the current in-memory state (set/remove not yet
	// committed). `committed` tracks the state as of the last SaveVersion
	// and is what a fresh `LoadVersion(lastVer)` reload should see.
	mirror := make(map[string][]byte)
	committed := make(map[string][]byte)
	var lastVer int64

	// Use `data` as both the opcode tape and the PRNG seed so the same
	// input is fully deterministic. Any crash that surfaces is reproducible.
	seed := binary.LittleEndian.Uint64(padTo8(data[:min(len(data), 8)]))
	rng := rand.New(rand.NewSource(int64(seed)))

	// Key namespace is narrow so splits and merges happen often.
	const keySpace = 256

	pos := 8
	for pos < len(data) && !t.Failed() {
		op := data[pos] & 0x7
		pos++

		switch op {
		case 0, 1, 2, 3: // Set (weighted)
			k := deriveKey(rng, keySpace)
			v := deriveValue(rng)
			if _, err := tree.Set(k, v); err != nil {
				t.Fatalf("Set(%x): %v", k, err)
			}
			mirror[string(k)] = v

		case 4: // Remove
			if len(mirror) == 0 {
				continue
			}
			k := pickMirrorKey(rng, mirror)
			if _, _, err := tree.Remove(k); err != nil {
				t.Fatalf("Remove(%x): %v", k, err)
			}
			delete(mirror, string(k))

		case 5: // SaveVersion
			_, v, err := tree.SaveVersion()
			if err != nil {
				t.Fatalf("SaveVersion: %v", err)
			}
			lastVer = v
			committed = copyMirror(mirror)
			assertMirrorMatchesTree(t, tree, mirror)

		case 6: // Prune a random surviving version (if any).
			if lastVer < 2 {
				continue
			}
			// Prune at most up to lastVer-1 (cannot prune latest).
			target := int64(1) + rng.Int63n(lastVer-1)
			if err := tree.PruneVersionsTo(target); err != nil {
				// Only ErrActiveReaders is an expected failure — no readers here.
				t.Fatalf("PruneVersionsTo(%d): %v", target, err)
			}
			assertMirrorMatchesTree(t, tree, mirror)

		case 7: // Reload via a fresh MutableTree and re-check the
			// committed state (not the live mirror, which may hold
			// uncommitted Set/Remove operations).
			if lastVer == 0 {
				continue
			}
			t2 := NewMutableTreeWithDB(db, 64, NewNopLogger())
			if _, err := t2.LoadVersion(lastVer); err != nil {
				t.Fatalf("LoadVersion(%d): %v", lastVer, err)
			}
			assertMirrorMatchesTree(t, t2, committed)
		}
	}
}

func deriveKey(rng *rand.Rand, keySpace int) []byte {
	// Short fixed-width key so splits happen within a small universe.
	k := make([]byte, 4)
	binary.BigEndian.PutUint32(k, uint32(rng.Intn(keySpace)))
	return k
}

func deriveValue(rng *rand.Rand) []byte {
	// Values are not load-bearing for the prune bug; keep them short.
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], rng.Uint64())
	return buf[:]
}

func pickMirrorKey(rng *rand.Rand, m map[string][]byte) []byte {
	// Deterministic selection: sort-like — pick by rng-indexed walk.
	// Iteration order over maps is randomised, so we take the first Nth
	// key from the first pass (n = rng mod size).
	n := rng.Intn(len(m))
	i := 0
	for k := range m {
		if i == n {
			return []byte(k)
		}
		i++
	}
	return nil // unreachable
}

// assertMirrorMatchesTree verifies that every key in the mirror resolves
// to the expected value on the tree. A future fuzz target may extend
// this with full Iterate coverage, but equality under Get is sufficient
// to catch the pruning-corruption class of bugs.
func assertMirrorMatchesTree(t *testing.T, tree *MutableTree, mirror map[string][]byte) {
	t.Helper()
	for k, v := range mirror {
		got, err := tree.Get([]byte(k))
		if err != nil {
			t.Fatalf("Get(%x): %v", []byte(k), err)
		}
		if !bytes.Equal(got, v) {
			t.Fatalf("value mismatch for key %x: got %x, want %x", []byte(k), got, v)
		}
	}
}

func copyMirror(m map[string][]byte) map[string][]byte {
	c := make(map[string][]byte, len(m))
	maps.Copy(c, m)
	return c
}

func padTo8(b []byte) []byte {
	if len(b) >= 8 {
		return b
	}
	pad := make([]byte, 8)
	copy(pad, b)
	return pad
}

// FuzzPruneDeepTree is a second fuzz target specialised for the
// cross-subtree prune-corruption regime (POTENTIAL_IMPROVEMENTS.md
// Finding #3, now resolved by the mark-and-sweep rewrite). Relative to
// FuzzPruneConsistency it:
//
//   - uses a 4096-key namespace so the tree reliably grows to 3+ levels,
//     where inner-node splits and merges actually produce cross-subtree
//     reshuffles,
//   - biases the opcode stream toward Set (75%) + SaveVersion (10%) +
//     Prune (12.5%), with only 2.5% Remove and no Reload — this matches
//     the "sustained random-insert + prune" workload under which the bug
//     was originally observed.
func FuzzPruneDeepTree(f *testing.F) {
	f.Add(uint32(1), uint32(200))
	f.Add(uint32(42), uint32(500))
	f.Add(uint32(0xdeadbeef), uint32(1000))
	f.Add(uint32(0xcafebabe), uint32(2000))

	f.Fuzz(func(t *testing.T, seed uint32, ops uint32) {
		if ops == 0 {
			t.Skip()
		}
		if ops > 20000 {
			ops = 20000 // keep each execution bounded
		}
		driveDeepTreeOps(t, int64(seed), int(ops))
	})
}

// driveDeepTreeOps runs a set-heavy, prune-heavy workload against a
// 4096-key namespace.
func driveDeepTreeOps(t *testing.T, seed int64, nOps int) {
	t.Helper()
	const keySpace = 4096

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 128, NewNopLogger())
	mirror := make(map[string][]byte)
	committed := make(map[string][]byte)
	var lastVer int64

	rng := rand.New(rand.NewSource(seed))

	for i := 0; i < nOps && !t.Failed(); i++ {
		op := rng.Intn(40)
		switch {
		case op < 30: // 75% Set
			k := deriveKey(rng, keySpace)
			v := deriveValue(rng)
			if _, err := tree.Set(k, v); err != nil {
				t.Fatalf("Set: %v", err)
			}
			mirror[string(k)] = v
		case op < 31: // 2.5% Remove
			if len(mirror) == 0 {
				continue
			}
			k := pickMirrorKey(rng, mirror)
			if _, _, err := tree.Remove(k); err != nil {
				t.Fatalf("Remove: %v", err)
			}
			delete(mirror, string(k))
		case op < 35: // 10% SaveVersion
			_, v, err := tree.SaveVersion()
			if err != nil {
				t.Fatalf("SaveVersion: %v", err)
			}
			lastVer = v
			committed = copyMirror(mirror)
			assertMirrorMatchesTree(t, tree, mirror)
		default: // 12.5% Prune
			if lastVer < 2 {
				continue
			}
			target := int64(1) + rng.Int63n(lastVer-1)
			if err := tree.PruneVersionsTo(target); err != nil {
				t.Fatalf("PruneVersionsTo(%d): %v", target, err)
			}
			// After prune, the latest version should still satisfy
			// the mirror and a fresh reload at lastVer should satisfy
			// the committed snapshot — this is the assertion that
			// would surface silent prune corruption if mark-and-sweep
			// ever incorrectly deleted a node still reachable from a
			// retained version.
			assertMirrorMatchesTree(t, tree, mirror)
			t2 := NewMutableTreeWithDB(db, 128, NewNopLogger())
			if _, err := t2.LoadVersion(lastVer); err != nil {
				t.Fatalf("post-prune LoadVersion(%d): %v", lastVer, err)
			}
			assertMirrorMatchesTree(t, t2, committed)
		}
	}
}

// TestPrune_FuzzLongRun is a deterministic, seeded long-running variant of
// the fuzz target. It complements the fuzzing by exercising a denser
// workload than `TestPrune_SustainedInsertPrune` without requiring the
// fuzz runner. It serves as a regression guard for the prune corruption
// class of bugs (Finding #3).
//
// Any new seed that reliably crashes should be promoted to the fuzz
// corpus via `testdata/fuzz/FuzzPruneConsistency/`.
func TestPrune_FuzzLongRun(t *testing.T) {
	if testing.Short() {
		t.Skip("long running; skip under -short")
	}
	// Seeds chosen to produce churn across splits+merges.
	seeds := []int64{1, 42, 0xdeadbeef, 0xbadc0ffee}
	// ~100k opcodes per seed — exercises thousands of
	// Set/Remove/SaveVersion/Prune cycles, which is where the known
	// pruning corruption (Finding #3) has historically surfaced.
	const opBytes = 100_000

	for _, seed := range seeds {
		t.Run(fmt.Sprintf("seed=%x", uint64(seed)), func(t *testing.T) {
			data := make([]byte, 8+opBytes)
			binary.LittleEndian.PutUint64(data, uint64(seed))
			rng := rand.New(rand.NewSource(seed))
			for i := 8; i < len(data); i++ {
				data[i] = byte(rng.Intn(256))
			}
			driveRandomOps(t, data)
		})
	}
}
