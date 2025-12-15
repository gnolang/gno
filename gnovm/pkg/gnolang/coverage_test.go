package gnolang

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestCoverBlock_String(t *testing.T) {
	tests := []struct {
		name     string
		block    CoverBlock
		expected string
	}{
		{
			name: "basic block",
			block: CoverBlock{
				File:      "test.gno",
				StartLine: 10,
				StartCol:  5,
				EndLine:   15,
				EndCol:    2,
				NumStmts:  3,
				Count:     1,
			},
			expected: "test.gno:10.5,15.2 3 1",
		},
		{
			name: "uncovered block",
			block: CoverBlock{
				File:      "main.gno",
				StartLine: 1,
				StartCol:  1,
				EndLine:   1,
				EndCol:    10,
				NumStmts:  1,
				Count:     0,
			},
			expected: "main.gno:1.1,1.10 1 0",
		},
		{
			name: "high execution count",
			block: CoverBlock{
				File:      "loop.gno",
				StartLine: 100,
				StartCol:  2,
				EndLine:   105,
				EndCol:    3,
				NumStmts:  5,
				Count:     1000,
			},
			expected: "loop.gno:100.2,105.3 5 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.block.String()
			if result != tt.expected {
				t.Errorf("CoverBlock.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewCoverageData(t *testing.T) {
	tests := []struct {
		name         string
		pkgPath      string
		mode         CoverageMode
		expectedMode CoverageMode
	}{
		{
			name:         "with set mode",
			pkgPath:      "gno.land/p/demo/avl",
			mode:         CoverageModeSet,
			expectedMode: CoverageModeSet,
		},
		{
			name:         "with count mode",
			pkgPath:      "gno.land/r/demo/users",
			mode:         CoverageModeCount,
			expectedMode: CoverageModeCount,
		},
		{
			name:         "with atomic mode",
			pkgPath:      "gno.land/p/demo/ufmt",
			mode:         CoverageModeAtomic,
			expectedMode: CoverageModeAtomic,
		},
		{
			name:         "empty mode defaults to set",
			pkgPath:      "gno.land/p/demo/test",
			mode:         "",
			expectedMode: CoverageModeSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := NewCoverageData(tt.pkgPath, tt.mode)
			if cd == nil {
				t.Fatal("NewCoverageData returned nil")
			}
			if cd.PkgPath != tt.pkgPath {
				t.Errorf("PkgPath = %q, want %q", cd.PkgPath, tt.pkgPath)
			}
			if cd.Mode != tt.expectedMode {
				t.Errorf("Mode = %q, want %q", cd.Mode, tt.expectedMode)
			}
			if cd.Blocks == nil {
				t.Error("Blocks is nil")
			}
			if cd.BlockMap == nil {
				t.Error("BlockMap is nil")
			}
		})
	}
}

func TestCoverageData_RegisterBlock(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	// Register first block
	cd.RegisterBlock("test.gno", 10, 1, 15, 2, 3)

	if len(cd.Blocks) != 1 {
		t.Errorf("len(Blocks) = %d, want 1", len(cd.Blocks))
	}
	if len(cd.BlockMap) != 1 {
		t.Errorf("len(BlockMap) = %d, want 1", len(cd.BlockMap))
	}

	block := cd.Blocks[0]
	if block.File != "test.gno" {
		t.Errorf("File = %q, want %q", block.File, "test.gno")
	}
	if block.StartLine != 10 {
		t.Errorf("StartLine = %d, want 10", block.StartLine)
	}
	if block.NumStmts != 3 {
		t.Errorf("NumStmts = %d, want 3", block.NumStmts)
	}
	if block.Count != 0 {
		t.Errorf("Count = %d, want 0", block.Count)
	}

	// Register second block
	cd.RegisterBlock("test.gno", 20, 1, 25, 2, 2)
	if len(cd.Blocks) != 2 {
		t.Errorf("len(Blocks) = %d, want 2", len(cd.Blocks))
	}

	// Try to register duplicate block (same file, line, col)
	cd.RegisterBlock("test.gno", 10, 1, 15, 2, 3)
	if len(cd.Blocks) != 2 {
		t.Errorf("len(Blocks) = %d after duplicate, want 2", len(cd.Blocks))
	}
}

func TestCoverageData_MarkCovered(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	// Register blocks
	cd.RegisterBlock("test.gno", 10, 1, 15, 2, 1)
	cd.RegisterBlock("test.gno", 20, 1, 25, 2, 1)

	// Mark first block as covered
	cd.MarkCovered("test.gno", 10, 1)

	if cd.Blocks[0].Count != 1 {
		t.Errorf("Block[0].Count = %d, want 1", cd.Blocks[0].Count)
	}
	if cd.Blocks[1].Count != 0 {
		t.Errorf("Block[1].Count = %d, want 0", cd.Blocks[1].Count)
	}

	// Mark first block again (count should increment)
	cd.MarkCovered("test.gno", 10, 1)
	if cd.Blocks[0].Count != 2 {
		t.Errorf("Block[0].Count = %d after second mark, want 2", cd.Blocks[0].Count)
	}

	// Try marking non-existent block (should not panic)
	cd.MarkCovered("test.gno", 999, 1)
	cd.MarkCovered("nonexistent.gno", 10, 1)
}

func TestCoverageData_Coverage(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*CoverageData)
		expected float64
	}{
		{
			name:     "empty coverage data",
			setup:    func(cd *CoverageData) {},
			expected: 0.0,
		},
		{
			name: "no blocks covered",
			setup: func(cd *CoverageData) {
				cd.RegisterBlock("test.gno", 10, 1, 15, 2, 2)
				cd.RegisterBlock("test.gno", 20, 1, 25, 2, 2)
			},
			expected: 0.0,
		},
		{
			name: "all blocks covered",
			setup: func(cd *CoverageData) {
				cd.RegisterBlock("test.gno", 10, 1, 15, 2, 2)
				cd.RegisterBlock("test.gno", 20, 1, 25, 2, 2)
				cd.MarkCovered("test.gno", 10, 1)
				cd.MarkCovered("test.gno", 20, 1)
			},
			expected: 100.0,
		},
		{
			name: "50% coverage by statements",
			setup: func(cd *CoverageData) {
				cd.RegisterBlock("test.gno", 10, 1, 15, 2, 2) // 2 statements
				cd.RegisterBlock("test.gno", 20, 1, 25, 2, 2) // 2 statements
				cd.MarkCovered("test.gno", 10, 1)             // cover first block
			},
			expected: 50.0,
		},
		{
			name: "weighted coverage",
			setup: func(cd *CoverageData) {
				cd.RegisterBlock("test.gno", 10, 1, 15, 2, 3) // 3 statements
				cd.RegisterBlock("test.gno", 20, 1, 25, 2, 1) // 1 statement
				cd.MarkCovered("test.gno", 10, 1)             // cover first block (3 stmts)
			},
			expected: 75.0, // 3 out of 4 statements
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := NewCoverageData("test/pkg", CoverageModeSet)
			tt.setup(cd)
			result := cd.Coverage()
			if result != tt.expected {
				t.Errorf("Coverage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCoverageData_TotalStatements(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	if cd.TotalStatements() != 0 {
		t.Errorf("TotalStatements() = %d for empty, want 0", cd.TotalStatements())
	}

	cd.RegisterBlock("test.gno", 10, 1, 15, 2, 3)
	if cd.TotalStatements() != 3 {
		t.Errorf("TotalStatements() = %d, want 3", cd.TotalStatements())
	}

	cd.RegisterBlock("test.gno", 20, 1, 25, 2, 5)
	if cd.TotalStatements() != 8 {
		t.Errorf("TotalStatements() = %d, want 8", cd.TotalStatements())
	}
}

func TestCoverageData_CoveredStatements(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	if cd.CoveredStatements() != 0 {
		t.Errorf("CoveredStatements() = %d for empty, want 0", cd.CoveredStatements())
	}

	cd.RegisterBlock("test.gno", 10, 1, 15, 2, 3)
	cd.RegisterBlock("test.gno", 20, 1, 25, 2, 5)

	if cd.CoveredStatements() != 0 {
		t.Errorf("CoveredStatements() = %d before marking, want 0", cd.CoveredStatements())
	}

	cd.MarkCovered("test.gno", 10, 1)
	if cd.CoveredStatements() != 3 {
		t.Errorf("CoveredStatements() = %d, want 3", cd.CoveredStatements())
	}

	cd.MarkCovered("test.gno", 20, 1)
	if cd.CoveredStatements() != 8 {
		t.Errorf("CoveredStatements() = %d, want 8", cd.CoveredStatements())
	}
}

func TestCoverageData_WriteProfile(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)
	cd.RegisterBlock("b.gno", 20, 1, 25, 2, 2)
	cd.RegisterBlock("a.gno", 10, 1, 15, 2, 3)
	cd.MarkCovered("a.gno", 10, 1)

	var buf bytes.Buffer
	err := cd.WriteProfile(&buf)
	if err != nil {
		t.Fatalf("WriteProfile() error = %v", err)
	}

	output := buf.String()

	// Check mode line
	if !strings.HasPrefix(output, "mode: set\n") {
		t.Errorf("WriteProfile() output should start with 'mode: set\\n', got %q", output[:min(len(output), 20)])
	}

	// Check that blocks are sorted by file name
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("WriteProfile() produced %d lines, want 3", len(lines))
	}

	// First block should be from a.gno (alphabetically first)
	if !strings.HasPrefix(lines[1], "a.gno:") {
		t.Errorf("First block should be from a.gno, got %q", lines[1])
	}

	// Second block should be from b.gno
	if !strings.HasPrefix(lines[2], "b.gno:") {
		t.Errorf("Second block should be from b.gno, got %q", lines[2])
	}

	// Check coverage count in output
	if !strings.Contains(lines[1], " 1") { // covered
		t.Errorf("a.gno block should show count 1, got %q", lines[1])
	}
	if !strings.Contains(lines[2], " 0") { // not covered
		t.Errorf("b.gno block should show count 0, got %q", lines[2])
	}
}

func TestNewCoverageCollector(t *testing.T) {
	tests := []struct {
		name         string
		mode         CoverageMode
		expectedMode CoverageMode
	}{
		{
			name:         "with set mode",
			mode:         CoverageModeSet,
			expectedMode: CoverageModeSet,
		},
		{
			name:         "with count mode",
			mode:         CoverageModeCount,
			expectedMode: CoverageModeCount,
		},
		{
			name:         "empty mode defaults to set",
			mode:         "",
			expectedMode: CoverageModeSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := NewCoverageCollector(tt.mode)
			if cc == nil {
				t.Fatal("NewCoverageCollector returned nil")
			}
			if cc.Mode != tt.expectedMode {
				t.Errorf("Mode = %q, want %q", cc.Mode, tt.expectedMode)
			}
			if cc.Packages == nil {
				t.Error("Packages map is nil")
			}
		})
	}
}

func TestCoverageCollector_GetOrCreate(t *testing.T) {
	cc := NewCoverageCollector(CoverageModeSet)

	// Get first package
	cd1 := cc.GetOrCreate("gno.land/p/demo/avl")
	if cd1 == nil {
		t.Fatal("GetOrCreate returned nil")
	}
	if cd1.PkgPath != "gno.land/p/demo/avl" {
		t.Errorf("PkgPath = %q, want %q", cd1.PkgPath, "gno.land/p/demo/avl")
	}
	if cd1.Mode != CoverageModeSet {
		t.Errorf("Mode = %q, want %q", cd1.Mode, CoverageModeSet)
	}

	// Get same package again (should return same instance)
	cd2 := cc.GetOrCreate("gno.land/p/demo/avl")
	if cd1 != cd2 {
		t.Error("GetOrCreate should return same instance for same package")
	}

	// Get different package
	cd3 := cc.GetOrCreate("gno.land/r/demo/users")
	if cd3 == nil {
		t.Fatal("GetOrCreate returned nil for new package")
	}
	if cd3 == cd1 {
		t.Error("GetOrCreate should return different instance for different package")
	}

	// Verify packages map
	if len(cc.Packages) != 2 {
		t.Errorf("len(Packages) = %d, want 2", len(cc.Packages))
	}
}

func TestCoverageCollector_TotalCoverage(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*CoverageCollector)
		expected float64
	}{
		{
			name:     "empty collector",
			setup:    func(cc *CoverageCollector) {},
			expected: 0.0,
		},
		{
			name: "single package fully covered",
			setup: func(cc *CoverageCollector) {
				cd := cc.GetOrCreate("pkg1")
				cd.RegisterBlock("test.gno", 10, 1, 15, 2, 2)
				cd.MarkCovered("test.gno", 10, 1)
			},
			expected: 100.0,
		},
		{
			name: "multiple packages mixed coverage",
			setup: func(cc *CoverageCollector) {
				// Package 1: 2 statements, all covered
				cd1 := cc.GetOrCreate("pkg1")
				cd1.RegisterBlock("a.gno", 10, 1, 15, 2, 2)
				cd1.MarkCovered("a.gno", 10, 1)

				// Package 2: 2 statements, none covered
				cd2 := cc.GetOrCreate("pkg2")
				cd2.RegisterBlock("b.gno", 10, 1, 15, 2, 2)
			},
			expected: 50.0, // 2 out of 4 total statements
		},
		{
			name: "weighted across packages",
			setup: func(cc *CoverageCollector) {
				// Package 1: 3 statements, all covered
				cd1 := cc.GetOrCreate("pkg1")
				cd1.RegisterBlock("a.gno", 10, 1, 15, 2, 3)
				cd1.MarkCovered("a.gno", 10, 1)

				// Package 2: 1 statement, not covered
				cd2 := cc.GetOrCreate("pkg2")
				cd2.RegisterBlock("b.gno", 10, 1, 15, 2, 1)
			},
			expected: 75.0, // 3 out of 4 total statements
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := NewCoverageCollector(CoverageModeSet)
			tt.setup(cc)
			result := cc.TotalCoverage()
			if result != tt.expected {
				t.Errorf("TotalCoverage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCoverageCollector_WriteProfile(t *testing.T) {
	cc := NewCoverageCollector(CoverageModeCount)

	// Add blocks from multiple packages
	cd1 := cc.GetOrCreate("pkg1")
	cd1.RegisterBlock("z.gno", 10, 1, 15, 2, 1)
	cd1.MarkCovered("z.gno", 10, 1)

	cd2 := cc.GetOrCreate("pkg2")
	cd2.RegisterBlock("a.gno", 5, 1, 10, 2, 2)

	var buf bytes.Buffer
	err := cc.WriteProfile(&buf)
	if err != nil {
		t.Fatalf("WriteProfile() error = %v", err)
	}

	output := buf.String()

	// Check mode line
	if !strings.HasPrefix(output, "mode: count\n") {
		t.Errorf("WriteProfile() should start with 'mode: count\\n', got prefix %q", output[:min(len(output), 20)])
	}

	// Check that blocks from all packages are included and sorted
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("WriteProfile() produced %d lines, want 3", len(lines))
	}

	// First block should be from a.gno (alphabetically first)
	if !strings.HasPrefix(lines[1], "a.gno:") {
		t.Errorf("First block should be from a.gno, got %q", lines[1])
	}

	// Second block should be from z.gno
	if !strings.HasPrefix(lines[2], "z.gno:") {
		t.Errorf("Second block should be from z.gno, got %q", lines[2])
	}
}

func TestBlockKey(t *testing.T) {
	tests := []struct {
		file     string
		line     int
		col      int
		expected string
	}{
		{"test.gno", 10, 5, "test.gno:10:5"},
		{"main.gno", 1, 1, "main.gno:1:1"},
		{"path/to/file.gno", 100, 20, "path/to/file.gno:100:20"},
	}

	for _, tt := range tests {
		result := blockKey(tt.file, tt.line, tt.col)
		if result != tt.expected {
			t.Errorf("blockKey(%q, %d, %d) = %q, want %q",
				tt.file, tt.line, tt.col, result, tt.expected)
		}
	}
}

func TestCoverageCollector_RegisterCoverableStatementsInFile(t *testing.T) {
	cc := NewCoverageCollector(CoverageModeSet)

	// Create some test statements with spans
	stmts := []Stmt{
		&ExprStmt{
			Attributes: Attributes{
				Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 10, Column: 20}},
			},
		},
		&ReturnStmt{
			Attributes: Attributes{
				Span: Span{Pos: Pos{Line: 15, Column: 1}, End: Pos{Line: 15, Column: 10}},
			},
		},
	}

	cc.RegisterCoverableStatementsInFile("test/pkg", "test.gno", stmts)

	cd := cc.Packages["test/pkg"]
	if cd == nil {
		t.Fatal("Package not created")
	}

	if len(cd.Blocks) != 2 {
		t.Errorf("len(Blocks) = %d, want 2", len(cd.Blocks))
	}
}

