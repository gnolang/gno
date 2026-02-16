package gnolang

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// CoverBlock represents a basic block of code for coverage tracking.
// It mirrors Go's cover.CoverBlock structure for compatibility.
type CoverBlock struct {
	File  string // qualified file path (pkgPath/fileName)
	Line0 int    // start line
	Col0  int    // start column
	Line1 int    // end line
	Col1  int    // end column
	Stmts int    // number of statements in this block
}

// CoverageCollector tracks which statements have been executed during
// test runs. It uses statement pointer identity for O(1) hit tracking
// at runtime, avoiding the need to resolve file names during execution.
type CoverageCollector struct {
	// stmtBlocks maps statement pointers to their coverage block index.
	stmtBlocks map[Stmt]int
	// blocks is the ordered list of all registered coverage blocks.
	blocks []CoverBlock
	// hits tracks execution counts per block index.
	hits []uint64
}

// NewCoverageCollector creates a new coverage collector.
func NewCoverageCollector() *CoverageCollector {
	return &CoverageCollector{
		stmtBlocks: make(map[Stmt]int),
	}
}

// RegisterBlock registers a coverage block for a statement.
func (c *CoverageCollector) RegisterBlock(stmt Stmt, file string, line0, col0, line1, col1, stmts int) {
	if _, exists := c.stmtBlocks[stmt]; exists {
		return
	}
	idx := len(c.blocks)
	c.blocks = append(c.blocks, CoverBlock{
		File:  file,
		Line0: line0,
		Col0:  col0,
		Line1: line1,
		Col1:  col1,
		Stmts: stmts,
	})
	c.hits = append(c.hits, 0)
	c.stmtBlocks[stmt] = idx
}

// HitStmt records that the given statement was executed.
// This is the hot path called from doOpExec; it must be fast.
func (c *CoverageCollector) HitStmt(stmt Stmt) {
	if idx, ok := c.stmtBlocks[stmt]; ok {
		c.hits[idx]++
	}
}

// Percentage returns the percentage of statements covered.
func (c *CoverageCollector) Percentage() float64 {
	if len(c.blocks) == 0 {
		return 0
	}
	var covered, total int
	for i, block := range c.blocks {
		total += block.Stmts
		if c.hits[i] > 0 {
			covered += block.Stmts
		}
	}
	if total == 0 {
		return 0
	}
	return float64(covered) / float64(total) * 100
}

// WriteCoverProfile writes coverage data in Go's cover profile format.
// The format is compatible with `go tool cover -html`.
func (c *CoverageCollector) WriteCoverProfile(w io.Writer, mode string) {
	fmt.Fprintf(w, "mode: %s\n", mode)
	// Collect and sort entries by file then position for deterministic output.
	type entry struct {
		block CoverBlock
		count uint64
	}
	entries := make([]entry, len(c.blocks))
	for i := range c.blocks {
		entries[i] = entry{c.blocks[i], c.hits[i]}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].block.File != entries[j].block.File {
			return entries[i].block.File < entries[j].block.File
		}
		if entries[i].block.Line0 != entries[j].block.Line0 {
			return entries[i].block.Line0 < entries[j].block.Line0
		}
		return entries[i].block.Col0 < entries[j].block.Col0
	})
	for _, e := range entries {
		fmt.Fprintf(w, "%s:%d.%d,%d.%d %d %d\n",
			e.block.File, e.block.Line0, e.block.Col0,
			e.block.Line1, e.block.Col1,
			e.block.Stmts, e.count)
	}
}

// RegisterFileBlocks walks a FileNode's declarations and registers
// coverage blocks for all executable statements. This should be called
// before test execution for each source file in the package under test.
// Only non-test source files should be registered (not *_test.gno).
func (c *CoverageCollector) RegisterFileBlocks(pkgPath, fileName string, fn *FileNode) {
	qualifiedFile := pkgPath + "/" + fileName
	for _, decl := range fn.Decls {
		fd, ok := decl.(*FuncDecl)
		if !ok {
			continue
		}
		if fd.Body == nil {
			continue
		}
		c.registerBodyBlocks(qualifiedFile, fd.Body)
	}
}

