package bptree

import "fmt"

// PruneVersionsTo deletes all versions from firstVersion through toVersion
// (inclusive), removing orphaned nodes via the dual-tree-walk algorithm.
// Adapted for B+ tree fan-out: uses child hash set comparison instead of
// positional matching.
func (t *MutableTree) PruneVersionsTo(toVersion int64) error {
	first := t.ndb.getFirstVersion()
	latest := t.ndb.getLatestVersion()
	if toVersion >= latest {
		return fmt.Errorf("cannot prune latest version %d", latest)
	}
	if toVersion < first {
		// Everything at or below toVersion is already pruned. Returning early
		// also keeps setFirstVersion below from REWINDING the version floor
		// (e.g. PruneVersionsTo(5) with first=100 must not set firstVersion=6).
		return nil
	}

	// Pruning commits the shared batch (threshold flushes + the final Commit)
	// and discards it on error, so the batch must contain nothing but prune
	// deletes: refuse to run with uncommitted working-session state. This also
	// blocks two session-staging hazards a mid-session prune would trigger: a
	// threshold flush persisting staged session writes early (breaking the
	// Rollback "nothing in the DB" guarantee), and — after LoadVersion(old) +
	// Set — flushing staged writes keyed into an already-committed version's
	// value namespace, corrupting it.
	if t.root != t.lastSaved || len(t.ndb.pendingVals) > 0 ||
		t.nextValueNonce > 0 || len(t.versionOrphans) > 0 {
		return fmt.Errorf("%w: SaveVersion or Rollback before pruning", ErrUncommittedChanges)
	}

	// Claim the prune lock and verify no version in [first, toVersion] has an
	// active reader. beginPruning holds pruneMu for the whole prune so no new
	// reader can register a to-be-deleted version (H3); endPruning releases it.
	if err := t.ndb.beginPruning(first, toVersion); err != nil {
		return err
	}
	defer t.ndb.endPruning()

	if err := t.pruneRange(first, toVersion, latest); err != nil {
		// Drop partially-staged deletes so a later unrelated Commit can't flush
		// them: a half-deleted version with a live root ref would be unreadable
		// AND unprunable (its retry fails on the missing nodes, and the store's
		// pruning Commit panics on that error). Intermediate threshold commits
		// land only on whole-version boundaries, so what was flushed is
		// consistent; firstVersion is unadvanced, so a retry re-processes.
		t.ndb.DiscardBatch()
		return err
	}
	t.ndb.setFirstVersion(toVersion + 1)
	return nil
}

// pruneRange deletes versions [first, toVersion] and commits the staged
// deletes. On error, staged-but-uncommitted deletes are left in the batch for
// the caller to discard.
func (t *MutableTree) pruneRange(first, toVersion, latest int64) error {
	// Flush cap in bytes. Commit whenever the pending batch grows past this
	// bound so PruneVersionsTo's working memory is O(flushThreshold) rather
	// than O(pruned-nodes-and-values), matching typical per-block usage but
	// also bounding startup catch-up prunes of many versions.
	//
	// Intermediate commits are safe: pruning is idempotent. A crash after a
	// partial commit means some root references are already deleted from
	// the DB, so discoverVersions on the next startup recomputes the
	// correct firstVersion and a retry re-processes only what's left.
	flushThreshold := t.ndb.opts.FlushThreshold
	if flushThreshold <= 0 {
		flushThreshold = 4 * 1024 * 1024 // 4 MiB default
	}

	for v := first; v <= toVersion; v++ {
		if !t.ndb.VersionExists(v) {
			continue
		}
		// findNextVersion always returns a non-zero version here: we rejected
		// toVersion >= latest above, so latest > v and findNextVersion will
		// find at least `latest` as a successor. That is the only guarantee
		// we need — a successor must exist for dual-tree-walk pruning to
		// know which values (from orphan lists) can safely be deleted.
		nextV := t.findNextVersion(v, latest)
		if nextV == 0 {
			// Defensive: should not happen given the toVersion < latest
			// check above. If it ever does, bail rather than silently
			// deleting nodes without processing value-orphan lists (the
			// old deleteAllNodesForVersion path walked nodes but left
			// every leaf value referenced by v in the DB).
			return fmt.Errorf("bptree: pruning v%d found no successor (invariant: toVersion=%d < latest=%d should guarantee one)", v, toVersion, latest)
		}
		if err := t.pruneVersion(v, nextV); err != nil {
			return err
		}
		if err := t.ndb.DeleteRoot(v); err != nil {
			return err
		}
		// Flush if the batch has grown beyond the threshold. Ignore
		// GetByteSize errors — fall back to the final Commit which will
		// still flush everything; we just lose the intermediate bound.
		if size, err := t.ndb.batch.GetByteSize(); err == nil && size >= flushThreshold {
			if err := t.ndb.Commit(); err != nil {
				return err
			}
		}
	}

	return t.ndb.Commit()
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

	// Process orphan value lists:
	//
	// - orphans[nextV] lists values displaced WHEN nextV was created, i.e.,
	//   values whose owning version is <=v. Pruning v means those values
	//   should disappear.
	//
	// - orphans[v] lists values displaced when v was created (owners <v).
	//   Those are normally consumed during pruneVersion(v-1, v), which runs
	//   in the previous loop iteration and DeleteOrphans(v) at its end. For
	//   the very first pruned version in a batch there is no such prior
	//   iteration; without this guard, any values listed in orphans[first]
	//   would leak. LoadOrphans returns an empty slice if the record was
	//   already deleted, so for iterations where v > first the call is a
	//   no-op (no batch bloat).
	for _, version := range [2]int64{v, nextV} {
		orphans, err := t.ndb.LoadOrphans(version)
		if err != nil {
			return fmt.Errorf("loading orphans for v%d: %w", version, err)
		}
		for _, vk := range orphans {
			if err := t.ndb.DeleteValue(vk); err != nil {
				return err
			}
		}
	}

	// Clean up orphan records for both versions.
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
			return fmt.Errorf("loading old child %d: %w", i, err)
		}

		// Find the corresponding child in the new tree by routing from the
		// new tree ROOT. This handles inner node splits where children move
		// to sibling nodes not reachable from the local newNode.
		var newChild Node
		if newRoot != nil {
			newChild, err = t.findCorrespondingChild(newRoot, child)
			if err != nil {
				return err
			}
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
// Descends to the same height as oldChild. Returns (nil, nil) when no
// corresponding node exists (genuinely orphaned). A node-load failure is
// returned as an error — it must NOT be treated as "orphaned", or a transient
// read error would delete subtrees still shared with the successor version.
func (t *MutableTree) findCorrespondingChild(newNode, oldChild Node) (Node, error) {
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
		return nil, nil
	}

	targetHeight := nodeHeight(oldChild)

	// Route through the new tree to find the node at the same height
	node := newNode
	for {
		if nodeHeight(node) == targetHeight {
			return node, nil
		}
		if nodeHeight(node) < targetHeight {
			return nil, nil // new tree is shorter
		}

		inner, ok := node.(*InnerNode)
		if !ok {
			return node, nil // leaf — can't descend further
		}
		idx := searchInner(inner, key)

		// In-memory child first; else load from the DB, propagating errors
		// (getChild would panic on a load failure — prune must return it so
		// the caller can discard the partially-staged batch).
		child := inner.childNodes[idx]
		if child == nil && inner.children[idx] != nil {
			var err error
			child, err = t.ndb.GetNode(inner.children[idx])
			if err != nil {
				return nil, fmt.Errorf("routing to corresponding child: %w", err)
			}
		}
		if child == nil {
			return nil, nil
		}
		node = child
	}
}