func TestCoverageCollector_RegisterFuncDecl(t *testing.T) {
	cc := NewCoverageCollector(CoverageModeSet)

	// Create a function declaration with body statements
	fd := &FuncDecl{
		Body: Body{
			&ExprStmt{
				Attributes: Attributes{
					Span: Span{Pos: Pos{Line: 5, Column: 2}, End: Pos{Line: 5, Column: 15}},
				},
			},
			&ReturnStmt{
				Attributes: Attributes{
					Span: Span{Pos: Pos{Line: 6, Column: 2}, End: Pos{Line: 6, Column: 10}},
				},
			},
		},
	}

	cc.RegisterFuncDecl("test/pkg", "func.gno", fd)

	cd := cc.Packages["test/pkg"]
	if cd == nil {
		t.Fatal("Package not created")
	}

	if len(cd.Blocks) != 2 {
		t.Errorf("len(Blocks) = %d, want 2", len(cd.Blocks))
	}
}

func TestCoverageCollector_RegisterFuncDecl_NilCases(t *testing.T) {
	cc := NewCoverageCollector(CoverageModeSet)

	// Nil function declaration
	cc.RegisterFuncDecl("test/pkg", "test.gno", nil)
	if len(cc.Packages) != 0 {
		t.Error("Should not create package for nil FuncDecl")
	}

	// Function declaration with nil body
	fd := &FuncDecl{Body: nil}
	cc.RegisterFuncDecl("test/pkg", "test.gno", fd)
	if len(cc.Packages) != 0 {
		t.Error("Should not create package for FuncDecl with nil body")
	}
}

