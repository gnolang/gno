package coverage

import (
	"go/ast"
)

// ExecutableNode represents a node that can be executed and cause side effects
type ExecutableNode interface {
	ast.Node
	IsExecutable() bool
	GetExecutionContext() ExecutionContext
}

// ExecutionContext defines the context in which a node executes
type ExecutionContext struct {
	Type                    ExecutionType
	RequiresInstrumentation bool
	Line                    int
}

// ExecutionType categorizes different types of executable nodes
type ExecutionType int

const (
	_ ExecutionType = iota

	// Block-level execution (Axiom A2: Block Entry Point)
	BlockEntry
	ConditionalBranch
	LoopEntry
	SwitchEntry
	SelectEntry
	CaseEntry

	// Statement-level execution
	AssignmentExecution
	ExpressionExecution
	ReturnExecution
	DeferExecution
	BranchExecution
)

// BranchingStrategy defines how different control flow constructs should be instrumented
type BranchingStrategy interface {
	ShouldInstrumentEntry(node ast.Node) bool
	ShouldInstrumentBranches(node ast.Node) bool
	GetBranches(node ast.Node) []ast.Node
}

// ExternalInstrumentationDetector detects if a file/node is externally instrumented
type ExternalInstrumentationDetector interface {
	IsExternallyInstrumented(node ast.Node) bool
	GetExternalMarkers() []string
}

// InstrumentationRule defines how a specific type of node should be instrumented
type InstrumentationRule interface {
	Applies(node ast.Node) bool
	Apply(node ast.Node, engine *InstrumentationEngine) error
	GetRuleName() string
}

// CoverageData represents the coverage data for a file
type CoverageData struct {
	TotalLines    int
	CoveredLines  int
	CoverageRatio float64 // ratio in [0, 100]
	LineData      map[int]int
}
