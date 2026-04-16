package bptree

import (
	"encoding/binary"
	"fmt"
)

// nodeKeyArr encodes a *NodeKey into a value-typed [NodeKeySize]byte
// suitable for use as a map key. Avoids the per-call []byte allocation
// in nk.GetKey() on the pruning hot path.
func nodeKeyArr(nk *NodeKey) [NodeKeySize]byte {
	var a [NodeKeySize]byte
	binary.BigEndian.PutUint64(a[:8], uint64(nk.Version))
	binary.BigEndian.PutUint32(a[8:], nk.Nonce)
	return a
}

// nodeKeyBytesToArr copies an already-serialized 12-byte NodeKey into a
// value-typed [NodeKeySize]byte for map lookup. A short slice would
// otherwise zero-pad the tail and silently miscompare against the
// reachable set (producing either a bogus hit or a bogus miss). A short
// slice reaching this helper indicates a desync between `InnerNode`
// serialisation and the reader (Finding #18 also guards this on the
// write side); panic with a clear diagnostic so the upstream bug is
// visible instead of corrupting prune decisions. See Finding #47.
func nodeKeyBytesToArr(b []byte) [NodeKeySize]byte {
	if len(b) != NodeKeySize {
		panic(fmt.Sprintf("bptree: nodeKeyBytesToArr: short slice len=%d want %d", len(b), NodeKeySize))
	}
	var a [NodeKeySize]byte
	copy(a[:], b)
	return a
}

// PruneVersionsTo deletes all versions from firstVersion through toVersion
// (inclusive) via a mark-and-sweep reachability pass.
//
// Algorithm:
//  1. Build a reachability set over all retained versions (toVersion+1..latest),
//     recording every live NodeKey.
//  2. For each pruned version v, walk v's tree and delete any node whose
//     NodeKey is NOT in the reachability set. Subtrees whose root is
//     shared with a retained version are skipped without descent
//     (NodeKeys are unique per save and nodes are immutable, so a shared
//     root implies a shared subtree).
//  3. Process each version's orphan list, deleting values displaced in the
//     transition to nextV.
//  4. Delete the pruned version's root record.
//
// The mark-and-sweep approach is correct under arbitrary split/merge
// restructurings of the B+ tree. See POTENTIAL_IMPROVEMENTS.md Finding #3.
func (t *MutableTree) PruneVersionsTo(toVersion int64) error {
	if t.ndb == nil {
		return nil
	}

	first := t.ndb.getFirstVersion()
	latest := t.ndb.getLatestVersion()
	if toVersion >= latest {
		return fmt.Errorf("cannot prune latest version %d", latest)
	}

	// Atomically claim the prune: verify no readers on [first, toVersion]
	// AND block new registrations for the duration of the prune. This
	// closes the check-vs-delete TOCTOU. See Finding #15.
	if err := t.ndb.beginPruning(first, toVersion); err != nil {
		return err
	}
	defer t.ndb.endPruning()

	// Build the NodeKey reachability set ONCE from all retained versions
	// (toVersion+1..latest). Sweeping then costs O(per-version divergence)
	// rather than O(retained-tree × versions) as a naive per-pair mark
	// would.
	reachable, err := t.buildRetainedReachableSet(toVersion+1, latest)
	if err != nil {
		return err
	}

	for v := first; v <= toVersion; v++ {
		if !t.ndb.VersionExists(v) {
			continue
		}
		nextV := t.findNextVersion(v, latest)
		if nextV == 0 {
			// No next version — just delete root ref and nodes
			if err := t.deleteAllNodesForVersion(v); err != nil {
				t.ndb.discardBatch() // Finding #52
				return err
			}
		} else {
			if err := t.sweepAndOrphanVersion(v, nextV, reachable); err != nil {
				t.ndb.discardBatch() // Finding #52
				return err
			}
		}
		if err := t.ndb.DeleteRoot(v); err != nil {
			t.ndb.discardBatch() // Finding #52
			return err
		}
	}

	if err := t.ndb.Commit(); err != nil {
		return err
	}
	// Advance-only: if a caller passes toVersion < firstVersion-1 (the
	// per-version loop body is a no-op in that case), we must not rewind
	// the bookkeeping below the true first retained version. See
	// Finding #41.
	if next := toVersion + 1; next > t.ndb.getFirstVersion() {
		t.ndb.setFirstVersion(next)
	}
	return nil
}

// findNextVersion finds the next existing version after v, up to latest.
func (t *MutableTree) findNextVersion(v, latest int64) int64 {
	for nv := v + 1; nv <= latest; nv++ {
		if t.ndb.VersionExists(nv) {
			return nv
		}
	}
	return 0
}