func TestRegisterStmtAndChildren_IfStmt(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	ifStmt := &IfStmt{
		Attributes: Attributes{
			Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 20, Column: 2}},
		},
		Init: &ExprStmt{
			Attributes: Attributes{
				Span: Span{Pos: Pos{Line: 10, Column: 5}, End: Pos{Line: 10, Column: 15}},
			},
		},
		Then: IfCaseStmt{
			Body: Body{
				&ExprStmt{
					Attributes: Attributes{
						Span: Span{Pos: Pos{Line: 11, Column: 2}, End: Pos{Line: 11, Column: 10}},
					},
				},
			},
		},
		Else: IfCaseStmt{
			Body: Body{
				&ExprStmt{
					Attributes: Attributes{
						Span: Span{Pos: Pos{Line: 15, Column: 2}, End: Pos{Line: 15, Column: 10}},
					},
				},
			},
		},
	}

	registerStmtAndChildren(cd, "test.gno", ifStmt)

	// Should register: if stmt itself, init, then body stmt, else body stmt
	if len(cd.Blocks) != 4 {
		t.Errorf("len(Blocks) = %d, want 4", len(cd.Blocks))
	}
}

func TestRegisterStmtAndChildren_ForStmt(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	forStmt := &ForStmt{
		Attributes: Attributes{
			Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 20, Column: 2}},
		},
		Init: &ExprStmt{
			Attributes: Attributes{
				Span: Span{Pos: Pos{Line: 10, Column: 5}, End: Pos{Line: 10, Column: 15}},
			},
		},
		Post: &ExprStmt{
			Attributes: Attributes{
				Span: Span{Pos: Pos{Line: 10, Column: 20}, End: Pos{Line: 10, Column: 25}},
			},
		},
		Body: Body{
			&ExprStmt{
				Attributes: Attributes{
					Span: Span{Pos: Pos{Line: 11, Column: 2}, End: Pos{Line: 11, Column: 10}},
				},
			},
		},
	}

	registerStmtAndChildren(cd, "test.gno", forStmt)

	// Should register: for stmt, init, post, body stmt
	if len(cd.Blocks) != 4 {
		t.Errorf("len(Blocks) = %d, want 4", len(cd.Blocks))
	}
}

