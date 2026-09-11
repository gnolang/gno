package bptree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// importKV holds a buffered key-value entry during import.
type importKV struct {
	key       []byte
	valueHash Hash
	valueKey  []byte
}

// importEntry is a built subtree on the importer's stack, carrying the
// smallest and largest key in that subtree so separator windows can be
// validated when the parent marker arrives (post-order delivery guarantees
// the full subtree is known before its parent).
type importEntry struct {
	node   Node
	minKey []byte
	maxKey []byte
}

// importerState is the Importer lifecycle. Add/Commit are legal only while
// active; any Commit failure poisons the importer (a retry would re-stage
// nodes but NOT the values the failed SaveVersion discarded, committing a
// tree whose root hash matches the trusted app hash while every value record
// is missing). Honest callers recover by re-importing from scratch.
type importerState int

const (
	importerActive importerState = iota
	importerCommitted
	importerClosed
	importerFailed
)

func (s importerState) String() string {
	switch s {
	case importerActive:
		return "active"
	case importerCommitted:
		return "committed"
	case importerClosed:
		return "closed"
	case importerFailed:
		return "failed"
	}
	return "unknown"
}

// Importer reconstructs a tree from a stream of ExportNodes,
// preserving the exact tree structure (and thus the root hash).
//
// The stream is structurally validated as it arrives: leaf keys must be
// strictly ascending across the whole stream, every leaf marker must drain
// the entry buffer exactly, separators must sit in the window
// (max(left child) < sep <= min(right child)), and inner heights must equal
// derived child height + 1. Combined with the caller's final root-hash check
// against a trusted app hash (which covers keys, value hashes, and shape), a
// malicious stream cannot produce a tree that mis-routes reads. The one
// residual freedom — a separator moved WITHIN its window under an unchanged
// root hash — does not affect reads, iteration, or pruning; its first
// consequence is an app-hash mismatch at the first write that routes
// differently, which halts the node loudly.
type Importer struct {
	tree        *MutableTree
	version     int64
	kvBuffer    []importKV
	stack       []importEntry
	nextNonce   uint32
	lastLeafKey []byte
	state       importerState
}

// Import creates an Importer that will reconstruct a tree at the given version.
//
// The target version must be beyond the latest committed version: importing
// into the live key namespace would overwrite node/value records shared with
// retained versions (silent corruption). Import then rolls back any
// uncommitted working-session state (pending batch, orphan list, value-nonce
// counter) so stale state from a prior un-committed session cannot leak into
// the import's SaveVersion. Import rebuilds the root from scratch, so
// reverting t.root to lastSaved here is harmless.
func (t *MutableTree) Import(version int64) (*Importer, error) {
	if t.VersionExists(version) {
		return nil, fmt.Errorf("version %d already exists", version)
	}
	if latest := t.ndb.getLatestVersion(); version <= latest {
		return nil, fmt.Errorf("import: version %d must exceed latest version %d", version, latest)
	}
	t.Rollback()
	// Importer.Commit bypasses per-entry index maintenance, so pre-existing
	// 'F' entries (stamped ≤ the import version) would be trusted
	// stale-present after Commit. Drop the whole index now, while the batch
	// is empty (just rolled back); see dropFastIndex for the abort-safety
	// ordering.
	if err := t.ndb.dropFastIndex(); err != nil {
		return nil, err
	}
	return &Importer{tree: t, version: version}, nil
}

