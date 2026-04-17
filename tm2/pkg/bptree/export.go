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
type Exporter struct {
	tree      *ImmutableTree
	ndb       *nodeDB // for fetching values; nil for in-memory
	ch        chan *ExportNode
	done      chan struct{} // closed by Close() to signal goroutine to exit
	err       error
	closeOnce sync.Once
}

// Export creates an Exporter for the tree. The tree's version is protected
// from pruning via version readers.
func (t *ImmutableTree) Export(ndb *nodeDB) (*Exporter, error) {
	if t.root == nil {
		return nil, ErrNotInitializedTree
	}

	if ndb != nil {
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
		// Resolver handles external slots; inline slots are returned
		// directly by valueAt without consulting the resolver.
		resolver := e.tree.valueResolver
		if resolver == nil && e.ndb != nil {
			resolver = e.ndb.GetValue
		}
		for i := 0; i < int(n.numKeys); i++ {
			value, err := n.valueAt(i, resolver)
			if err != nil {
				return err
			}
			if err := e.send(&ExportNode{
				Key:    n.keys[i],
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
		// Emit inner node marker with ALL separator keys. Copy each key
		// instead of sharing the backing slice: a concurrent mutator on
		// the source tree (export is valid on ImmutableTree but the
		// same underlying array may be visible to a MutableTree root via
		// structural sharing before the first cowRoot) could otherwise
		// tear key bytes mid-export, and a consumer that retains the
		// ExportNode after the next Set would see corrupted keys. See
		// Finding #27.
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