func TestRegisterStmtAndChildren_SwitchStmt(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	switchStmt := &SwitchStmt{
		Attributes: Attributes{
			Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 30, Column: 2}},
		},
		Init: &ExprStmt{
			Attributes: Attributes{
				Span: Span{Pos: Pos{Line: 10, Column: 8}, End: Pos{Line: 10, Column: 15}},
			},
		},
		Clauses: []SwitchClauseStmt{
			{
				Body: Body{
					&ExprStmt{
						Attributes: Attributes{
							Span: Span{Pos: Pos{Line: 12, Column: 3}, End: Pos{Line: 12, Column: 15}},
						},
					},
				},
			},
			{
				Body: Body{
					&ExprStmt{
						Attributes: Attributes{
							Span: Span{Pos: Pos{Line: 15, Column: 3}, End: Pos{Line: 15, Column: 15}},
						},
					},
				},
			},
		},
	}

	registerStmtAndChildren(cd, "test.gno", switchStmt)

	// Should register: switch stmt, init, 2 case body stmts
	if len(cd.Blocks) != 4 {
		t.Errorf("len(Blocks) = %d, want 4", len(cd.Blocks))
	}
}

func TestRegisterStmtAndChildren_NilStmt(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	// Should not panic on nil
	registerStmtAndChildren(cd, "test.gno", nil)

	if len(cd.Blocks) != 0 {
		t.Errorf("len(Blocks) = %d, want 0 for nil stmt", len(cd.Blocks))
	}
}