// Add adds an ExportNode to the tree being imported.
// Nodes must arrive in depth-first post-order as produced by the Exporter.
// A validation failure leaves the importer's stack and buffer untouched, but
// values staged by earlier Adds remain staged until Close (which rolls the
// session back) — a rejected stream must be Closed, not Committed.
func (imp *Importer) Add(node *ExportNode) error {
	if imp.state != importerActive {
		return fmt.Errorf("import: Add on %s importer", imp.state)
	}
	switch {
	case node.Height == 0:
		// Leaf entry: validate, compute value hash, allocate valueKey, save
		// value, buffer. All validation precedes the nonce/value staging so a
		// rejected entry consumes nothing.
		if len(node.Key) == 0 {
			return fmt.Errorf("import: empty leaf key")
		}
		if len(node.Key) > MaxKeyLen {
			return fmt.Errorf("import: leaf key length %d exceeds MaxKeyLen %d", len(node.Key), MaxKeyLen)
		}
		if imp.lastLeafKey != nil && bytes.Compare(node.Key, imp.lastLeafKey) <= 0 {
			return fmt.Errorf("import: leaf key %x not strictly greater than previous key %x (stream must be sorted and duplicate-free)",
				node.Key, imp.lastLeafKey)
		}
		keyCopy := append([]byte(nil), node.Key...)
		valueHash := sha256.Sum256(node.Value)
		vk := (&NodeKey{Version: imp.version, Nonce: imp.nextNonce}).GetKey()
		imp.nextNonce++
		if err := imp.tree.ndb.SaveValue(node.Value, vk); err != nil {
			// An I/O failure (unlike a validation rejection) leaves the
			// stream un-resumable: the entry was never buffered, so a
			// continued stream would either trip the exact-drain check or,
			// if this leaf was the whole stream, commit an empty tree.
			// Poison so Commit refuses; the caller re-imports.
			imp.state = importerFailed
			return err
		}
		imp.kvBuffer = append(imp.kvBuffer, importKV{
			key:       keyCopy,
			valueHash: valueHash,
			valueKey:  vk,
		})
		imp.lastLeafKey = keyCopy
		return nil

	case node.Height == -1:
		// Leaf boundary marker: drain the buffered entries, build LeafNode.
		nk := int(node.NumKeys)
		// A legitimate Exporter never emits a zero-key marker (an empty tree
		// fails Export; non-root minima are enforced); a zero-key SAVED node
		// would also break the prune's first-key routing. Reject.
		if nk < 1 || nk > B {
			return fmt.Errorf("import: leaf numKeys %d out of range [1,%d]", nk, B)
		}
		// Exact drain: the exporter emits a leaf's entries immediately
		// followed by its marker, so leftovers mean a malformed/regrouped
		// stream (which could otherwise smuggle entries across leaf
		// boundaries until the separator window caught it later).
		if len(imp.kvBuffer) != nk {
			return fmt.Errorf("import: leaf boundary expects exactly %d buffered entries, have %d", nk, len(imp.kvBuffer))
		}
		entries := imp.kvBuffer
		leaf := &LeafNode{
			numKeys:  node.NumKeys,
			miniTree: NewMiniMerkle(),
		}
		for i := range nk {
			leaf.keys[i] = entries[i].key
			leaf.valueHashes[i] = entries[i].valueHash
			leaf.valueKeys[i] = entries[i].valueKey
		}
		leaf.RebuildMiniMerkle()
		imp.kvBuffer = imp.kvBuffer[:0]
		imp.stack = append(imp.stack, importEntry{
			node:   leaf,
			minKey: leaf.keys[0],
			maxKey: leaf.keys[nk-1],
		})
		return nil

	case node.Height > 0:
		// Inner node marker: validate, then pop NumKeys+1 children from the
		// stack and build the InnerNode. Every check precedes the stack
		// mutation so a rejected marker leaves the importer state unchanged.
		if node.NumKeys < 1 || node.NumKeys > B-1 {
			// NumKeys==0 (a single-child inner) is rejected: this package never
			// saves one (single-child roots collapse), and a zero-key node
			// breaks the prune's first-key routing.
			return fmt.Errorf("import: inner numKeys %d out of range [1,%d]", node.NumKeys, B-1)
		}
		numChildren := int(node.NumKeys) + 1
		if len(imp.stack) < numChildren {
			return fmt.Errorf("import: inner marker expects %d children, stack has %d", numChildren, len(imp.stack))
		}
		if len(node.SeparatorKeys) != int(node.NumKeys) {
			return fmt.Errorf("import: inner marker has %d separator keys, expected %d", len(node.SeparatorKeys), node.NumKeys)
		}
		for i, sk := range node.SeparatorKeys {
			if len(sk) == 0 {
				return fmt.Errorf("import: empty inner separator key %d", i)
			}
			if len(sk) > MaxKeyLen {
				return fmt.Errorf("import: inner separator key %d length %d exceeds MaxKeyLen %d", i, len(sk), MaxKeyLen)
			}
		}
		children := imp.stack[len(imp.stack)-numChildren:]

		// Heights: derive from the built children (uniform-depth B+ tree) and
		// require the stream's claim to match — a disagreeing exporter is
		// confused and gets a loud error rather than a silent "fix". The
		// height field routes cross-version pruning, so trusting the stream
		// would let a bogus value mis-route a later prune.
		childHeight := nodeHeight(children[0].node)
		for i := 1; i < numChildren; i++ {
			if h := nodeHeight(children[i].node); h != childHeight {
				return fmt.Errorf("import: inner marker children have non-uniform heights (%d and %d)", childHeight, h)
			}
		}
		if derived := childHeight + 1; int16(node.Height) != derived {
			return fmt.Errorf("import: inner marker height %d disagrees with derived child height+1 = %d", node.Height, derived)
		}

		// Separator windows: max(left subtree) < sep <= min(right subtree) —
		// the exact invariant honest trees satisfy (equality is common at
		// splits; strict-less arises when deletions raise the right subtree's
		// min; violations mis-route searches).
		for i, sk := range node.SeparatorKeys {
			left, right := children[i], children[i+1]
			if bytes.Compare(left.maxKey, sk) >= 0 || bytes.Compare(sk, right.minKey) > 0 {
				return fmt.Errorf("import: separator %d (%x) outside window: max(left)=%x < sep <= min(right)=%x violated",
					i, sk, left.maxKey, right.minKey)
			}
		}

		inner := &InnerNode{
			numKeys:  node.NumKeys,
			height:   int16(node.Height),
			miniTree: NewMiniMerkle(),
		}

		// Set separator keys
		for i := 0; i < int(node.NumKeys); i++ {
			inner.keys[i] = append([]byte(nil), node.SeparatorKeys[i]...)
		}

		// Set children: compute childSizes and childHashes
		for i := range numChildren {
			child := children[i].node
			inner.childNodes[i] = child
			inner.childHashes[i] = child.Hash()
			inner.childSizes[i] = nodeSize(child)
		}
		inner.RebuildMiniMerkle()

		entry := importEntry{
			node:   inner,
			minKey: children[0].minKey,
			maxKey: children[numChildren-1].maxKey,
		}
		imp.stack = imp.stack[:len(imp.stack)-numChildren]
		imp.stack = append(imp.stack, entry)
		return nil

	default:
		return fmt.Errorf("import: unexpected Height %d", node.Height)
	}
}

