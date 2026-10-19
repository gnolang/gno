package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin Machine.releaseBlock's pooling contract directly: a plain
// runtime scope block is pooled (zeroed, pointer-identically reused), and
// each documented exclusion actually excludes. Without them the exclusion
// list lives only in releaseBlock's comment, and the filetests that ride on
// recycling (defer_block_recycle.gno et al.) would stay green even if
// pooling silently stopped happening.

// poolSource returns a runtime-scope BlockNode (a BlockStmt) with numNames
// names, none of them heap items — the source shape releaseBlock pools.
func poolSource(numNames int) *BlockStmt {
	src := &BlockStmt{}
	src.NumNames = uint16(numNames)
	src.HeapItems = make([]bool, numNames)
	return src
}

// poolableBlock returns a block satisfying every releaseBlock pooling
// condition; each exclusion test below perturbs exactly one of them.
func poolableBlock(m *Machine) *Block {
	return m.Alloc.newPooledBlock(poolSource(2), nil)
}

// Positive control for the exclusion tests: a runtime scope block IS pooled —
// zeroed on release, and handed back pointer-identically by the next acquire.
func TestReleaseBlockPoolsAndReusesRuntimeBlock(t *testing.T) {
	m := NewMachineWithOptions(MachineOptions{})
	defer m.Release()

	b := poolableBlock(m)
	b.Parent = &Block{}
	b.Values[0] = typedString("sentinel")

	m.releaseBlock(b)
	require.Len(t, m.blockPool, 1, "a plain runtime scope block must be pooled")
	assert.Same(t, b, m.blockPool[0])

	// Zeroed: the pooled block retains no references.
	assert.Nil(t, b.Source)
	assert.Nil(t, b.Parent)
	assert.Empty(t, b.Values)
	require.Equal(t, blockPoolValueCap, cap(b.Values))
	for i, tv := range b.Values[:cap(b.Values)] {
		assert.True(t, tv.IsUndefined(), "backing slot %d must be cleared", i)
	}

	// Reused: acquire returns the same *Block, re-initialized for the new scope.
	src := poolSource(3)
	b2 := m.acquireBlock(src, nil)
	assert.Same(t, b, b2)
	assert.Same(t, src, b2.Source)
	assert.Len(t, b2.Values, 3)
	assert.Empty(t, m.blockPool)
}

func TestReleaseBlockSkipsWhenPoolFull(t *testing.T) {
	m := NewMachineWithOptions(MachineOptions{})
	defer m.Release()
	for range blockPoolLimit {
		m.blockPool = append(m.blockPool, &Block{Values: make([]TypedValue, 0, blockPoolValueCap)})
	}

	b := poolableBlock(m)
	m.releaseBlock(b)
	assert.Len(t, m.blockPool, blockPoolLimit, "a full pool must drop the block")
	assert.NotSame(t, b, m.blockPool[blockPoolLimit-1])
}

func TestReleaseBlockSkipsWhilePanicking(t *testing.T) {
	m := NewMachineWithOptions(MachineOptions{})
	defer m.Release()
	m.Exception = &Exception{}

	m.releaseBlock(poolableBlock(m))
	assert.Empty(t, m.blockPool, "blocks discarded during panic unwinding must not be pooled")
}

// Blocks whose Values capacity is not exactly blockPoolValueCap are dropped:
// pooling an oversized backing array would pin it (and its tail slots) for
// the machine's lifetime while serving at most blockPoolValueCap slots.
func TestReleaseBlockSkipsNonUniformCap(t *testing.T) {
	m := NewMachineWithOptions(MachineOptions{})
	defer m.Release()

	// Oversized: more names than the uniform capacity.
	over := m.Alloc.newPooledBlock(poolSource(blockPoolValueCap+1), nil)
	require.Greater(t, cap(over.Values), blockPoolValueCap)
	m.releaseBlock(over)
	assert.Empty(t, m.blockPool, "oversized block must not be pooled")

	// Undersized: a block not allocated through the pool sizing.
	under := &Block{Source: poolSource(0), Values: make([]TypedValue, 0, blockPoolValueCap-1)}
	m.releaseBlock(under)
	assert.Empty(t, m.blockPool, "undersized block must not be pooled")
}

// Long-lived blocks that travel on the block stack but are not runtime scope
// blocks — store-loaded (RefNode), file and package blocks — must never be
// pooled, nor may a node-owned static block.
func TestReleaseBlockSkipsNonRuntimeSources(t *testing.T) {
	m := NewMachineWithOptions(MachineOptions{})
	defer m.Release()

	for _, tc := range []struct {
		name string
		src  BlockNode
	}{
		{"nil source", nil},
		{"RefNode", RefNode{}},
		{"FileNode", &FileNode{}},
		{"PackageNode", &PackageNode{}},
	} {
		b := &Block{Source: tc.src, Values: make([]TypedValue, 0, blockPoolValueCap)}
		m.releaseBlock(b)
		assert.Empty(t, m.blockPool, "%s block must not be pooled", tc.name)
	}

	// A node-owned static block: Source's static block IS the block itself.
	src := poolSource(0)
	sb := src.GetStaticBlock().GetBlock()
	sb.Source = src
	sb.Values = make([]TypedValue, 0, blockPoolValueCap)
	m.releaseBlock(sb)
	assert.Empty(t, m.blockPool, "node-owned static block must not be pooled")
}

// Realm-attached blocks — already persisted (finalized ObjectID) or pending
// an ObjectID at finalize (new-real) — are insurance-excluded from the pool.
func TestReleaseBlockSkipsRealmAttachedBlocks(t *testing.T) {
	m := NewMachineWithOptions(MachineOptions{})
	defer m.Release()

	finalized := poolableBlock(m)
	finalized.ObjectInfo.ID.NewTime = 1
	require.True(t, finalized.ObjectInfo.ID.IsFinalized())
	m.releaseBlock(finalized)
	assert.Empty(t, m.blockPool, "finalized (persisted) block must not be pooled")

	newReal := poolableBlock(m)
	newReal.SetIsNewReal(true)
	m.releaseBlock(newReal)
	assert.Empty(t, m.blockPool, "new-real block must not be pooled")
}