func TestRegisterStmtAndChildren_ZeroSpan(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	// Statement with zero span should not be registered
	stmt := &ExprStmt{
		Attributes: Attributes{
			Span: Span{}, // zero span
		},
	}

	registerStmtAndChildren(cd, "test.gno", stmt)

	if len(cd.Blocks) != 0 {
		t.Errorf("len(Blocks) = %d, want 0 for zero span", len(cd.Blocks))
	}
}

func TestConcurrentAccess(t *testing.T) {
	cc := NewCoverageCollector(CoverageModeSet)
	cd := cc.GetOrCreate("test/pkg")

	// Register some blocks
	for i := 0; i < 100; i++ {
		cd.RegisterBlock("test.gno", i, 1, i+1, 2, 1)
	}

	// Concurrent marking
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				cd.MarkCovered("test.gno", j, 1)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// All blocks should be covered
	if cd.CoveredStatements() != 100 {
		t.Errorf("CoveredStatements() = %d, want 100", cd.CoveredStatements())
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// errorWriter is a writer that always returns an error after n writes
type errorWriter struct {
	writesBeforeError int
	writeCount        int
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.writeCount > w.writesBeforeError {
		return 0, fmt.Errorf("simulated write error")
	}
	return len(p), nil
}

func TestCoverageData_WriteProfile_ModeLineError(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)
	cd.RegisterBlock("test.gno", 10, 1, 15, 2, 1)

	// Error on first write (mode line)
	w := &errorWriter{writesBeforeError: 0}
	err := cd.WriteProfile(w)
	if err == nil {
		t.Error("WriteProfile() should return error when mode line write fails")
	}
}

func TestCoverageData_WriteProfile_BlockLineError(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)
	cd.RegisterBlock("test.gno", 10, 1, 15, 2, 1)

	// Error on second write (block line)
	w := &errorWriter{writesBeforeError: 1}
	err := cd.WriteProfile(w)
	if err == nil {
		t.Error("WriteProfile() should return error when block line write fails")
	}
}