// Commit finalizes the import by saving the version.
//
// The caller wiring state-sync MUST compare the committed version's hash
// against the consensus-trusted app hash: the structural checks in Add cover
// what the hash cannot (separator windows, heights), and the hash covers
// what the checks cannot (keys, value hashes, shape). An empty stream
// commits an empty tree — accepted for tests, though no honest Exporter can
// produce it (Export fails on an empty tree), so a state-sync caller should
// treat it like any other hash mismatch.
func (imp *Importer) Commit() error {
	if imp.state != importerActive {
		return fmt.Errorf("import: Commit on %s importer", imp.state)
	}
	if len(imp.kvBuffer) > 0 {
		imp.state = importerFailed
		return fmt.Errorf("import: %d unbounded leaf entries remaining", len(imp.kvBuffer))
	}

	switch len(imp.stack) {
	case 0:
		// Empty tree
		imp.tree.root = nil
		imp.tree.size = 0
	case 1:
		imp.tree.root = imp.stack[0].node
		imp.tree.size = nodeSize(imp.stack[0].node)
	default:
		imp.state = importerFailed
		return fmt.Errorf("import: expected 1 root on stack, have %d", len(imp.stack))
	}

	// Set version so SaveVersion uses the target version, clearing
	// initialVersion to avoid the WorkingVersion() special case — and restore
	// BOTH on failure: nothing else does (Rollback restores only root/size),
	// and leaving them mutated would make the next honest SaveVersion commit
	// old content under the import-target version number.
	prevVersion, prevInitial := imp.tree.version, imp.tree.initialVersion
	imp.tree.version = imp.version - 1
	imp.tree.initialVersion = 0
	// Import staged values via SaveValue directly, bypassing per-entry
	// fast-index maintenance; Import() already cleared the index and its
	// stamp. Suppress SaveVersion's completeness stamp for this one commit
	// (toggle off, restore after) so the next Load rebuilds the index from
	// the imported tree rather than stamping the cleared one as complete.
	// No-op when the fast index is already disabled.
	prevFast := imp.tree.ndb.opts.FastIndex
	imp.tree.ndb.opts.FastIndex = false
	_, _, err := imp.tree.SaveVersion()
	imp.tree.ndb.opts.FastIndex = prevFast
	if err != nil {
		imp.tree.version, imp.tree.initialVersion = prevVersion, prevInitial
		imp.state = importerFailed
		return err
	}
	imp.state = importerCommitted
	return nil
}

// Close releases the importer. If the import was not committed (abandoned,
// rejected stream, or failed Commit), the staged session state — value
// records Add wrote into the shared batch, and any root/version mutation a
// failed Commit left behind — is rolled back; otherwise those value records
// would ride into the NEXT commit as unreferenced, unreclaimable records in a
// future version's namespace. Idempotent; safe after Commit.
func (imp *Importer) Close() error {
	switch imp.state {
	case importerActive, importerFailed:
		imp.tree.Rollback()
		imp.state = importerClosed
	}
	return nil
}
