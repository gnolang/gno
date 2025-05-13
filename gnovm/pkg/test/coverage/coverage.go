package coverage

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

var globalTracker = NewCoverageTracker()

// CoverageTracker tracks the coverage data
type CoverageTracker struct {
	data map[string]map[int]int // filename -> line number -> count
}

func NewCoverageTracker() *CoverageTracker {
	return &CoverageTracker{
		data: make(map[string]map[int]int),
	}
}

// MarkLine mark the line as executed
func (ct *CoverageTracker) MarkLine(filename string, line int) {
	if _, ok := ct.data[filename]; !ok {
		ct.data[filename] = make(map[int]int)
	}
	ct.data[filename][line]++
}

// GetCoverage return the coverage data for a specific file
func (ct *CoverageTracker) GetCoverage(filename string) map[int]int {
	return ct.data[filename]
}

// CoverageInstrumenter instrument the AST to add coverage
type CoverageInstrumenter struct {
	fset     *token.FileSet
	tracker  *CoverageTracker
	filename string
}

// NewCoverageInstrumenter create a new CoverageInstrumenter
func NewCoverageInstrumenter(tracker *CoverageTracker, filename string) *CoverageInstrumenter {
	return &CoverageInstrumenter{
		fset:     token.NewFileSet(),
		tracker:  tracker,
		filename: filename,
	}
}

// InstrumentFile instrument the file
func (ci *CoverageInstrumenter) InstrumentFile(content []byte) ([]byte, error) {
	// parse the file
	f, err := parser.ParseFile(ci.fset, ci.filename, string(content), parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	// add testing import if not already present
	hasTestingImport := false
	for _, imp := range f.Imports {
		if imp.Path.Value == "\"testing\"" {
			hasTestingImport = true
			break
		}
	}
	if !hasTestingImport {
		// Create a new import declaration
		importDecl := &ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: "\"testing\"",
					},
				},
			},
		}

		// If there are existing imports, add testing to the first import declaration
		if len(f.Imports) > 0 {
			// Find the first import declaration
			for _, decl := range f.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
					// Add testing to the existing import declaration
					genDecl.Specs = append(genDecl.Specs, importDecl.Specs[0])
					break
				}
			}
		} else {
			// Add the import declaration to the beginning of the file
			if len(f.Decls) > 0 {
				f.Decls = append([]ast.Decl{importDecl}, f.Decls...)
			} else {
				f.Decls = []ast.Decl{importDecl}
			}
		}
	}

	// modify the AST
	ast.Walk(ci, f)

	// convert the modified AST to code
	var buf strings.Builder
	if err := printer.Fprint(&buf, ci.fset, f); err != nil {
		return nil, fmt.Errorf("code generation failed: %w", err)
	}

	return []byte(buf.String()), nil
}

// createMarkLineStmt creates an ast.ExprStmt that records the given filename and line number
func (ci *CoverageInstrumenter) createMarkLineStmt(filename string, line int) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "testing"},
				Sel: &ast.Ident{Name: "MarkLine"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", filename)},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", line)},
			},
		},
	}
}

// instrumentBlockStmt adds a call to the markLine function to the BlockStmt
func (ci *CoverageInstrumenter) instrumentBlockStmt(block *ast.BlockStmt, line int) {
	if block == nil {
		return
	}

	markStmt := ci.createMarkLineStmt(ci.filename, line)
	block.List = append([]ast.Stmt{markStmt}, block.List...)

	// handle return statements
	for i := 0; i < len(block.List); i++ {
		if ret, ok := block.List[i].(*ast.ReturnStmt); ok {
			retLine := ci.fset.Position(ret.Pos()).Line
			markLineStmt := ci.createMarkLineStmt(ci.filename, retLine)
			block.List = append(block.List[:i], append([]ast.Stmt{markLineStmt}, block.List[i:]...)...)
			i++ // due to the inserted code, increment the index
		}
	}
}

// instrumentCaseList adds a call to the markLine function to the case list
func (ci *CoverageInstrumenter) instrumentCaseStmts(body []ast.Stmt, line int) []ast.Stmt {
	markStmt := ci.createMarkLineStmt(ci.filename, line)
	return append([]ast.Stmt{markStmt}, body...)
}

// Visit visit the AST node and add coverage
func (ci *CoverageInstrumenter) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	// get the position information of the node
	pos := ci.fset.Position(node.Pos())
	line := pos.Line

	// only instrument executable nodes
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Body != nil {
			ci.instrumentBlockStmt(n.Body, line)
		}
	case *ast.IfStmt:
		if n.Cond != nil {
			condLine := ci.fset.Position(n.Cond.Pos()).Line
			ci.instrumentBlockStmt(n.Body, condLine)
		}
	case *ast.ForStmt:
		if n.Cond != nil {
			condLine := ci.fset.Position(n.Cond.Pos()).Line
			ci.instrumentBlockStmt(n.Body, condLine)
		}
	case *ast.RangeStmt:
		ci.instrumentBlockStmt(n.Body, line)
	case *ast.SwitchStmt:
		ci.instrumentBlockStmt(n.Body, line)
	case *ast.SelectStmt:
		ci.instrumentBlockStmt(n.Body, line)
	case *ast.CaseClause:
		n.Body = ci.instrumentCaseStmts(n.Body, line)
	case *ast.CommClause:
		n.Body = ci.instrumentCaseStmts(n.Body, line)
	case *ast.BlockStmt:
		// handle return statements in the block
		for i := 0; i < len(n.List); i++ {
			if ret, ok := n.List[i].(*ast.ReturnStmt); ok {
				retLine := ci.fset.Position(ret.Pos()).Line
				markLineStmt := ci.createMarkLineStmt(ci.filename, retLine)
				n.List = append(n.List[:i], append([]ast.Stmt{markLineStmt}, n.List[i:]...)...)
				i++ // due to the inserted code, increment the index
			}
		}
	}

	return ci
}

// InstrumentPackage instrument the package
func InstrumentPackage(pkg *std.MemPackage) error {
	for _, file := range pkg.Files {
		// skip test files
		if strings.HasSuffix(file.Name, "_test.gno") {
			continue
		}

		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}

		instrumenter := NewCoverageInstrumenter(globalTracker, file.Name)
		instrumented, err := instrumenter.InstrumentFile([]byte(file.Body))
		if err != nil {
			return fmt.Errorf("failed to instrument file %s: %w", file.Name, err)
		}
		file.Body = string(instrumented)
	}
	return nil
}