func TestCoverageCollector_WriteProfile_ModeLineError(t *testing.T) {
	cc := NewCoverageCollector(CoverageModeSet)
	cd := cc.GetOrCreate("test/pkg")
	cd.RegisterBlock("test.gno", 10, 1, 15, 2, 1)

	// Error on first write (mode line)
	w := &errorWriter{writesBeforeError: 0}
	err := cc.WriteProfile(w)
	if err == nil {
		t.Error("WriteProfile() should return error when mode line write fails")
	}
}

func TestCoverageCollector_WriteProfile_BlockLineError(t *testing.T) {
	cc := NewCoverageCollector(CoverageModeSet)
	cd := cc.GetOrCreate("test/pkg")
	cd.RegisterBlock("test.gno", 10, 1, 15, 2, 1)

	// Error on second write (block line)
	w := &errorWriter{writesBeforeError: 1}
	err := cc.WriteProfile(w)
	if err == nil {
		t.Error("WriteProfile() should return error when block line write fails")
	}
}

func TestRegisterStmtAndChildren_BlockStmt(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	blockStmt := &BlockStmt{
		Attributes: Attributes{
			Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 20, Column: 2}},
		},
		Body: Body{
			&ExprStmt{
				Attributes: Attributes{
					Span: Span{Pos: Pos{Line: 11, Column: 2}, End: Pos{Line: 11, Column: 10}},
				},
			},
			&ReturnStmt{
				Attributes: Attributes{
					Span: Span{Pos: Pos{Line: 12, Column: 2}, End: Pos{Line: 12, Column: 15}},
				},
			},
		},
	}

	registerStmtAndChildren(cd, "test.gno", blockStmt)

	// Should register: block stmt itself + 2 body statements
	if len(cd.Blocks) != 3 {
		t.Errorf("len(Blocks) = %d, want 3", len(cd.Blocks))
	}
}