// buildRetainedReachableSet walks every retained version in [from, to]
// and records the NodeKey of every reachable node into a set. The
// resulting set is the "must-not-delete" mask for the sweep phase:
// any NodeKey not in the set is dead (not reachable from any retained
// version) and can be deleted safely.
//
// Building the set once per PruneVersionsTo call amortises the mark
// cost across all pruned versions; in the common case where the only
// retained version is `latest`, this is a single tree walk.
func (t *MutableTree) buildRetainedReachableSet(from, to int64) (map[[NodeKeySize]byte]struct{}, error) {
	reachable := make(map[[NodeKeySize]byte]struct{})
	for rv := from; rv <= to; rv++ {
		if !t.ndb.VersionExists(rv) {
			continue
		}
		rvRootNK, _, err := t.ndb.GetRoot(rv)
		if err != nil {
			return nil, fmt.Errorf("loading retained v%d root: %w", rv, err)
		}
		if rvRootNK == nil {
			continue
		}
		rvRoot, err := t.ndb.GetNode(rvRootNK)
		if err != nil {
			return nil, fmt.Errorf("loading retained v%d root node: %w", rv, err)
		}
		if err := t.markReachable(rvRoot, reachable); err != nil {
			return nil, err
		}
	}
	return reachable, nil
}

// sweepAndOrphanVersion performs the sweep and orphan-processing for a
// single old version v using a pre-built reachable set.
//
// When v has no tree (vRootNK is nil), the sweep phase is skipped but
// the orphan-processing block still runs — values displaced when nextV
// was created are stored under nextV's orphan list and must be deleted
// regardless of whether v itself contained any nodes. See Finding #2.
func (t *MutableTree) sweepAndOrphanVersion(v, nextV int64, reachable map[[NodeKeySize]byte]struct{}) error {
	vRootNK, _, err := t.ndb.GetRoot(v)
	if err != nil {
		return err
	}
	if vRootNK != nil {
		vRoot, err := t.ndb.GetNode(vRootNK)
		if err != nil {
			return fmt.Errorf("loading v%d root: %w", v, err)
		}
		if err := t.sweepOld(vRoot, reachable); err != nil {
			return err
		}
	}

	// Values displaced when nextV was created are deleted now.
	orphans, err := t.ndb.LoadOrphans(nextV)
	if err != nil {
		return fmt.Errorf("loading orphans for v%d: %w", nextV, err)
	}
	for _, vk := range orphans {
		if err := t.ndb.DeleteValue(vk); err != nil {
			return err
		}
	}
	if err := t.ndb.DeleteOrphans(nextV); err != nil {
		return err
	}
	if err := t.ndb.DeleteOrphans(v); err != nil {
		return err
	}
	return nil
}

