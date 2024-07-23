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
	Pos      token.Pos
	Taken    bool
	Filename string
	Line     int
}

type BranchCoverage struct {
	branches map[token.Pos]*Branch
	fset     *token.FileSet
	debug    bool
}

func NewBranchCoverage(files []*std.MemFile) *BranchCoverage {
	bc := &BranchCoverage{
		branches: make(map[token.Pos]*Branch),
		fset:     token.NewFileSet(),
		debug:    true,
	}

	for _, file := range files {
		if astFile, err := parser.ParseFile(bc.fset, file.Name, file.Body, parser.AllErrors); err == nil {
			ast.Inspect(astFile, func(n ast.Node) bool {
				switch stmt := n.(type) {
				case *ast.IfStmt:
					pos := stmt.If
					bc.branches[pos] = &Branch{Pos: pos, Taken: false}
					if stmt.Else != nil {
						pos = stmt.Else.Pos()
						bc.branches[pos] = &Branch{Pos: pos, Taken: false}
					}
					// TODO: add more cases
				}
				return true
			})
		}
	}

	return bc
}

func (bc *BranchCoverage) Instrument(file *std.MemFile) *std.MemFile {
	bc.log("instrumenting file: %s", file.Name)
	astFile, err := parser.ParseFile(bc.fset, file.Name, file.Body, parser.AllErrors)
	if err != nil {
		bc.log("error parsing file %s: %v", file.Name, err)
		return file
	}

	instrumentedFile := astutil.Apply(astFile, nil, func(c *astutil.Cursor) bool {
		node := c.Node()
		switch n := node.(type) {
		case *ast.IfStmt:
			bc.instrumentIfStmt(n)
		case *ast.SwitchStmt:
			bc.instrumentSwitchStmt(n)
		case *ast.ForStmt:
			bc.instrumentForStmt(n)
		}
		return true
	})

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, bc.fset, instrumentedFile); err != nil {
		bc.log("error printing instrumented file: %v", err)
		return file
	}

	return &std.MemFile{
		Name: file.Name,
		Body: buf.String(),
	}
}

func (bc *BranchCoverage) instrumentIfStmt(ifStmt *ast.IfStmt) {
	// instrument if statement
	ifPos := bc.fset.Position(ifStmt.If)
	bc.branches[ifStmt.If] = &Branch{Pos: ifStmt.If, Filename: ifPos.Filename, Line: ifPos.Line, Taken: false}
	ifStmt.Body.List = append([]ast.Stmt{bc.createMarkBranchStmt(ifStmt.If)}, ifStmt.Body.List...)

	// instrument else statement
	if ifStmt.Else != nil {
		elsePos := bc.fset.Position(ifStmt.Else.Pos())
		bc.branches[ifStmt.Else.Pos()] = &Branch{Pos: ifStmt.Else.Pos(), Taken: false, Filename: elsePos.Filename, Line: elsePos.Line}
		switch elseBody := ifStmt.Else.(type) {
		case *ast.BlockStmt:
			elseBody.List = append([]ast.Stmt{bc.createMarkBranchStmt(ifStmt.Else.Pos())}, elseBody.List...)
		case *ast.IfStmt:
			// For 'else if', recursively instrument
			bc.instrumentIfStmt(elseBody)
		}
	}
}

func (bc *BranchCoverage) instrumentSwitchStmt(switchStmt *ast.SwitchStmt) {
	for _, stmt := range switchStmt.Body.List {
		if caseClause, ok := stmt.(*ast.CaseClause); ok {
			casePos := bc.fset.Position(caseClause.Pos())
			bc.branches[caseClause.Pos()] = &Branch{Pos: caseClause.Pos(), Taken: false, Filename: casePos.Filename, Line: casePos.Line}
			caseClause.Body = append([]ast.Stmt{bc.createMarkBranchStmt(caseClause.Pos())}, caseClause.Body...)
		}
	}
}

func (bc *BranchCoverage) instrumentForStmt(forStmt *ast.ForStmt) {
	forPos := bc.fset.Position(forStmt.For)
	bc.branches[forStmt.For] = &Branch{Pos: forStmt.For, Taken: false, Filename: forPos.Filename, Line: forPos.Line}
	forStmt.Body.List = append([]ast.Stmt{bc.createMarkBranchStmt(forStmt.For)}, forStmt.Body.List...)
}

func (bc *BranchCoverage) createMarkBranchStmt(pos token.Pos) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("bc"),
				Sel: ast.NewIdent("MarkBranchTaken"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.INT,
					Value: fmt.Sprintf("%d", bc.fset.Position(pos).Offset),
				},
			},
		},
	}
}

func (bc *BranchCoverage) MarkBranchTaken(pos token.Pos) {
	if branch, exists := bc.branches[pos]; exists {
		branch.Taken = true
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

	return float64(taken) / float64(total)
}

func (bc *BranchCoverage) log(format string, args ...interface{}) {
	if bc.debug {
		fmt.Printf(format+"\n", args...)
	}
}
