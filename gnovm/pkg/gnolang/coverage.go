package gnolang

import (
	"fmt"
	"io"
	"sort"
	"sync"
)

// CoverageMode represents how coverage data is collected.
type CoverageMode string

const (
	// CoverageModeSet tracks whether each block was executed (boolean).
	CoverageModeSet CoverageMode = "set"
	// CoverageModeCount tracks how many times each block was executed.
	CoverageModeCount CoverageMode = "count"
	// CoverageModeAtomic is like count but uses atomic operations (for concurrent tests).
	CoverageModeAtomic CoverageMode = "atomic"
)

// CoverBlock represents a basic block of code that can be covered.
// This mirrors Go's cover tool format.
type CoverBlock struct {
	File      string // filename
	StartLine int    // starting line number
	StartCol  int    // starting column number
	EndLine   int    // ending line number
	EndCol    int    // ending column number
	NumStmts  int    // number of statements in this block
	Count     int64  // execution count (atomic for CoverageModeAtomic)
}

// String returns a Go cover profile compatible string representation.
// Format: "file:startLine.startCol,endLine.endCol numStmts count"
func (b *CoverBlock) String() string {
	return fmt.Sprintf("%s:%d.%d,%d.%d %d %d",
		b.File, b.StartLine, b.StartCol, b.EndLine, b.EndCol, b.NumStmts, b.Count)
}

// CoverageData tracks code coverage for a package during test execution.
type CoverageData struct {
	mu       sync.RWMutex
	Mode     CoverageMode
	PkgPath  string
	Blocks   []*CoverBlock          // all coverable blocks
	BlockMap map[string]*CoverBlock // keyed by "file:line:col" for fast lookup
}

// NewCoverageData creates a new CoverageData instance for the given package.
func NewCoverageData(pkgPath string, mode CoverageMode) *CoverageData {
	if mode == "" {
		mode = CoverageModeSet
	}
	return &CoverageData{
		Mode:     mode,
		PkgPath:  pkgPath,
		Blocks:   make([]*CoverBlock, 0),
		BlockMap: make(map[string]*CoverBlock),
	}
}

// blockKey generates a unique key for a code location.
func blockKey(file string, line, col int) string {
	return fmt.Sprintf("%s:%d:%d", file, line, col)
}

// RegisterBlock registers a coverable code block.
func (cd *CoverageData) RegisterBlock(file string, startLine, startCol, endLine, endCol, numStmts int) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	key := blockKey(file, startLine, startCol)
	if _, exists := cd.BlockMap[key]; exists {
		return // already registered
	}

	block := &CoverBlock{
		File:      file,
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		NumStmts:  numStmts,
		Count:     0,
	}
	cd.Blocks = append(cd.Blocks, block)
	cd.BlockMap[key] = block
}

// MarkCovered marks a code location as executed.
func (cd *CoverageData) MarkCovered(file string, line, col int) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	key := blockKey(file, line, col)
	if block, exists := cd.BlockMap[key]; exists {
		block.Count++
	}
}

// Coverage returns the coverage percentage (0.0 to 100.0).
func (cd *CoverageData) Coverage() float64 {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	if len(cd.Blocks) == 0 {
		return 0.0
	}

	totalStmts := 0
	coveredStmts := 0
	for _, block := range cd.Blocks {
		totalStmts += block.NumStmts
		if block.Count > 0 {
			coveredStmts += block.NumStmts
		}
	}

	if totalStmts == 0 {
		return 0.0
	}
	return float64(coveredStmts) / float64(totalStmts) * 100.0
}

// TotalStatements returns the total number of coverable statements.
func (cd *CoverageData) TotalStatements() int {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	total := 0
	for _, block := range cd.Blocks {
		total += block.NumStmts
	}
	return total
}

// CoveredStatements returns the number of covered statements.
func (cd *CoverageData) CoveredStatements() int {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	covered := 0
	for _, block := range cd.Blocks {
		if block.Count > 0 {
			covered += block.NumStmts
		}
	}
	return covered
}

// WriteProfile writes the coverage profile in Go's cover format.
func (cd *CoverageData) WriteProfile(w io.Writer) error {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	if _, err := fmt.Fprintf(w, "mode: %s\n", cd.Mode); err != nil {
		return err
	}

	sorted := make([]*CoverBlock, len(cd.Blocks))
	copy(sorted, cd.Blocks)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].File != sorted[j].File {
			return sorted[i].File < sorted[j].File
		}
		if sorted[i].StartLine != sorted[j].StartLine {
			return sorted[i].StartLine < sorted[j].StartLine
		}
		return sorted[i].StartCol < sorted[j].StartCol
	})

	for _, block := range sorted {
		if _, err := fmt.Fprintln(w, block.String()); err != nil {
			return err
		}
	}
	return nil
}

