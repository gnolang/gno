package gnofmt

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"path/filepath"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
)

const tabWidth = 8

type (
	declMap map[*ast.Ident]ast.Decl
	fileMap map[string]*ast.File
)

type parsedPackage struct {
	error error
	files fileMap
	decls declMap
}

type Processor struct {
	resolver Resolver
	fset     *token.FileSet

	// cache package parsing in `FormatFile` call
	pkgdirCache map[string]Package // dir -> pkg cache package dir

	// cache for global parsed package
	parsedPackage map[string]*parsedPackage // pkgdir -> parsed package
}

func NewProcessor(r Resolver) *Processor {
	return &Processor{
		resolver:      r,
		fset:          token.NewFileSet(),
		pkgdirCache:   make(map[string]Package),
		parsedPackage: make(map[string]*parsedPackage),
	}
}

// FormatImportFromSource parse and format the source from src. The type of the argument
// for the src parameter must be string, []byte, or [io.Reader].
func (p *Processor) FormatImportFromSource(filename string, src any) ([]byte, error) {
	// Parse the source file
	nodefile, err := p.parseFile(filename, src)
	if err != nil {
		return nil, fmt.Errorf("unable to parse source: %w", err)
	}

	// Collect top level declarations within the source
	pkgDecls := make(declMap)
	collectTopDeclaration(nodefile, pkgDecls)

	// Process and format the parsed node.
	return p.processAndFormat(nodefile, filename, pkgDecls)
}

// FormatPackageFile processes a single Gno file from the given Package and filename.
func (p *Processor) FormatPackageFile(pkg Package, filename string) ([]byte, error) {
	// Process package files.
	pkgc := p.processPackageFiles(pkg.Path(), pkg)
	if pkgc.error != nil {
		return nil, fmt.Errorf("unable to process package: %w", pkgc.error)
	}

	// Retrieve the nodefile for the file.
	nodefile, ok := pkgc.files[filename]
	if !ok {
		return nil, fmt.Errorf("not a valid gno file: %s", filename)
	}

	return p.processAndFormat(nodefile, filename, pkgc.decls)
}

// FormatFile processes a single Gno file from the given file path.
func (p *Processor) FormatFile(file string) ([]byte, error) {
	filename := filepath.Base(file)
	dir := filepath.Dir(file)

	pkg, ok := p.pkgdirCache[dir]
	if !ok {
		var err error
		pkg, err = ParsePackage(p.fset, "", dir)
		if err != nil {
			return nil, fmt.Errorf("unable to parse package %q: %w", dir, err)
		}
		p.pkgdirCache[dir] = pkg
	}

	if pkg == nil {
		fmt.Printf("-> parsing file: %q, %q\n", file, filename)
		// Fallback on src
		return p.FormatImportFromSource(filename, nil)
	}

	path := pkg.Path()
	if path == "" {
		// Use dir as package path
		path = dir
	}

	// Process package files.
	pkgc := p.processPackageFiles(dir, pkg)
	if pkgc.error != nil {
		return nil, fmt.Errorf("unable to process package: %w", pkgc.error)
	}

	// Retrieve the nodefile for the file.
	nodefile, ok := pkgc.files[filename]
	if !ok {
		return nil, fmt.Errorf("not a valid gno file: %s", filename)
	}

	return p.processAndFormat(nodefile, filename, pkgc.decls)
}

func (p *Processor) parseFile(path string, src any) (file *ast.File, err error) {
	// Parse the source file
	file, err = parser.ParseFile(p.fset, path, src, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file %q: %w", path, err)
	}

	return file, nil
}

// Helper function to process and format a parsed AST node.
func (p *Processor) processAndFormat(file *ast.File, filename string, topDecls declMap) ([]byte, error) {
	// Collect unresolved
	unresolved := collectUnresolved(file, topDecls)

	// Cleanup and remove previous unused import
	p.cleanupPreviousImports(file, topDecls, unresolved)

	// Resolve unresolved declarations
	p.resolve(file, unresolved)

	// Process output
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, p.fset, file); err != nil {
		return nil, fmt.Errorf("unable to format file: %w", err)
	}

	// Use go/imports for formating and sort imports.
	ret, err := imports.Process(filename, buf.Bytes(), &imports.Options{
		TabWidth:   tabWidth,
		Comments:   true,
		TabIndent:  true,
		FormatOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to format import: %w", err)
	}

	return ret, nil
}

