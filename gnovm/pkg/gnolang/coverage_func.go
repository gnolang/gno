package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type FuncCoverage struct {
	Name      string
	StartLine int
	EndLine   int
	Covered   int
	Total     int
}

// findFuncs finds the functions in a file.
func findFuncs(name string) ([]*FuncExtent, error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, name, nil, 0)
	if err != nil {
		return nil, err
	}

	v := &FuncVisitor{
		fset:    fset,
		name:    name,
		astFile: parsed,
	}

	ast.Walk(v, parsed)

	return v.funcs, nil
}

// FuncExtent describes a function's extent in the source by file and position.
type FuncExtent struct {
	name      string
	startLine int
	startCol  int
	endLine   int
	endCol    int
}

// FuncVisitor implements the visitor that builds the function position list for a file.
type FuncVisitor struct {
	fset    *token.FileSet
	name    string // name of file
	astFile *ast.File
	funcs   []*FuncExtent
}

func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Body == nil {
			break // ignore functions with no body
		}
		start := v.fset.Position(n.Pos())
		end := v.fset.Position(n.End())
		fe := &FuncExtent{
			name:      n.Name.Name,
			startLine: start.Line,
			startCol:  start.Column,
			endLine:   end.Line,
			endCol:    end.Column,
		}
		v.funcs = append(v.funcs, fe)
	}
	return v
}

func (c *CoverageData) ParseFile(filePath string) error {
	funcs, err := findFuncs(filePath)
	if err != nil {
		return err
	}

	c.Functions[filePath] = make([]FuncCoverage, len(funcs))
	for i, f := range funcs {
		c.Functions[filePath][i] = FuncCoverage{
			Name:      f.name,
			StartLine: f.startLine,
			EndLine:   f.endLine,
			Covered:   0,
			Total:     f.endLine - f.startLine + 1,
		}
	}

	return nil
}
