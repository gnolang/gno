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

type StatementCoverage struct {
	covered map[token.Pos]bool
	files   map[string]*ast.File
	fset    *token.FileSet
	debug   bool
}

func NewStatementCoverage(files []*std.MemFile) *StatementCoverage {
	sc := &StatementCoverage{
		covered: make(map[token.Pos]bool),
		files:   make(map[string]*ast.File),
		fset:    token.NewFileSet(),
		// debug:   true,
	}

	for _, file := range files {
		sc.log("parsing file: %s", file.Name)
		if astFile, err := parser.ParseFile(sc.fset, file.Name, file.Body, parser.AllErrors); err == nil {
			sc.files[file.Name] = astFile
			ast.Inspect(astFile, func(n ast.Node) bool {
				if stmt, ok := n.(ast.Stmt); ok {
					switch stmt.(type) {
					case *ast.BlockStmt:
						// Skip block statements
					default:
						sc.covered[stmt.Pos()] = false
					}
				}
				return true
			})
		} else {
			sc.log("error parsing file %s: %v", file.Name, err)
		}
	}

	sc.log("total statements found: %d", len(sc.covered))
	return sc
}

func (sc *StatementCoverage) Instrument(file *std.MemFile) *std.MemFile {
	sc.log("instrumenting file: %s", file.Name)
	astFile, ok := sc.files[file.Name]
	if !ok {
		return file
	}

	instrumentedFile := astutil.Apply(astFile, nil, func(c *astutil.Cursor) bool {
		node := c.Node()
		if stmt, ok := node.(ast.Stmt); ok {
			if _, exists := sc.covered[stmt.Pos()]; exists {
				pos := sc.fset.Position(stmt.Pos())
				sc.log("instrumenting statement at %s:%d%d", pos.Filename, pos.Line, pos.Column)
				markStmt := &ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("sc"),
							Sel: ast.NewIdent("MarkCovered"),
						},
						Args: []ast.Expr{
							&ast.BasicLit{
								Kind:  token.INT,
								Value: fmt.Sprintf("%d", sc.fset.Position(stmt.Pos()).Offset),
							},
						},
					},
				}

				switch s := stmt.(type) {
				case *ast.BlockStmt:
					s.List = append([]ast.Stmt{markStmt}, s.List...)
				case *ast.ForStmt:
					if s.Body != nil {
						s.Body.List = append([]ast.Stmt{markStmt}, s.Body.List...)
					}
				default:
					c.Replace(&ast.BlockStmt{
						List: []ast.Stmt{markStmt, stmt},
					})
				}
			}
		}
		return true
	})

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, sc.fset, instrumentedFile); err != nil {
		return file
	}

	return &std.MemFile{
		Name: file.Name,
		Body: buf.String(),
	}
}

func (sc *StatementCoverage) MarkCovered(pos token.Pos) {
	for stmt := range sc.covered {
		if stmt == pos {
			filePos := sc.fset.Position(pos)
			sc.log("marking covered: %s:%d:%d", filePos.Filename, filePos.Line, filePos.Column)
			sc.covered[stmt] = true
			return
		}
	}
}

func (sc *StatementCoverage) CalculateCoverage() float64 {
	total := len(sc.covered)
	if total == 0 {
		return 0
	}

	covered := 0
	for stmt, isCovered := range sc.covered {
		pos := sc.fset.Position(stmt)
		if isCovered {
			sc.log("covered: %s:%d:%d", pos.Filename, pos.Line, pos.Column)
			covered++
		} else {
			sc.log("not covered: %s:%d:%d", pos.Filename, pos.Line, pos.Column)
		}
	}

	coverage := float64(covered) / float64(total)
	sc.log("total statement: %d, covered: %d, coverage: %.2f", total, covered, coverage)
	return coverage
}

func (sc *StatementCoverage) log(format string, args ...interface{}) {
	if sc.debug {
		fmt.Printf(format+"\n", args...)
	}
}