// processPackageFiles processes Gno package files and collects top-level declarations.
func (p *Processor) processPackageFiles(path string, pkg Package) *parsedPackage {
	pkgc, ok := p.parsedPackage[path]
	if ok {
		return pkgc
	}

	pkgc = &parsedPackage{
		decls: make(declMap),
		files: make(fileMap),
	}
	pkgc.error = ReadWalkPackage(pkg, func(filename string, r io.Reader, err error) error {
		if err != nil {
			return fmt.Errorf("unable to read %q: %w", filename, err)
		}

		file, err := p.parseFile(filename, r)
		if err != nil {
			return err
		}

		collectTopDeclaration(file, pkgc.decls)
		pkgc.files[filename] = file
		return nil
	})
	p.parsedPackage[path] = pkgc

	return pkgc
}

// collectTopDeclaration collects top-level declarations from a single file.
func collectTopDeclaration(file *ast.File, topDecls declMap) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					topDecls[s.Name] = d
				case *ast.ValueSpec:
					for _, name := range s.Names {
						topDecls[name] = d
					}
				}
			}
		case *ast.FuncDecl:
			// Check for top-level function
			if d.Recv == nil && d.Name != nil && d.Name.Obj != nil {
				topDecls[d.Name] = d
			}
		}
	}
}

// collectUnresolved collects unresolved identifiers and declarations.
func collectUnresolved(file *ast.File, topDecls declMap) map[string]map[string]bool {
	unresolved := map[string]map[string]bool{}
	unresolvedList := []*ast.Ident{}
	for _, u := range file.Unresolved {
		if _, ok := unresolved[u.Name]; ok {
			continue
		}

		if isPredeclared(u.Name) {
			continue
		}

		unresolved[u.Name] = map[string]bool{}
		unresolvedList = append(unresolvedList, u)
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch e := n.(type) {
		case *ast.Ident:
			if _, ok := topDecls[e]; ok {
				delete(unresolved, e.Name)
			}
		case *ast.SelectorExpr:
			for _, u := range unresolvedList {
				if u == e.X {
					ident := e.X.(*ast.Ident)
					unresolved[ident.Name][e.Sel.Name] = true
					break
				}
			}
		}

		return true
	})

	// Delete unresolved identifier without any selector
	for u, v := range unresolved {
		if len(v) == 0 { // no selector
			delete(unresolved, u)
		}
	}

	return unresolved
}

// cleanupPreviousImports removes resolved imports from the unresolved list.
func (p *Processor) cleanupPreviousImports(node *ast.File, knownDecls declMap, unresolved map[string]map[string]bool) {
	imports := astutil.Imports(p.fset, node)
	for _, imps := range imports {
		for _, imp := range imps {
			pkgpath := imp.Path.Value[1 : len(imp.Path.Value)-1] // unquote the value

			name := gno.LastPathElement(pkgpath)
			if pkg := p.resolver.ResolvePath(pkgpath); pkg != nil {
				name = pkg.Name()
			}

			isNamedImport := imp.Name != nil && imp.Name.Name != "_"
			if isNamedImport {
				name = imp.Name.Name
			}

			if _, ok := unresolved[name]; ok {
				delete(unresolved, name)
				continue
			}

			if isNamedImport {
				astutil.DeleteNamedImport(p.fset, node, name, pkgpath)
			} else {
				astutil.DeleteImport(p.fset, node, pkgpath)
			}
		}
	}

	// Mark knownDecls as resolved
	for ident := range knownDecls {
		delete(unresolved, ident.Name)
	}
}

// resolve tries to resolve unresolved package using `Resolver`
func (p *Processor) resolve(
	node *ast.File,
	unresolved map[string]map[string]bool,
) {
	for decl, sels := range unresolved {
		for _, pkg := range p.resolver.ResolveName(decl) {
			if !hasDeclExposed(p, sels, pkg) {
				continue
			}

			astutil.AddImport(p.fset, node, pkg.Path())
			delete(unresolved, decl)
			break
		}
	}
}

// hasDeclExposed checks if declarations are exposed in the specified path.
func hasDeclExposed(p *Processor, decls map[string]bool, pkg Package) bool {
	exposed := p.processPackageFiles(pkg.Path(), pkg)
	if exposed.error != nil {
		return false
	}

	for obj := range exposed.decls {
		if !ast.IsExported(obj.Name) {
			continue
		}

		if decls[obj.Name] {
			return true
		}
	}

	return false
}