func TestRegisterStmtAndChildren_RangeStmt(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	rangeStmt := &RangeStmt{
		Attributes: Attributes{
			Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 20, Column: 2}},
		},
		Body: Body{
			&ExprStmt{
				Attributes: Attributes{
					Span: Span{Pos: Pos{Line: 11, Column: 2}, End: Pos{Line: 11, Column: 10}},
				},
			},
			&ExprStmt{
				Attributes: Attributes{
					Span: Span{Pos: Pos{Line: 12, Column: 2}, End: Pos{Line: 12, Column: 10}},
				},
			},
		},
	}

	registerStmtAndChildren(cd, "test.gno", rangeStmt)

	// Should register: range stmt + 2 body statements
	if len(cd.Blocks) != 3 {
		t.Errorf("len(Blocks) = %d, want 3", len(cd.Blocks))
	}
}

func TestRegisterStmtAndChildren_SelectStmt(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	selectStmt := &SelectStmt{
		Attributes: Attributes{
			Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 30, Column: 2}},
		},
		Cases: []SelectCaseStmt{
			{
				Body: Body{
					&ExprStmt{
						Attributes: Attributes{
							Span: Span{Pos: Pos{Line: 12, Column: 3}, End: Pos{Line: 12, Column: 15}},
						},
					},
				},
			},
			{
				Body: Body{
					&ExprStmt{
						Attributes: Attributes{
							Span: Span{Pos: Pos{Line: 15, Column: 3}, End: Pos{Line: 15, Column: 15}},
						},
					},
					&ExprStmt{
						Attributes: Attributes{
							Span: Span{Pos: Pos{Line: 16, Column: 3}, End: Pos{Line: 16, Column: 15}},
						},
					},
				},
			},
		},
	}

	registerStmtAndChildren(cd, "test.gno", selectStmt)

	// Should register: select stmt + 3 case body statements (1 + 2)
	if len(cd.Blocks) != 4 {
		t.Errorf("len(Blocks) = %d, want 4", len(cd.Blocks))
	}
}

func TestRegisterStmtAndChildren_NestedStatements(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	// Create a for loop with an if statement inside
	forStmt := &ForStmt{
		Attributes: Attributes{
			Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 30, Column: 2}},
		},
		Body: Body{
			&IfStmt{
				Attributes: Attributes{
					Span: Span{Pos: Pos{Line: 11, Column: 2}, End: Pos{Line: 20, Column: 3}},
				},
				Then: IfCaseStmt{
					Body: Body{
						&ExprStmt{
							Attributes: Attributes{
								Span: Span{Pos: Pos{Line: 12, Column: 3}, End: Pos{Line: 12, Column: 15}},
							},
						},
					},
				},
				Else: IfCaseStmt{
					Body: Body{
						&ExprStmt{
							Attributes: Attributes{
								Span: Span{Pos: Pos{Line: 15, Column: 3}, End: Pos{Line: 15, Column: 15}},
							},
						},
					},
				},
			},
		},
	}

	registerStmtAndChildren(cd, "test.gno", forStmt)

	// Should register: for stmt, if stmt, then body stmt, else body stmt = 4 total
	if len(cd.Blocks) != 4 {
		t.Errorf("len(Blocks) = %d, want 4", len(cd.Blocks))
	}
}

func TestCoverageData_WriteProfile_EmptyBlocks(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	var buf bytes.Buffer
	err := cd.WriteProfile(&buf)
	if err != nil {
		t.Fatalf("WriteProfile() error = %v", err)
	}

	output := buf.String()
	// Should only have mode line
	if output != "mode: set\n" {
		t.Errorf("WriteProfile() for empty = %q, want %q", output, "mode: set\n")
	}
}

func TestCoverageCollector_WriteProfile_EmptyPackages(t *testing.T) {
	cc := NewCoverageCollector(CoverageModeCount)

	var buf bytes.Buffer
	err := cc.WriteProfile(&buf)
	if err != nil {
		t.Fatalf("WriteProfile() error = %v", err)
	}

	output := buf.String()
	// Should only have mode line
	if output != "mode: count\n" {
		t.Errorf("WriteProfile() for empty = %q, want %q", output, "mode: count\n")
	}
}