// markReachable records the NodeKey of every node reachable from node
// into reachable. Loads children lazily through the nodeDB.
//
// Two optimisations avoid unnecessary DB work:
//
//   - Already-marked short-circuit: if a node's NodeKey is already in
//     the set (possible when marking across multiple retained versions
//     that share structure), the subtree is already fully recorded —
//     NodeKeys are uniquely assigned per SaveVersion and nodes are
//     immutable after save.
//
//   - Leaf-skip: when the parent is directly above the leaf level
//     (height == 1), children are leaves with no descendants. Their
//     NodeKeys can be marked from the parent's `children[i]` reference
//     without a DB load + RebuildMiniMerkle. At B=32, every inner node
//     at height 1 owns up to 32 leaf children and the inner itself is
//     one node, so leaves outnumber parents-of-leaves 32:1 and the
//     leaf-skip covers 32/33 ≈ 97% of descents at the penultimate
//     level. See Finding #51.
//
// The leaf-skip correctness depends on the invariant that a height-1
// inner's children are all leaves. The first child is verified on
// entry (Finding #46): if it deserialises as an InnerNode, the tree
// is corrupt and prune must bail rather than silently leak subtrees.
func (t *MutableTree) markReachable(node Node, reachable map[[NodeKeySize]byte]struct{}) error {
	if node == nil {
		return nil
	}
	nk := node.GetNodeKey()
	if nk != nil {
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
	if err := t.assertLeafChildren(inner, leafChildren); err != nil {
		return err
	}
	for i := 0; i < inner.NumChildren(); i++ {
		if inner.childNodes[i] != nil {
			if err := t.markReachable(inner.childNodes[i], reachable); err != nil {
				return err
			}
			continue
		}
		if inner.children[i] == nil {
			continue
		}
		arr := nodeKeyBytesToArr(inner.children[i])
		if _, seen := reachable[arr]; seen {
			continue
		}
		if leafChildren {
			reachable[arr] = struct{}{}
			continue
		}
		child, err := t.ndb.GetNode(inner.children[i])
		if err != nil {
			return fmt.Errorf("markReachable: loading child %d: %w", i, err)
		}
		if err := t.markReachable(child, reachable); err != nil {
			return err
		}
	}
	return nil
}

// assertLeafChildren verifies that inner.height == 1 implies its children
// are LeafNodes (and height > 1 implies InnerNodes). The check peeks at
// the first available child — either the in-memory reference or the
// first serialised ref — which suffices because height invariants are
// uniform across all children of a given inner node. One DB load per
// height-1 parent is paid at most once per prune pass; the cost is
// amortised by the leaf-skip optimisation that follows. See Finding
// #46.
func (t *MutableTree) assertLeafChildren(inner *InnerNode, leafChildren bool) error {
	for i := 0; i < inner.NumChildren(); i++ {
		var probe Node
		if inner.childNodes[i] != nil {
			probe = inner.childNodes[i]
		} else if inner.children[i] != nil {
			c, err := t.ndb.GetNode(inner.children[i])
			if err != nil {
				return fmt.Errorf("assertLeafChildren: loading child %d: %w", i, err)
			}
			probe = c
		} else {
			continue
		}
		_, isLeaf := probe.(*LeafNode)
		if leafChildren != isLeaf {
			return fmt.Errorf("%w: height=%d child[%d] isLeaf=%v",
				ErrHeightInvariantViolated, inner.height, i, isLeaf)
		}
		return nil
	}
	return nil
}

// sweepOld walks the old-version tree and deletes each node whose
// NodeKey is not in reachable. If a node's NodeKey IS in reachable, the
// entire subtree rooted at it is known to be shared with a retained
// version — NodeKeys are uniquely assigned per SaveVersion and nodes are
// immutable after save, so every descendant is also in reachable — and
// the descent is skipped.
//
// A parent-of-leaves optimisation mirrors markReachable: orphan leaves
// are deleted by NodeKey alone without a DB load.
func (t *MutableTree) sweepOld(node Node, reachable map[[NodeKeySize]byte]struct{}) error {
	if node == nil {
		return nil
	}
	nk := node.GetNodeKey()
	if nk != nil {
		if _, shared := reachable[nodeKeyArr(nk)]; shared {
			return nil
		}
		if err := t.ndb.DeleteNode(nk.GetKey()); err != nil {
			return err
		}
	}
	inner, ok := node.(*InnerNode)
	if !ok {
		return nil
	}
	leafChildren := inner.height == 1
	if err := t.assertLeafChildren(inner, leafChildren); err != nil {
		return err
	}
	for i := 0; i < inner.NumChildren(); i++ {
		var child Node
		if inner.childNodes[i] != nil {
			// In-memory child. Mirror the serialised-ref fast path:
			// if the child's NodeKey is already in the reachable set
			// the whole subtree is shared and descent is pointless.
			// Otherwise fall through to sweepOld which rewalks this
			// subtree and its descendants. See Finding #49.
			if ck := inner.childNodes[i].GetNodeKey(); ck != nil {
				if _, shared := reachable[nodeKeyArr(ck)]; shared {
					continue
				}
			}
			child = inner.childNodes[i]
		} else if inner.children[i] != nil {
			if _, shared := reachable[nodeKeyBytesToArr(inner.children[i])]; shared {
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
				return fmt.Errorf("sweepOld: loading child %d: %w", i, err)
			}
			child = c
		}
		if child == nil {
			continue
		}
		if err := t.sweepOld(child, reachable); err != nil {
			return err
		}
	}
	return nil
}

// deleteAllNodesForVersion deletes the root node and all nodes reachable
// from it, and processes v's own orphan list.
//
// Today the only caller path in PruneVersionsTo gates this on
// `toVersion >= latest` at L59, so `findNextVersion(v, latest) == 0`
// implies `v == latest` and the orphan-list of v is for keys displaced
// when v+1 was created — a version that by definition does not exist
// under the current guard. Loading and deleting v's orphans here
// defends against any future relaxation of that guard that would
// otherwise silently leak values. See Finding #43.
func (t *MutableTree) deleteAllNodesForVersion(v int64) error {
	nkBytes, _, err := t.ndb.GetRoot(v)
	if err != nil {
		return err
	}
	if nkBytes != nil {
		root, err := t.ndb.GetNode(nkBytes)
		if err != nil {
			return fmt.Errorf("loading root node for v%d: %w", v, err)
		}
		if err := t.deleteSubtree(root); err != nil {
			return err
		}
	}
	orphans, err := t.ndb.LoadOrphans(v)
	if err != nil {
		return fmt.Errorf("loading orphans for v%d: %w", v, err)
	}
	for _, vk := range orphans {
		if err := t.ndb.DeleteValue(vk); err != nil {
			return err
		}
	}
	return t.ndb.DeleteOrphans(v)
}

// deleteSubtree recursively deletes a node and all its descendants.
//
// Child resolution prefers the in-memory reference (childNodes[i]) when
// set, falling back to loading from the serialised ref (children[i]).
// Today the only caller (deleteAllNodesForVersion) passes a freshly
// DB-loaded root whose childNodes begin nil, so the fast path never
// fires — but a future caller handing in a tree with unsaved children
// would leak those subtrees without this check. See Finding #50.
func (t *MutableTree) deleteSubtree(node Node) error {
	nk := node.GetNodeKey()
	if nk != nil {
		if err := t.ndb.DeleteNode(nk.GetKey()); err != nil {
			return err
		}
	}

	if inner, ok := node.(*InnerNode); ok {
		for i := 0; i < inner.NumChildren(); i++ {
			if inner.childNodes[i] != nil {
				if err := t.deleteSubtree(inner.childNodes[i]); err != nil {
					return err
				}
				continue
			}
			if inner.children[i] != nil {
				child, err := t.ndb.GetNode(inner.children[i])
				if err != nil {
					return fmt.Errorf("loading child %d in deleteSubtree: %w", i, err)
				}
				if err := t.deleteSubtree(child); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
