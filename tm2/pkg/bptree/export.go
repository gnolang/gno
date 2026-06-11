package bptree

import (
	"errors"
	"sync"
)

// ExportNode is the format for export/import streaming.
type ExportNode struct {
	Key           []byte
	Value         []byte   // actual value (inlined, not hash)
	Height        int8     // 0=leaf entry, -1=leaf boundary marker, >0=inner marker
	NumKeys       int16    // leaf boundary: keys in this leaf; inner: separator key count
	SeparatorKeys [][]byte // inner marker only: all separator keys
}

// errExportClosed is a sentinel used internally when the done channel is closed.
var errExportClosed = errors.New("export closed")

// Exporter streams tree nodes in depth-first post-order (children before parent).
//
// Lifecycle: the caller MUST Close the Exporter when done — including on error
// or early return (use defer). Close stops the streaming goroutine and releases
// the version-reader reservation taken by Export. Abandoning an Exporter without
// Close leaks that goroutine (it blocks once the 32-entry channel buffer fills,
// i.e. for any tree with >32 nodes) AND permanently pins its version against
// pruning: PruneVersionsTo of that version returns ErrActiveReaders, and the
// store's auto-prune at Commit panics on it. Fully consuming the stream (Next
// until ErrExportDone) lets the goroutine exit but does NOT release the
// reservation — Close is still required. Close is idempotent.
type Exporter struct {
	tree      *ImmutableTree
	ndb       *nodeDB // for fetching values; may be nil (then tree.valueResolver is used)
	ch        chan *ExportNode
	done      chan struct{} // closed by Close() to signal goroutine to exit
	err       error
	closeOnce sync.Once
}

// Export creates an Exporter for the tree. The tree's version is protected from
// pruning via a version reader held for the Exporter's lifetime; the caller MUST
// Close the returned Exporter (use defer) to release it — see the Exporter docs.
func (t *ImmutableTree) Export(ndb *nodeDB) (*Exporter, error) {
	if t.root == nil {
		return nil, ErrNotInitializedTree
	}

	// version > 0 matches the registration convention elsewhere (a version-0
	// entry is meaningless — nothing prunable exists below firstVersion 1);
	// Close's unconditional decrement is a no-op for an unregistered version.
	if ndb != nil && t.version > 0 {
		ndb.incrVersionReaders(t.version)
	}

	e := &Exporter{
		tree: t,
		ndb:  ndb,
		ch:   make(chan *ExportNode, 32),
		done: make(chan struct{}),
	}
	go e.run()
	return e, nil
}

func (e *Exporter) run() {
	defer close(e.ch)
	e.err = e.exportNode(e.tree.root)
}

// send sends a node on the channel, or returns errExportClosed if Close() was called.
func (e *Exporter) send(node *ExportNode) error {
	select {
	case e.ch <- node:
		return nil
	case <-e.done:
		return errExportClosed
	}
}

func (e *Exporter) exportNode(node Node) error {
	switch n := node.(type) {
	case *LeafNode:
		// Emit each key-value entry
		for i := 0; i < int(n.numKeys); i++ {
			var value []byte
			if e.ndb != nil {
				// Export always runs on a committed ImmutableTree, so resolve
				// DB-only (never the writer's pendingVals buffer): the export
				// goroutine runs concurrently with the writer and must not race
				// SaveValue's map write.
				var err error
				value, err = e.ndb.getCommittedValue(n.valueKeys[i])
				if err != nil {
					return err
				}
			} else if e.tree.valueResolver != nil {
				var err error
				value, err = e.tree.valueResolver(n.valueKeys[i])
				if err != nil {
					return err
				}
			} else {
				return errors.New("export: no value resolver available (ndb is nil and no valueResolver set)")
			}
			if err := e.send(&ExportNode{
				// Copy: the consumer owns the ExportNode; the raw key slice
				// belongs to a live leaf shared with the tree and cache.
				Key:    copyKey(n.keys[i]),
				Value:  value,
				Height: 0,
			}); err != nil {
				return err
			}
		}
		// Emit leaf boundary marker
		if err := e.send(&ExportNode{
			Height:  -1,
			NumKeys: n.numKeys,
		}); err != nil {
			return err
		}
		return nil

	case *InnerNode:
		// Recurse children first (depth-first post-order)
		for i := 0; i < n.NumChildren(); i++ {
			child := n.getChild(i)
			if child != nil {
				if err := e.exportNode(child); err != nil {
					return err
				}
			}
		}
		// Emit inner node marker with ALL separator keys (copies — the
		// consumer owns the ExportNode; the raw slices belong to live nodes).
		sepKeys := make([][]byte, n.numKeys)
		for i := 0; i < int(n.numKeys); i++ {
			sepKeys[i] = copyKey(n.keys[i])
		}
		if err := e.send(&ExportNode{
			Height:        int8(n.height),
			NumKeys:       n.numKeys,
			SeparatorKeys: sepKeys,
		}); err != nil {
			return err
		}
		return nil

	default:
		return errors.New("unknown node type in export")
	}
}

// Next returns the next ExportNode, or ErrExportDone when finished.
func (e *Exporter) Next() (*ExportNode, error) {
	node, ok := <-e.ch
	if !ok {
		if e.err != nil && e.err != errExportClosed {
			return nil, e.err
		}
		return nil, ErrExportDone
	}
	return node, nil
}

// Close releases the exporter and decrements version readers.
// Safe to call multiple times. The goroutine exits promptly without
// needing to drain the channel.
func (e *Exporter) Close() {
	e.closeOnce.Do(func() {
		close(e.done)
		// Drain channel to let goroutine exit if it's blocked on send
		for range e.ch {
		}
		if e.ndb != nil {
			e.ndb.decrVersionReaders(e.tree.version)
		}
	})
}