func TestCoverageData_WriteProfile_SortingByLine(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	// Register blocks out of order (same file, different lines)
	cd.RegisterBlock("test.gno", 30, 1, 35, 2, 1)
	cd.RegisterBlock("test.gno", 10, 1, 15, 2, 1)
	cd.RegisterBlock("test.gno", 20, 1, 25, 2, 1)

	var buf bytes.Buffer
	err := cd.WriteProfile(&buf)
	if err != nil {
		t.Fatalf("WriteProfile() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("Expected 4 lines, got %d", len(lines))
	}

	// Verify sorting by line number
	if !strings.Contains(lines[1], "10.1") {
		t.Errorf("First block should be line 10, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "20.1") {
		t.Errorf("Second block should be line 20, got %q", lines[2])
	}
	if !strings.Contains(lines[3], "30.1") {
		t.Errorf("Third block should be line 30, got %q", lines[3])
	}
}

func TestCoverageData_WriteProfile_SortingByColumn(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	// Register blocks with same file and line, different columns
	cd.RegisterBlock("test.gno", 10, 20, 15, 2, 1)
	cd.RegisterBlock("test.gno", 10, 5, 15, 2, 1)
	cd.RegisterBlock("test.gno", 10, 10, 15, 2, 1)

	var buf bytes.Buffer
	err := cd.WriteProfile(&buf)
	if err != nil {
		t.Fatalf("WriteProfile() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("Expected 4 lines, got %d", len(lines))
	}

	// Verify sorting by column number
	if !strings.Contains(lines[1], "10.5") {
		t.Errorf("First block should be col 5, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "10.10") {
		t.Errorf("Second block should be col 10, got %q", lines[2])
	}
	if !strings.Contains(lines[3], "10.20") {
		t.Errorf("Third block should be col 20, got %q", lines[3])
	}
}

func TestRegisterStmtAndChildren_ForStmtNilInitPost(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	// ForStmt with nil Init and Post
	forStmt := &ForStmt{
		Attributes: Attributes{
			Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 20, Column: 2}},
		},
		Init: nil,
		Post: nil,
		Body: Body{
			&ExprStmt{
				Attributes: Attributes{
					Span: Span{Pos: Pos{Line: 11, Column: 2}, End: Pos{Line: 11, Column: 10}},
				},
			},
		},
	}

	registerStmtAndChildren(cd, "test.gno", forStmt)

	// Should register: for stmt + body stmt (no init/post)
	if len(cd.Blocks) != 2 {
		t.Errorf("len(Blocks) = %d, want 2", len(cd.Blocks))
	}
}

func TestRegisterStmtAndChildren_IfStmtNilInit(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	// IfStmt with nil Init
	ifStmt := &IfStmt{
		Attributes: Attributes{
			Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 20, Column: 2}},
		},
		Init: nil,
		Then: IfCaseStmt{
			Body: Body{
				&ExprStmt{
					Attributes: Attributes{
						Span: Span{Pos: Pos{Line: 11, Column: 2}, End: Pos{Line: 11, Column: 10}},
					},
				},
			},
		},
		Else: IfCaseStmt{
			Body: Body{}, // empty else
		},
	}

	registerStmtAndChildren(cd, "test.gno", ifStmt)

	// Should register: if stmt + then body stmt (no init, empty else)
	if len(cd.Blocks) != 2 {
		t.Errorf("len(Blocks) = %d, want 2", len(cd.Blocks))
	}
}

func TestRegisterStmtAndChildren_SwitchStmtNilInit(t *testing.T) {
	cd := NewCoverageData("test/pkg", CoverageModeSet)

	// SwitchStmt with nil Init
	switchStmt := &SwitchStmt{
		Attributes: Attributes{
			Span: Span{Pos: Pos{Line: 10, Column: 1}, End: Pos{Line: 20, Column: 2}},
		},
		Init: nil,
		Clauses: []SwitchClauseStmt{
			{
				Body: Body{
					&ExprStmt{
						Attributes: Attributes{
							Span: Span{Pos: Pos{Line: 12, Column: 3}, End: Pos{Line: 12, Column: 15}},
						},
					},
				},
			},
		},
	}

	registerStmtAndChildren(cd, "test.gno", switchStmt)

	// Should register: switch stmt + case body stmt (no init)
	if len(cd.Blocks) != 2 {
		t.Errorf("len(Blocks) = %d, want 2", len(cd.Blocks))
	}
}
