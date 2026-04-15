package bptree

import "fmt"

// PruneVersionsTo deletes all versions from firstVersion through toVersion
// (inclusive), removing orphaned nodes via the dual-tree-walk algorithm.
// Adapted for B+ tree fan-out: uses child hash set comparison instead of
// positional matching.
func (t *MutableTree) PruneVersionsTo(toVersion int64) error {
	if t.ndb == nil {
		return nil
	}

	first := t.ndb.getFirstVersion()
	latest := t.ndb.getLatestVersion()
	if toVersion >= latest {
		return fmt.Errorf("cannot prune latest version %d", latest)
	}

	// Check for active readers
	for v := first; v <= toVersion; v++ {
		if t.ndb.hasVersionReaders(v) {
			return fmt.Errorf("%w: version %d", ErrActiveReaders, v)
		}
	}

	for v := first; v <= toVersion; v++ {
		if !t.ndb.VersionExists(v) {
			continue
		}
		nextV := t.findNextVersion(v, latest)
		if nextV == 0 {
			// No next version — just delete root ref and nodes
			if err := t.deleteAllNodesForVersion(v); err != nil {
				return err
			}
		} else {
			// Dual-tree-walk: find orphaned nodes in version v
			// that are not referenced by version nextV.
			if err := t.pruneVersion(v, nextV); err != nil {
				return err
			}
		}
		if err := t.ndb.DeleteRoot(v); err != nil {
			return err
		}
	}

	if err := t.ndb.Commit(); err != nil {
		return err
	}
	t.ndb.setFirstVersion(toVersion + 1)
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

// pruneVersion performs the dual-tree-walk between version v and nextV,
// deleting nodes in v that are not referenced by nextV.
func (t *MutableTree) pruneVersion(v, nextV int64) error {
	// Load roots for both versions
	vRootNK, _, err := t.ndb.GetRoot(v)
	if err != nil {
		return err
	}
	nextRootNK, _, err := t.ndb.GetRoot(nextV)
	if err != nil {
		return err
	}

	if vRootNK == nil {
		// Empty tree at version v — nothing to prune
		return nil
	}

	vRoot, err := t.ndb.GetNode(vRootNK)
	if err != nil {
		return fmt.Errorf("loading v%d root: %w", v, err)
	}

	var nextRoot Node
	if nextRootNK != nil {
		nextRoot, err = t.ndb.GetNode(nextRootNK)
		if err != nil {
			return fmt.Errorf("loading v%d root: %w", nextV, err)
		}
	}

	if err := t.walkAndPrune(vRoot, nextRoot, nextRoot); err != nil {
		return err
	}

	// Process orphan list: delete values displaced when nextV was created
	orphans, err := t.ndb.LoadOrphans(nextV)
	if err != nil {
		return fmt.Errorf("loading orphans for v%d: %w", nextV, err)
	}
	for _, vk := range orphans {
		if err := t.ndb.DeleteValue(vk); err != nil {
			return err
		}
	}
	// Clean up orphan records for both versions
	if err := t.ndb.DeleteOrphans(nextV); err != nil {
		return err
	}
	if err := t.ndb.DeleteOrphans(v); err != nil {
		return err
	}

	return nil
}

// walkAndPrune compares two subtrees and deletes nodes from oldNode
// that are not referenced by newNode. Uses child hash set comparison
// to handle split/merge position shifts.
//
// newRoot is the root of the new version's tree, used to find children
// that moved to a different part of the tree due to inner node splits.
func (t *MutableTree) walkAndPrune(oldNode, newNode, newRoot Node) error {
	if oldNode == nil {
		return nil
	}

	// If both nodes have the same hash, the entire subtree is shared — skip.
	if newNode != nil && oldNode.Hash() == newNode.Hash() {
		return nil
	}

	// Delete the old node itself (it's been replaced or removed in the new version)
	oldNK := oldNode.GetNodeKey()
	if oldNK != nil {
		if err := t.ndb.DeleteNode(oldNK.GetKey()); err != nil {
			return err
		}
	}

	// For inner nodes, recurse into children
	oldInner, isInner := oldNode.(*InnerNode)
	if !isInner {
		return nil // leaf — already deleted above, no children
	}

	// Build a set of child hashes from the new node (if it's also an inner node)
	newChildHashes := make(map[Hash]bool)
	if newInner, ok := newNode.(*InnerNode); ok {
		for i := 0; i < newInner.NumChildren(); i++ {
			newChildHashes[newInner.childHashes[i]] = true
		}
	}

	// For each child in the old inner node:
	// - If the child hash exists in the new node's children → shared, skip
	// - If not → the child subtree may be orphaned or moved, check from root
	for i := 0; i < oldInner.NumChildren(); i++ {
		childHash := oldInner.childHashes[i]
		if newChildHashes[childHash] {
			continue // shared subtree
		}

		// Load and recurse into the orphaned child
		if oldInner.children[i] == nil {
			continue // no serialized ref (shouldn't happen for saved nodes)
		}
		child, err := t.ndb.GetNode(oldInner.children[i])
		if err != nil {
			// Node may have been deleted in a previous iteration — skip
			continue
		}

		// Find the corresponding child in the new tree by routing from the
		// new tree ROOT. This handles inner node splits where children move
		// to sibling nodes not reachable from the local newNode.
		var newChild Node
		if newRoot != nil {
			newChild = t.findCorrespondingChild(newRoot, child)
		}

		// If the child was found in the new tree with the same hash,
		// it's shared (moved to a different part of the tree due to split) — skip.
		if newChild != nil && child.Hash() == newChild.Hash() {
			continue
		}

		if err := t.walkAndPrune(child, newChild, newRoot); err != nil {
			return err
		}
	}

	return nil
}

// findCorrespondingChild finds the child in newNode's tree that covers
// the same key range as oldChild. Uses the first key of oldChild to route.
// Descends to the same height as oldChild.
func (t *MutableTree) findCorrespondingChild(newNode, oldChild Node) Node {
	// Get a representative key from the old child
	var key []byte
	switch c := oldChild.(type) {
	case *LeafNode:
		if c.numKeys > 0 {
			key = c.keys[0]
		}
	case *InnerNode:
		if c.numKeys > 0 {
			key = c.keys[0]
		}
	}
	if key == nil {
		return nil
	}

	targetHeight := nodeHeight(oldChild)

	// Route through the new tree to find the node at the same height
	node := newNode
	for {
		if nodeHeight(node) == targetHeight {
			return node
		}
		if nodeHeight(node) < targetHeight {
			return nil // new tree is shorter
		}

		inner, ok := node.(*InnerNode)
		if !ok {
			return node // leaf — can't descend further
		}
		idx := searchInner(inner, key)

		// Try in-memory child first
		child := inner.getChild(idx)
		if child == nil && inner.children[idx] != nil {
			// Load from DB
			var err error
			child, err = t.ndb.GetNode(inner.children[idx])
			if err != nil {
				return nil
			}
		}
		if child == nil {
			return nil
		}
		node = child
	}
}

// deleteAllNodesForVersion deletes the root node and all nodes reachable
// from it. Used when there is no next version to compare against.
func (t *MutableTree) deleteAllNodesForVersion(v int64) error {
	nkBytes, _, err := t.ndb.GetRoot(v)
	if err != nil || nkBytes == nil {
		return err
	}
	root, err := t.ndb.GetNode(nkBytes)
	if err != nil {
		return nil // node may already be deleted
	}
	return t.deleteSubtree(root)
}

// deleteSubtree recursively deletes a node and all its descendants.
func (t *MutableTree) deleteSubtree(node Node) error {
	nk := node.GetNodeKey()
	if nk != nil {
		if err := t.ndb.DeleteNode(nk.GetKey()); err != nil {
			return err
		}
	}

	if inner, ok := node.(*InnerNode); ok {
		for i := 0; i < inner.NumChildren(); i++ {
			if inner.children[i] != nil {
				child, err := t.ndb.GetNode(inner.children[i])
				if err != nil {
					continue // may be already deleted (shared with another version)
				}
				if err := t.deleteSubtree(child); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
