package bptree

import (
	"crypto/sha256"
	"fmt"
)

// importKV holds a buffered key-value entry during import.
type importKV struct {
	key       []byte
	valueHash Hash
	valueKey  []byte
}

// Importer reconstructs a tree from a stream of ExportNodes,
// preserving the exact tree structure (and thus the root hash).
type Importer struct {
	tree       *MutableTree
	version    int64
	kvBuffer   []importKV
	stack      []Node
	nextNonce  uint32
}

// Import creates an Importer that will reconstruct a tree at the given version.
func (t *MutableTree) Import(version int64) (*Importer, error) {
	if t.VersionExists(version) {
		return nil, fmt.Errorf("version %d already exists", version)
	}
	// nonce=0 is reserved to avoid collision with the "missing" sentinel
	// in LeafNode.Serialize (12 zero bytes). See Finding #6.
	return &Importer{tree: t, version: version, nextNonce: 1}, nil
}

// Add adds an ExportNode to the tree being imported.
// Nodes must arrive in depth-first post-order as produced by the Exporter.
func (imp *Importer) Add(node *ExportNode) error {
	switch {
	case node.Height == 0:
		// Leaf entry: compute value hash, allocate valueKey, save value, buffer.
		valueHash := sha256.Sum256(node.Value)
		vk := encodeNodeKeyBytes(imp.version, imp.nextNonce)
		imp.nextNonce++
		if imp.tree.ndb != nil {
			if err := imp.tree.ndb.SaveValue(node.Value, vk); err != nil {
				return err
			}
		} else if imp.tree.memValues != nil {
			valCopy := make([]byte, len(node.Value))
			copy(valCopy, node.Value)
			imp.tree.memValues[string(vk)] = valCopy
		}
		imp.kvBuffer = append(imp.kvBuffer, importKV{
			key:       append([]byte(nil), node.Key...),
			valueHash: valueHash,
			valueKey:  vk,
		})
		return nil

	case node.Height == -1:
		// Leaf boundary marker: pop NumKeys entries from kvBuffer, build LeafNode.
		nk := int(node.NumKeys)
		if nk < 0 || nk > B {
			return fmt.Errorf("import: leaf numKeys %d out of range [0,%d]", nk, B)
		}
		if len(imp.kvBuffer) < nk {
			return fmt.Errorf("import: leaf boundary expects %d entries, have %d", nk, len(imp.kvBuffer))
		}
		entries := imp.kvBuffer[len(imp.kvBuffer)-nk:]
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
		imp.kvBuffer = imp.kvBuffer[:len(imp.kvBuffer)-nk]
		imp.stack = append(imp.stack, leaf)
		return nil

	case node.Height > 0:
		// Inner node marker: pop NumKeys+1 children from stack, build InnerNode.
		if node.NumKeys < 0 || node.NumKeys > B-1 {
			return fmt.Errorf("import: inner numKeys %d out of range [0,%d]", node.NumKeys, B-1)
		}
		numChildren := int(node.NumKeys) + 1
		if len(imp.stack) < numChildren {
			return fmt.Errorf("import: inner marker expects %d children, stack has %d", numChildren, len(imp.stack))
		}
		if len(node.SeparatorKeys) != int(node.NumKeys) {
			return fmt.Errorf("import: inner marker has %d separator keys, expected %d", len(node.SeparatorKeys), node.NumKeys)
		}

		children := imp.stack[len(imp.stack)-numChildren:]
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
			child := children[i]
			inner.childNodes[i] = child
			inner.childHashes[i] = child.Hash()
			inner.childSizes[i] = nodeSize(child)
		}
		inner.RebuildMiniMerkle()

		imp.stack = imp.stack[:len(imp.stack)-numChildren]
		imp.stack = append(imp.stack, inner)
		return nil

	default:
		return fmt.Errorf("import: unexpected Height %d", node.Height)
	}
}

// Commit finalizes the import by saving the version.
func (imp *Importer) Commit() error {
	if len(imp.kvBuffer) > 0 {
		return fmt.Errorf("import: %d unbounded leaf entries remaining", len(imp.kvBuffer))
	}

	switch len(imp.stack) {
	case 0:
		// Empty tree
		imp.tree.root = nil
		imp.tree.size = 0
	case 1:
		imp.tree.root = imp.stack[0]
		imp.tree.size = nodeSize(imp.stack[0])
	default:
		return fmt.Errorf("import: expected 1 root on stack, have %d", len(imp.stack))
	}

	// Set version so SaveVersion uses the target version.
	// Clear initialVersion to avoid the WorkingVersion() special case.
	imp.tree.version = imp.version - 1
	imp.tree.initialVersion = 0
	_, _, err := imp.tree.SaveVersion()
	return err
}

// Close is a no-op for cleanup compatibility.
func (imp *Importer) Close() error {
	return nil
}