// registerBodyBlocks recursively registers coverage blocks for a list of statements.
func (c *CoverageCollector) registerBodyBlocks(file string, body Body) {
	for _, stmt := range body {
		c.registerStmtBlock(file, stmt)
	}
}

// registerStmtBlock registers a coverage block for a single statement
// and recurses into sub-bodies (if/for/switch/etc).
func (c *CoverageCollector) registerStmtBlock(file string, stmt Stmt) {
	span := stmt.GetSpan()
	if span.IsZero() {
		return
	}
	// Skip bodyStmt (internal VM construct, not user code).
	if _, ok := stmt.(*bodyStmt); ok {
		return
	}
	// Register this statement as a 1-statement block.
	c.RegisterBlock(stmt, file, span.Line, span.Column, span.End.Line, span.End.Column, 1)

	// Recurse into compound statements to register inner blocks.
	switch s := stmt.(type) {
	case *IfStmt:
		c.registerBodyBlocks(file, s.Then.Body)
		c.registerBodyBlocks(file, s.Else.Body)
	case *ForStmt:
		c.registerBodyBlocks(file, s.Body)
	case *RangeStmt:
		c.registerBodyBlocks(file, s.Body)
	case *BlockStmt:
		c.registerBodyBlocks(file, s.Body)
	case *SwitchStmt:
		for i := range s.Clauses {
			c.registerBodyBlocks(file, s.Clauses[i].Body)
		}
	case *SelectStmt:
		for i := range s.Cases {
			c.registerBodyBlocks(file, s.Cases[i].Body)
		}
	}
}

// Reset clears hit counts but keeps block registrations.
func (c *CoverageCollector) Reset() {
	for i := range c.hits {
		c.hits[i] = 0
	}
}

// Merge adds hits from another collector into this one.
func (c *CoverageCollector) Merge(other *CoverageCollector) {
	// Merge by matching block positions (file + line/col).
	type blockID struct {
		file  string
		line0 int
		col0  int
		line1 int
		col1  int
	}
	myBlocks := make(map[blockID]int, len(c.blocks))
	for i, b := range c.blocks {
		myBlocks[blockID{b.File, b.Line0, b.Col0, b.Line1, b.Col1}] = i
	}
	for i, b := range other.blocks {
		bid := blockID{b.File, b.Line0, b.Col0, b.Line1, b.Col1}
		if myIdx, ok := myBlocks[bid]; ok {
			c.hits[myIdx] += other.hits[i]
		}
	}
}

// blockKey returns the canonical key for a coverage block (used for profile output).
func blockKey(file string, line0, col0, line1, col1 int) string {
	return fmt.Sprintf("%s:%d.%d,%d.%d", file, line0, col0, line1, col1)
}

// HasBlocks returns true if any coverage blocks have been registered.
func (c *CoverageCollector) HasBlocks() bool {
	return len(c.blocks) > 0
}

// String returns a summary string like "coverage: 75.1% of statements".
func (c *CoverageCollector) String() string {
	if !c.HasBlocks() {
		return "coverage: [no statements]"
	}
	return fmt.Sprintf("coverage: %.1f%% of statements", c.Percentage())
}

// FilterByPackage returns a percentage only counting blocks from the given package path prefix.
func (c *CoverageCollector) FilterByPackage(pkgPath string) float64 {
	prefix := pkgPath + "/"
	var covered, total int
	for i, block := range c.blocks {
		if !strings.HasPrefix(block.File, prefix) {
			continue
		}
		total += block.Stmts
		if c.hits[i] > 0 {
			covered += block.Stmts
		}
	}
	if total == 0 {
		return 0
	}
	return float64(covered) / float64(total) * 100
}
