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
	gnoFiles := make(map[string]*ast.File, len(files))

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

		if p.Name == "" {
			p.Name = f.Name.Name
		}

		if f.Doc != nil {
			doc := f.Doc.Text()
			if p.Doc != "" {
				p.Doc += "\n"
			}
			p.Doc += doc
		}
	}

	for _, f := range gnoFiles {
		for _, decl := range f.Decls {
			switch x := decl.(type) {
			case *ast.FuncDecl:
				if isFuncExported(x) {
					fn := extractFunc(x)
					p.Funcs = append(p.Funcs, fn)
				}
			case *ast.GenDecl:
				switch x.Tok {
				case token.TYPE:
					for _, spec := range x.Specs {
						if ident, ok := spec.(*ast.TypeSpec); ok && ident.Name.IsExported() {
							newType, _ := extractType(fset, ident)
							if x.Doc != nil {
								newType.Doc = x.Doc.Text() + newType.Doc
							}
							p.Types = append(p.Types, newType)
						}
					}
				case token.VAR:
					value, _ := extractValue(fset, x)
					p.Vars = append(p.Vars, value)
				case token.CONST:
					value, _ := extractValue(fset, x)
					p.Consts = append(p.Consts, value)
				}
			}
		}
	}

	for _, t := range p.Types {
		t.Funcs, t.Methods = p.filterTypeFuncs(t.Name)
		t.Vars, t.Consts = p.filterTypeValues(t.Name)
	}

	sort.Slice(p.Types, func(i, j int) bool {
		return p.Types[i].Name < p.Types[j].Name
	})

	sort.Slice(p.Funcs, func(i, j int) bool {
		return p.Funcs[i].Name < p.Funcs[j].Name
	})

	sort.Strings(p.Filenames)

	return &p, nil
}
