package bptree

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// recomputeHash recomputes node's hash from scratch, bottom-up, ignoring every
// cached mini-merkle in the tree. For leaves it re-hashes each (key,
// valueHash) slot; for inners it recomputes each child's hash recursively
// rather than trusting childHashes.
func recomputeHash(t *testing.T, ndb *nodeDB, node Node) Hash {
	t.Helper()
	switch n := node.(type) {
	case *LeafNode:
		var m MiniMerkle
		m.Clear()
		for i := range int(n.numKeys) {
			m.tree[B+i] = HashLeafSlotFromValueHash(n.keys[i], n.valueHashes[i])
		}
		m.Build()
		return m.Root()
	case *InnerNode:
		var m MiniMerkle
		m.Clear()
		for i := range n.NumChildren() {
			child := n.childNodes[i]
			if child == nil {
				if n.children[i] == nil {
					t.Fatalf("inner child %d has neither in-memory node nor ref", i)
				}
				var err error
				child, err = ndb.GetNode(n.children[i])
				if err != nil {
					t.Fatalf("loading child %d: %v", i, err)
				}
			}
			m.tree[B+i] = recomputeHash(t, ndb, child)
		}
		m.Build()
		return m.Root()
	default:
		t.Fatalf("unknown node type %T", node)
		return Hash{}
	}
}

// checkMiniMerkleConsistent asserts that every node reachable from node has a
// cached mini-merkle root equal to the one recomputed from scratch, and that
// each inner's childHashes agree with its children's own cached hashes.
func checkMiniMerkleConsistent(t *testing.T, ndb *nodeDB, node Node, path string) {
	t.Helper()
	if want, got := recomputeHash(t, ndb, node), node.Hash(); want != got {
		t.Fatalf("%s: cached hash %x != recomputed %x", path, got, want)
	}
	inner, ok := node.(*InnerNode)
	if !ok {
		return
	}
	for i := range inner.NumChildren() {
		child := inner.childNodes[i]
		if child == nil {
			continue // not loaded in this session; its ref/hash are durable
		}
		if inner.childHashes[i] != child.Hash() {
			t.Fatalf("%s/%d: childHashes %x != child.Hash() %x", path, i, inner.childHashes[i], child.Hash())
		}
		if got := inner.miniTree.GetSlot(i); got != child.Hash() {
			t.Fatalf("%s/%d: mini-merkle slot %x != child.Hash() %x", path, i, got, child.Hash())
		}
		checkMiniMerkleConsistent(t, ndb, child, fmt.Sprintf("%s/%d", path, i))
	}
}

// TestMiniMerkleInvariant_RandomMutations is the safety net for the
// incremental mini-merkle maintenance in insert.go/remove.go: every mutating
// path must leave the node's cached mini-merkle equal to a from-scratch
// rebuild. saveNode used to paper over any gap with an unconditional
// RebuildMiniMerkle on every dirty node; this test is what makes dropping
// that affordable.
func TestMiniMerkleInvariant_RandomMutations(t *testing.T) {
	for _, seed := range []int64{1, 2, 3, 12345} {
		t.Run(fmt.Sprintf("seed-%d", seed), func(t *testing.T) {
			r := rand.New(rand.NewSource(seed))
			db := memdb.NewMemDB()
			tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
			defer tree.Close()

			live := map[string][]byte{}
			for round := range 40 {
				ops := 1 + r.Intn(40)
				for range ops {
					if len(live) > 0 && r.Intn(100) < 35 {
						// Remove or overwrite an existing key.
						var pick string
						for k := range live {
							pick = k
							break
						}
						if r.Intn(2) == 0 {
							if _, _, err := tree.Remove([]byte(pick)); err != nil {
								t.Fatal(err)
							}
							delete(live, pick)
							continue
						}
						v := make([]byte, 1+r.Intn(50))
						r.Read(v)
						if _, err := tree.Set([]byte(pick), v); err != nil {
							t.Fatal(err)
						}
						live[pick] = v
						continue
					}
					k := make([]byte, 1+r.Intn(12))
					r.Read(k)
					v := make([]byte, 1+r.Intn(50))
					r.Read(v)
					if _, err := tree.Set(k, v); err != nil {
						t.Fatal(err)
					}
					live[string(k)] = v
				}
				if tree.root != nil {
					checkMiniMerkleConsistent(t, tree.ndb, tree.root, fmt.Sprintf("round%d/root", round))
				}
				workingHash := append([]byte(nil), tree.WorkingHash()...)
				savedHash, ver, err := tree.SaveVersion()
				if err != nil {
					t.Fatal(err)
				}
				if string(workingHash) != string(savedHash) {
					t.Fatalf("round %d: WorkingHash %x != SaveVersion hash %x", round, workingHash, savedHash)
				}
				// Reload from disk: the reloaded tree rebuilds every
				// mini-merkle from the serialized inputs, so a stale cached
				// hash on the write path shows up as a mismatch here.
				reloaded := NewMutableTreeWithDB(db, 0, NewNopLogger())
				if _, err := reloaded.LoadVersion(ver); err != nil {
					t.Fatal(err)
				}
				if string(reloaded.Hash()) != string(savedHash) {
					t.Fatalf("round %d: reloaded hash %x != saved %x", round, reloaded.Hash(), savedHash)
				}
				for k, want := range live {
					got, err := reloaded.Get([]byte(k))
					if err != nil {
						t.Fatal(err)
					}
					if string(got) != string(want) {
						t.Fatalf("round %d: key %x = %x, want %x", round, k, got, want)
					}
				}
				if got := reloaded.Size(); got != int64(len(live)) {
					t.Fatalf("round %d: size %d, want %d", round, got, len(live))
				}
				reloaded.Close()
			}
		})
	}
}
