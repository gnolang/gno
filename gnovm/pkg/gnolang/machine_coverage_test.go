package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCoverageTracker for testing
type MockCoverageTracker struct {
	enabled        bool
	executionCalls []ExecutionCall
}

type ExecutionCall struct {
	pkgPath string
	file    string
	line    int
}

func (m *MockCoverageTracker) TrackExecution(pkgPath, file string, line int) {
	m.executionCalls = append(m.executionCalls, ExecutionCall{
		pkgPath: pkgPath,
		file:    file,
		line:    line,
	})
}

func (m *MockCoverageTracker) TrackStatement(stmt Stmt)  {}
func (m *MockCoverageTracker) TrackExpression(expr Expr) {}
func (m *MockCoverageTracker) IsEnabled() bool           { return m.enabled }
func (m *MockCoverageTracker) SetEnabled(enabled bool)   { m.enabled = enabled }

// create a minimal machine with coverage tracking
func setupMachineWithCoverage(t *testing.T) (*Machine, *MockCoverageTracker) {
	t.Helper()
	store := NewStore(nil, nil, nil)
	m := NewMachine("test", store)

	mockTracker := &MockCoverageTracker{enabled: true}
	m.CoverageTracker = mockTracker

	// Setup a package
	pkg := &PackageValue{
		PkgPath: "test/pkg",
	}
	m.Package = pkg

	return m, mockTracker
}

// Test node implementation for testing
type testNode struct {
	line int
	span Span
}

func (n *testNode) assertNode()                              {}
func (n *testNode) String() string                           { return "testNode" }
func (n *testNode) Copy() Node                               { return &testNode{line: n.line, span: n.span} }
func (n *testNode) GetPos() Pos                              { return n.span.Pos }
func (n *testNode) GetLine() int                             { return n.line }
func (n *testNode) GetColumn() int                           { return n.span.Pos.Column }
func (n *testNode) GetSpan() Span                            { return n.span }
func (n *testNode) SetSpan(span Span)                        { n.span = span }
func (n *testNode) GetLabel() Name                           { return "" }
func (n *testNode) SetLabel(Name)                            {}
func (n *testNode) HasAttribute(key GnoAttribute) bool       { return false }
func (n *testNode) GetAttribute(key GnoAttribute) any        { return nil }
func (n *testNode) SetAttribute(key GnoAttribute, value any) {}
func (n *testNode) DelAttribute(key GnoAttribute)            {}

func TestTrackCoverageForNode_NilNode(t *testing.T) {
	m, mockTracker := setupMachineWithCoverage(t)

	// Should not panic with nil node
	m.trackCoverageForNode(nil)

	assert.Empty(t, mockTracker.executionCalls, "No execution should be tracked for nil node")
}

func TestTrackCoverageForNode_InvalidLine(t *testing.T) {
	m, mockTracker := setupMachineWithCoverage(t)

	node := &testNode{line: 0}
	m.trackCoverageForNode(node)

	assert.Empty(t, mockTracker.executionCalls, "No execution should be tracked for line 0")

	node.line = -1
	m.trackCoverageForNode(node)

	assert.Empty(t, mockTracker.executionCalls, "No execution should be tracked for negative line")
}

func TestTrackCoverageForNode_NoPackage(t *testing.T) {
	m, mockTracker := setupMachineWithCoverage(t)
	m.Package = nil

	node := &testNode{line: 10}
	m.trackCoverageForNode(node)

	assert.Empty(t, mockTracker.executionCalls, "No execution should be tracked without package")
}

func TestTrackCoverageForNode_NoBlock(t *testing.T) {
	// Create machine without using setupMachineWithCoverage to avoid default blocks
	store := NewStore(nil, nil, nil)
	m := NewMachineWithOptions(MachineOptions{
		Store:       store,
		SkipPackage: true, // Skip package setup to avoid creating blocks
	})

	mockTracker := &MockCoverageTracker{enabled: true}
	m.CoverageTracker = mockTracker

	// Setup package without creating blocks
	pkg := &PackageValue{
		PkgPath: "test/pkg",
	}
	m.Package = pkg

	// Machine without blocks
	node := &testNode{line: 10}
	m.trackCoverageForNode(node)

	assert.Empty(t, mockTracker.executionCalls, "No execution should be tracked without blocks")
}

func TestTrackCoverageForNode_Success(t *testing.T) {
	m, mockTracker := setupMachineWithCoverage(t)

	// Create a proper block with source location
	loc := Location{
		PkgPath: "test/pkg",
		File:    "test.go",
	}

	// Create a file node as source
	fileNode := &FileNode{
		Attributes: Attributes{
			Span: Span{
				Pos: Pos{Line: 1, Column: 1},
				End: Pos{Line: 100, Column: 1},
			},
		},
	}

	// Create and push a block
	fileNode.Location = loc
	alloc := NewAllocator(1024)
	block := NewBlock(alloc, fileNode, nil)
	m.PushBlock(block)

	// Track coverage for a node
	node := &testNode{line: 42}
	m.trackCoverageForNode(node)

	// Verify execution was tracked
	require.Len(t, mockTracker.executionCalls, 1)
	assert.Equal(t, "test/pkg", mockTracker.executionCalls[0].pkgPath)
	assert.Equal(t, "test.go", mockTracker.executionCalls[0].file)
	assert.Equal(t, 42, mockTracker.executionCalls[0].line)
}

func TestTrackCoverageForNode_WithExprAndStmt(t *testing.T) {
	m, mockTracker := setupMachineWithCoverage(t)

	loc := Location{
		PkgPath: "test/pkg",
		File:    "expr_stmt.go",
	}

	fileNode := &FileNode{
		Attributes: Attributes{
			Span: Span{
				Pos: Pos{Line: 1, Column: 1},
				End: Pos{Line: 100, Column: 1},
			},
		},
	}

	fileNode.Location = loc
	alloc := NewAllocator(1024)
	block := NewBlock(alloc, fileNode, nil)
	m.PushBlock(block)

	// Test with an actual expression node
	expr := &BasicLitExpr{
		Attributes: Attributes{
			Span: Span{
				Pos: Pos{Line: 25, Column: 1},
				End: Pos{Line: 25, Column: 10},
			},
		},
		Kind:  INT,
		Value: "42",
	}

	m.trackCoverageForNode(expr)

	// Test with an actual statement node
	stmt := &ExprStmt{
		Attributes: Attributes{
			Span: Span{
				Pos: Pos{Line: 30, Column: 1},
				End: Pos{Line: 30, Column: 20},
			},
		},
		X: expr,
	}

	m.trackCoverageForNode(stmt)

	require.Len(t, mockTracker.executionCalls, 2)
	assert.Equal(t, 25, mockTracker.executionCalls[0].line)
	assert.Equal(t, 30, mockTracker.executionCalls[1].line)
}
