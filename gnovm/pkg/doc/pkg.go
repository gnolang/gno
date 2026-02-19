//nolint:staticcheck // code copied from golang/go, staticcheck complains for usage of ast.Package.
package doc

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type pkgData struct {
	name      string
	dir       bfsDir
	fset      *token.FileSet
	files     []*ast.File
	testFiles []*ast.File
	symbols   []symbolData
}

const (
	symbolDataValue byte = iota
	symbolDataType
	symbolDataFunc
	symbolDataMethod
	symbolDataStructField
	symbolDataInterfaceMethod
)

type symbolData struct {
	symbol     string
	accessible string
	typ        byte
}

func newPkgData(dir bfsDir, unexported bool) (*pkgData, error) {
	mptype := gno.MPAnyProd
	mpkg, err := gno.ReadMemPackage(dir.dir, dir.importPath, mptype)
	if err != nil {
		return nil, fmt.Errorf("commands/doc: read files %q: %w", dir.dir, err)
	}
	return newPkgDataFromMemPkg(mpkg, unexported)
}

func newPkgDataFromMemPkg(mpkg *std.MemPackage, unexported bool) (*pkgData, error) {
	pkg := &pkgData{
		dir: bfsDir{
			importPath: mpkg.Name,
			dir:        mpkg.Path,
		},
		fset: token.NewFileSet(),
	}
	for _, file := range mpkg.Files {
		n := file.Name
		// Ignore files with prefix . or _ like go tools do.
		// Ignore _filetest.gno, but not _test.gno, as we use those to compute
		// examples.
		if !strings.HasSuffix(n, ".gno") ||
			strings.HasPrefix(n, ".") ||
			strings.HasPrefix(n, "_") ||
			strings.HasSuffix(n, "_filetest.gno") {
			continue
		}
		err := pkg.parseFile(n, file.Body, unexported)
		if err != nil {
			fullPath := filepath.Join(mpkg.Path, n)
			return nil, fmt.Errorf("commands/doc: parse file %q: %w", fullPath, err)
		}
	}

	if len(pkg.files) == 0 {
		return nil, fmt.Errorf("commands/doc: no valid gno files in %q", mpkg.Path)
	}
	pkgName := pkg.files[0].Name.Name
	for _, file := range pkg.files[1:] {
		if file.Name.Name != pkgName {
			return nil, fmt.Errorf("commands/doc: multiple packages (%q / %q) in dir %q", pkgName, file.Name.Name, mpkg.Path)
		}
	}
	pkg.name = pkgName

	return pkg, nil
}

func (pkg *pkgData) parseFile(fileName string, body string, unexported bool) error {
	astf, err := parser.ParseFile(pkg.fset, fileName, body, parser.ParseComments)
	if err != nil {
		return err
	}
	if strings.HasSuffix(fileName, "_test.gno") {
		// add test files separately - we should not add their symbols to the package.
		pkg.testFiles = append(pkg.testFiles, astf)
		return nil
	}
	pkg.files = append(pkg.files, astf)

	// add symbols
	for _, decl := range astf.Decls {
		switch x := decl.(type) {
		case *ast.FuncDecl:
			// prepend receiver if this is a method
			sd := symbolData{
				symbol: x.Name.Name,
				typ:    symbolDataFunc,
			}
			if x.Recv != nil {
				sd.symbol, sd.accessible = typeExprString(x.Recv.List[0].Type), sd.symbol
				if !unexported && !token.IsExported(sd.symbol) {
					continue
				}
				sd.typ = symbolDataMethod
			}
			pkg.symbols = append(pkg.symbols, sd)
		case *ast.GenDecl:
			for _, spec := range x.Specs {
				pkg.appendSpec(spec, unexported)
			}
		}
	}
	return nil
}

func (pkg *pkgData) appendSpec(spec ast.Spec, unexported bool) {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		if !unexported && !token.IsExported(s.Name.Name) {
			return
		}
		pkg.symbols = append(pkg.symbols, symbolData{symbol: s.Name.Name, typ: symbolDataType})
		switch st := s.Type.(type) {
		case *ast.StructType:
			pkg.appendFieldList(s.Name.Name, st.Fields, unexported, symbolDataStructField)
		case *ast.InterfaceType:
			pkg.appendFieldList(s.Name.Name, st.Methods, unexported, symbolDataInterfaceMethod)
		}
	case *ast.ValueSpec:
		for _, name := range s.Names {
			if !unexported && !token.IsExported(name.Name) {
				continue
			}
			pkg.symbols = append(pkg.symbols, symbolData{symbol: name.Name, typ: symbolDataValue})
		}
	}
}

func (pkg *pkgData) appendFieldList(tName string, fl *ast.FieldList, unexported bool, typ byte) {
	if fl == nil {
		return
	}
	for _, field := range fl.List {
		if field.Names == nil {
			if typ == symbolDataInterfaceMethod {
				continue
			}
			embName := typeExprString(field.Type)
			if !unexported && !token.IsExported(embName) {
				continue
			}
			// embedded struct
			pkg.symbols = append(pkg.symbols, symbolData{symbol: tName, accessible: embName, typ: typ})
			continue
		}
		for _, name := range field.Names {
			if !unexported && !token.IsExported(name.Name) {
				continue
			}
			pkg.symbols = append(pkg.symbols, symbolData{symbol: tName, accessible: name.Name, typ: typ})
		}
	}
}

func typeExprString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return typeExprString(t.X)
	}
	return ""
}

func (pkg *pkgData) docPackage() (*ast.Package, *doc.Package, error) {
	// largely taken from go/doc.NewFromFiles source

	// Collect .gno files in a map for ast.NewPackage.
	fileMap := make(map[string]*ast.File)
	for i, file := range pkg.files {
		f := pkg.fset.File(file.Pos())
		if f == nil {
			return nil, nil, fmt.Errorf("commands/doc: file pkg.files[%d] is not found in the provided file set", i)
		}
		fileMap[f.Name()] = file
	}

	// from cmd/doc/pkg.go:
	// go/doc does not include typed constants in the constants
	// list, which is what we want. For instance, time.Sunday is of type
	// time.Weekday, so it is defined in the type but not in the
	// Consts list for the package. This prevents
	//	go doc time.Sunday
	// from finding the symbol. This is why we always have AllDecls.
	mode := doc.AllDecls
	// Always keep the function body to check for crossing(). The caller can set the Body nil if needed
	mode |= doc.PreserveAST

	// Compute package documentation.
	// Assign to blank to ignore errors that can happen due to unresolved identifiers.
	astpkg, _ := ast.NewPackage(pkg.fset, fileMap, simpleImporter, nil)
	p := doc.New(astpkg, pkg.dir.importPath, mode)
	// TODO: classifyExamples(p, Examples(testGoFiles...))

	return astpkg, p, nil
}

func simpleImporter(imports map[string]*ast.Object, path string) (*ast.Object, error) {
	pkg := imports[path]
	if pkg == nil {
		pkg = ast.NewObj(ast.Pkg, gno.LastPathElement(path))
		pkg.Data = ast.NewScope(nil) // required by ast.NewPackage for dot-import
		imports[path] = pkg
	}
	return pkg, nil
}