// CoverageCollector manages coverage data across multiple packages.
type CoverageCollector struct {
	mu       sync.RWMutex
	Mode     CoverageMode
	Packages map[string]*CoverageData
}

// NewCoverageCollector creates a new coverage collector.
func NewCoverageCollector(mode CoverageMode) *CoverageCollector {
	if mode == "" {
		mode = CoverageModeSet
	}
	return &CoverageCollector{
		Mode:     mode,
		Packages: make(map[string]*CoverageData),
	}
}

// GetOrCreate returns the CoverageData for a package, creating it if necessary.
func (cc *CoverageCollector) GetOrCreate(pkgPath string) *CoverageData {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cd, exists := cc.Packages[pkgPath]; exists {
		return cd
	}

	cd := NewCoverageData(pkgPath, cc.Mode)
	cc.Packages[pkgPath] = cd
	return cd
}

// TotalCoverage returns the aggregate coverage across all packages.
func (cc *CoverageCollector) TotalCoverage() float64 {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	totalStmts := 0
	coveredStmts := 0

	for _, cd := range cc.Packages {
		totalStmts += cd.TotalStatements()
		coveredStmts += cd.CoveredStatements()
	}

	if totalStmts == 0 {
		return 0.0
	}
	return float64(coveredStmts) / float64(totalStmts) * 100.0
}

// WriteProfile writes coverage profiles for all packages.
func (cc *CoverageCollector) WriteProfile(w io.Writer) error {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	if _, err := fmt.Fprintf(w, "mode: %s\n", cc.Mode); err != nil {
		return err
	}

	var allBlocks []*CoverBlock
	for _, cd := range cc.Packages {
		allBlocks = append(allBlocks, cd.Blocks...)
	}

	sort.Slice(allBlocks, func(i, j int) bool {
		if allBlocks[i].File != allBlocks[j].File {
			return allBlocks[i].File < allBlocks[j].File
		}
		if allBlocks[i].StartLine != allBlocks[j].StartLine {
			return allBlocks[i].StartLine < allBlocks[j].StartLine
		}
		return allBlocks[i].StartCol < allBlocks[j].StartCol
	})

	for _, block := range allBlocks {
		if _, err := fmt.Fprintln(w, block.String()); err != nil {
			return err
		}
	}
	return nil
}

// RegisterCoverableStatementsInFile walks a list of statements and registers all coverable blocks.
func (cc *CoverageCollector) RegisterCoverableStatementsInFile(pkgPath, fileName string, body []Stmt) {
	cd := cc.GetOrCreate(pkgPath)
	for _, stmt := range body {
		registerStmtAndChildren(cd, fileName, stmt)
	}
}

// registerStmtAndChildren recursively registers a statement and all its children.
func registerStmtAndChildren(cd *CoverageData, fileName string, stmt Stmt) {
	if stmt == nil {
		return
	}

	span := stmt.GetSpan()
	if !span.IsZero() {
		cd.RegisterBlock(fileName, span.Line, span.Column, span.End.Line, span.End.Column, 1)
	}

	// Recurse into child statements based on statement type
	switch s := stmt.(type) {
	case *BlockStmt:
		for _, child := range s.Body {
			registerStmtAndChildren(cd, fileName, child)
		}
	case *IfStmt:
		if s.Init != nil {
			registerStmtAndChildren(cd, fileName, s.Init)
		}
		for _, child := range s.Then.Body {
			registerStmtAndChildren(cd, fileName, child)
		}
		for _, child := range s.Else.Body {
			registerStmtAndChildren(cd, fileName, child)
		}
	case *ForStmt:
		if s.Init != nil {
			registerStmtAndChildren(cd, fileName, s.Init)
		}
		if s.Post != nil {
			registerStmtAndChildren(cd, fileName, s.Post)
		}
		for _, child := range s.Body {
			registerStmtAndChildren(cd, fileName, child)
		}
	case *RangeStmt:
		for _, child := range s.Body {
			registerStmtAndChildren(cd, fileName, child)
		}
	case *SelectStmt:
		for _, clause := range s.Cases {
			for _, child := range clause.Body {
				registerStmtAndChildren(cd, fileName, child)
			}
		}
	case *SwitchStmt:
		if s.Init != nil {
			registerStmtAndChildren(cd, fileName, s.Init)
		}
		for _, clause := range s.Clauses {
			for _, child := range clause.Body {
				registerStmtAndChildren(cd, fileName, child)
			}
		}
	}
}

// RegisterFuncDecl registers coverable statements from a function declaration.
func (cc *CoverageCollector) RegisterFuncDecl(pkgPath, fileName string, fd *FuncDecl) {
	if fd == nil || fd.Body == nil {
		return
	}
	cc.RegisterCoverableStatementsInFile(pkgPath, fileName, fd.Body)
}
