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

// Exporter streams tree nodes in depth-first post-order (children before parent).
type Exporter struct {
	tree      *ImmutableTree
	ndb       *nodeDB // for fetching values; nil for in-memory
	ch        chan *ExportNode
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
	}
	go e.run()
	return e, nil
}

func (e *Exporter) run() {
	defer close(e.ch)
	e.err = e.exportNode(e.tree.root)
}

func (e *Exporter) exportNode(node Node) error {
	switch n := node.(type) {
	case *LeafNode:
		// Emit each key-value entry
		for i := 0; i < int(n.numKeys); i++ {
			var value []byte
			if e.ndb != nil {
				var err error
				value, err = e.ndb.GetValue(n.valueHashes[i])
				if err != nil {
					return err
				}
			} else {
				vh := n.valueHashes[i]
				value = vh[:]
			}
			e.ch <- &ExportNode{
				Key:    n.keys[i],
				Value:  value,
				Height: 0,
			}
		}
		// Emit leaf boundary marker
		e.ch <- &ExportNode{
			Height:  -1,
			NumKeys: n.numKeys,
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
		// Emit inner node marker with ALL separator keys
		sepKeys := make([][]byte, n.numKeys)
		for i := 0; i < int(n.numKeys); i++ {
			sepKeys[i] = n.keys[i]
		}
		e.ch <- &ExportNode{
			Height:        int8(n.height),
			NumKeys:       n.numKeys,
			SeparatorKeys: sepKeys,
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
		if e.err != nil {
			return nil, e.err
		}
		return nil, ErrExportDone
	}
	return node, nil
}

// Close releases the exporter and decrements version readers.
// Safe to call multiple times.
func (e *Exporter) Close() {
	e.closeOnce.Do(func() {
		// Drain channel to let goroutine exit
		for range e.ch {
		}
		if e.ndb != nil {
			e.ndb.decrVersionReaders(e.tree.version)
		}
	})
}
