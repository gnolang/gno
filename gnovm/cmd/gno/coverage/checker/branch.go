package checker

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"github.com/gnolang/gno/tm2/pkg/std"
	"golang.org/x/tools/go/ast/astutil"
)

type Branch struct {
	Pos    token.Pos
	Taken  bool
	Offset int
}

type BranchCoverage struct {
	branches map[int]*Branch
	fset     *token.FileSet
	debug    bool
}

func NewBranchCoverage(files []*std.MemFile) *BranchCoverage {
	bc := &BranchCoverage{
		branches: make(map[int]*Branch),
		fset:     token.NewFileSet(),
		debug:    true,
	}

	for _, file := range files {
		if astFile, err := parser.ParseFile(bc.fset, file.Name, file.Body, parser.AllErrors); err == nil {
			ast.Inspect(astFile, func(n ast.Node) bool {
				bc.identifyBranch(n)
				return true
			})
		}
	}

	bc.log("Total branches identified: %d", len(bc.branches))
	return bc
}

func (bc *BranchCoverage) identifyBranch(n ast.Node) {
	switch stmt := n.(type) {
	case *ast.IfStmt:
		bc.addBranch(stmt.If)
		bc.handleComplexCondition(stmt.Cond)
		if stmt.Else != nil {
			switch elseStmt := stmt.Else.(type) {
			case *ast.BlockStmt:
				bc.addBranch(elseStmt.Pos())
			case *ast.IfStmt:
				bc.addBranch(elseStmt.If)
			}
		} else {
			// implicit else branch
			bc.addBranch(stmt.Body.End())
		}
	case *ast.CaseClause:
		bc.addBranch(stmt.Pos())
	case *ast.SwitchStmt:
		for _, s := range stmt.Body.List {
			if cc, ok := s.(*ast.CaseClause); ok {
				bc.addBranch(cc.Pos())
			}
		}
	case *ast.FuncDecl:
		bc.addBranch(stmt.Body.Lbrace)
	case *ast.ForStmt:
		if stmt.Cond != nil {
			bc.addBranch(stmt.Cond.Pos())
		}
	case *ast.RangeStmt:
		bc.addBranch(stmt.For)
	case *ast.DeferStmt:
		bc.addBranch(stmt.Defer)
	}
}

func (bc *BranchCoverage) addBranch(pos token.Pos) {
	offset := bc.fset.Position(pos).Offset
	bc.branches[offset] = &Branch{Pos: pos, Taken: false, Offset: offset}
	bc.log("Branch added at offset %d", offset)
}

func (bc *BranchCoverage) handleComplexCondition(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		if e.Op == token.LAND || e.Op == token.LOR {
			bc.addBranch(e.X.Pos())
			bc.addBranch(e.Y.Pos())
			bc.handleComplexCondition(e.X)
			bc.handleComplexCondition(e.Y)
		}
	case *ast.ParenExpr:
		bc.handleComplexCondition(e.X)
	case *ast.UnaryExpr:
		if e.Op == token.NOT {
			bc.addBranch(e.X.Pos())
			bc.handleComplexCondition(e.X)
		}
	}
}

func (bc *BranchCoverage) Instrument(file *std.MemFile) *std.MemFile {
	astFile, err := parser.ParseFile(bc.fset, file.Name, file.Body, parser.AllErrors)
	if err != nil {
		bc.log("Error parsing file %s: %v", file.Name, err)
		return file
	}

	astutil.Apply(astFile, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch stmt := n.(type) {
		case *ast.IfStmt:
			bc.instrumentIfStmt(stmt)
		case *ast.CaseClause:
			bc.instrumentCaseClause(stmt)
		}
		return true
	}, nil)

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, bc.fset, astFile); err != nil {
		bc.log("Error printing instrumented file: %v", err)
		return file
	}

	return &std.MemFile{
		Name: file.Name,
		Body: buf.String(),
	}
}

func (bc *BranchCoverage) instrumentIfStmt(ifStmt *ast.IfStmt) {
	bc.insertMarkBranchStmt(ifStmt.Body, ifStmt.If)
	if ifStmt.Else != nil {
		switch elseStmt := ifStmt.Else.(type) {
		case *ast.BlockStmt:
			bc.insertMarkBranchStmt(elseStmt, ifStmt.Else.Pos())
		case *ast.IfStmt:
			bc.instrumentIfStmt(elseStmt)
		}
	}
}

func (bc *BranchCoverage) instrumentCaseClause(caseClause *ast.CaseClause) {
	bc.insertMarkBranchStmt(&ast.BlockStmt{List: caseClause.Body}, caseClause.Pos())
}

func (bc *BranchCoverage) insertMarkBranchStmt(block *ast.BlockStmt, pos token.Pos) {
	offset := bc.fset.Position(pos).Offset
	bc.log("Inserting branch mark at offset %d", offset)
	markStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("bc"),
				Sel: ast.NewIdent("MarkBranchTaken"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.INT,
					Value: fmt.Sprintf("%d", offset),
				},
			},
		},
	}
	block.List = append([]ast.Stmt{markStmt}, block.List...)
}

func (bc *BranchCoverage) MarkBranchTaken(offset int) {
	if branch, exists := bc.branches[offset]; exists {
		branch.Taken = true
		bc.log("Branch taken at offset %d", offset)
	} else {
		bc.log("No branch found at offset %d", offset)
	}
}

func (bc *BranchCoverage) CalculateCoverage() float64 {
	total := len(bc.branches)
	if total == 0 {
		return 0
	}

	taken := 0
	for _, branch := range bc.branches {
		if branch.Taken {
			taken++
		}
	}

	coverage := float64(taken) / float64(total)
	bc.log("Total branches: %d, Taken: %d, Coverage: %.2f", total, taken, coverage)
	return coverage
}

func (bc *BranchCoverage) log(format string, args ...interface{}) {
	if bc.debug {
		fmt.Printf(format+"\n", args...)
	}
}
