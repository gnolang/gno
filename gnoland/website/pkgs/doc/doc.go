package doc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

func New(pkgPath string, files map[string]string) (*Package, error) {
	p := Package{
		ImportPath: pkgPath,
		Filenames:  make([]string, 0, len(files)),
	}

	fset := token.NewFileSet()
	gnoFiles := make(map[string]*ast.File)

	for filename, fileContent := range files {
		p.Filenames = append(p.Filenames, filename)

		f, err := parser.ParseFile(fset, filename, fileContent, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		ast.FileExports(f)

		if strings.HasSuffix(filename, "_test.gno") || strings.HasSuffix(filename, "_filetest.gno") {
			continue
		}

		gnoFiles[filename] = f

		if f.Doc != nil {
			doc := f.Doc.Text()
			if p.Doc != "" {
				p.Doc += "\n"
			}
			p.Doc += doc
		}
	}

	sort.Strings(p.Filenames)

	astPkg, _ := ast.NewPackage(fset, gnoFiles, nil, nil)

	p.Name = astPkg.Name

	ast.Inspect(astPkg, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			fn := extractFunc(x)
			p.Funcs = append(p.Funcs, fn)

		case *ast.GenDecl:
			if x.Tok == token.VAR {
				value, _ := extractValue(fset, x)
				p.Vars = append(p.Vars, value)
			}
			if x.Tok == token.CONST {
				value, _ := extractValue(fset, x)
				p.Consts = append(p.Consts, value)
			}
			if x.Tok == token.TYPE {
				for _, spec := range x.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						newType, _ := extractType(fset, ts)
						p.Types = append(p.Types, newType)
					}
				}
			}
		}

		return true
	})

	p.populateType()

	return &p, nil
}
