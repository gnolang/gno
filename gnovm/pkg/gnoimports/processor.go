package gnoimports

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
)

const tabWidth = 8

type ParseError error

type Package struct {
	Path string
	Name string
	Dir  string
}

type Processor struct {
	resolver Resolver
	fset     *token.FileSet

	// cache parsed file
	fcache map[string]*ast.File
}

func NewProcessor(r Resolver) *Processor {
	return &Processor{
		resolver: r,
		fset:     token.NewFileSet(),
		fcache:   map[string]*ast.File{},
	}
}

// FormatImportFromSource processes a single Gno file and adds necessary imports.
// FormatImportFromSource parse and format the source from src. The type of the argument
// for the src parameter must be string, []byte, or [io.Reader].
func (p *Processor) FormatImportFromSource(filename string, src any) ([]byte, error) {
	if src == nil {
		return nil, fmt.Errorf("source input cannot be nil")
	}

	// Parse the source file
	node, err := p.parseFile(filename, src)
	if err != nil {
		return nil, fmt.Errorf("unable to parse source: %w", ParseError(err))
	}

	pkgDecls := make(map[string]*ast.File)
	// Process and format the parsed node.
	return p.processAndFormat(node, filename, pkgDecls)
}

// FormatImportFromFile processes a single Gno file, format it and adds necessary imports.
func (p *Processor) FormatImportFromFile(filep string) ([]byte, error) {
	pkgDecls := make(map[string]*ast.File)

	// Process package files.
	_, err := p.processPackageFiles(filepath.Dir(filep), pkgDecls)
	if err != nil {
		return nil, fmt.Errorf("unable to process package: %w", err)
	}

	// Retrieve the node for the file.
	node, ok := pkgDecls[filepath.Base(filep)]
	if !ok {
		return nil, fmt.Errorf("not a valid gno file: %s", filep)
	}

	// Process and format the parsed node.
	return p.processAndFormat(node, filep, pkgDecls)
}

func (p *Processor) parseFile(filename string, src any) (file *ast.File, err error) {
	var ok bool
	if file, ok = p.fcache[filename]; !ok {
		// Parse the source file
		file, err = parser.ParseFile(p.fset, filename, src, parser.ParseComments|parser.AllErrors)
		if err != nil {
			return nil, fmt.Errorf("unable to parse file %q: %w", filename, ParseError(err))
		}

		p.fcache[filename] = file
	}

	return file, nil
}

// Helper function to process and format a parsed AST node.
func (p *Processor) processAndFormat(file *ast.File, filename string, pkgDecls map[string]*ast.File) ([]byte, error) {
	// Collect top declarations.
	topDecls := make(map[*ast.Object]ast.Decl)
	collectTopDeclarations(pkgDecls, topDecls)

	// Collect unresolved
	unresolved := collectUnresolved(file, topDecls)

	p.cleanupPreviousImports(file, topDecls, unresolved)
	p.resolve(file, unresolved)

	// Buffer to store formatted output.
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, p.fset, file); err != nil {
		return nil, fmt.Errorf("unable to format file: %w", err)
	}

	// Use go/imports for formating and managing import sorting.
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
func (p *Processor) processPackageFiles(root string, filesNode map[string]*ast.File) (map[*ast.Object]ast.Decl, error) {
	declmap := make(map[*ast.Object]ast.Decl)
	return declmap, filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("unable to walk on %q: %w", path, err)
		}

		if info.IsDir() {
			if path == root {
				return nil
			}

			return filepath.SkipDir
		}

		filename := info.Name()
		if strings.HasPrefix(filename, ".") || filepath.Ext(path) != ".gno" {
			return nil
		}

		file, err := p.parseFile(path, nil)
		if err != nil {
			return ParseError(err)
		}

		collectTopDeclaration(file, declmap)
		filesNode[filename] = file
		return nil
	})
}

// collectTopDeclarations collects top-level declarations from package files.
func collectTopDeclarations(pkgDecls map[string]*ast.File, topDecls map[*ast.Object]ast.Decl) {
	for _, file := range pkgDecls {
		collectTopDeclaration(file, topDecls)
	}
}

// collectTopDeclaration collects top-level declarations from a single file.
func collectTopDeclaration(file *ast.File, topDecls map[*ast.Object]ast.Decl) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					topDecls[s.Name.Obj] = d
				case *ast.ValueSpec:
					for _, name := range s.Names {
						topDecls[name.Obj] = d
					}
				}
			}
		case *ast.FuncDecl:
			// Check for top-level function
			if d.Recv == nil && d.Name != nil && d.Name.Obj != nil {
				topDecls[d.Name.Obj] = d
			}
		}
	}
}

// collectUnresolved collects unresolved identifiers and declarations.
func collectUnresolved(file *ast.File, topDecls map[*ast.Object]ast.Decl) map[string]map[string]bool {
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
			if d := topDecls[e.Obj]; d != nil {
				delete(unresolved, e.Name)
			}

			return true
		case *ast.SelectorExpr:
			for _, u := range unresolvedList {
				if u == e.X {
					ident := e.X.(*ast.Ident)
					unresolved[ident.Name][e.Sel.Name] = true
					break
				}
			}

			return true
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
func (p *Processor) cleanupPreviousImports(node *ast.File, knownDecls map[*ast.Object]ast.Decl, unresolved map[string]map[string]bool) {
	imports := astutil.Imports(p.fset, node)
	for _, imps := range imports {
		for _, imp := range imps {
			pkgpath := imp.Path.Value[1 : len(imp.Path.Value)-1] // unquote the value

			name := filepath.Base(pkgpath)
			if pkg := p.resolver.ResolvePath(pkgpath); pkg != nil {
				name = pkg.Name
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
	for obj := range knownDecls {
		delete(unresolved, obj.Name)
	}
}

// resolve tries to resolve unresolved package using `Resolver`
func (p *Processor) resolve(
	node *ast.File,
	unresolved map[string]map[string]bool,
) {
	for decl, sels := range unresolved {
		for _, pkg := range p.resolver.ResolveName(decl) {
			if !hasDeclExposed(p, sels, pkg.Dir) {
				continue
			}

			astutil.AddImport(p.fset, node, pkg.Path)
			delete(unresolved, decl)
			break
		}
	}
}

// hasDeclExposed checks if declarations are exposed in the specified path.
func hasDeclExposed(p *Processor, decls map[string]bool, path string) bool {
	filesNode := make(map[string]*ast.File)
	exposed, err := p.processPackageFiles(path, filesNode)
	if err != nil {
		return false
	}

	for obj := range exposed {
		if !ast.IsExported(obj.Name) {
			continue
		}

		if decls[obj.Name] {
			return true
		}
	}

	return false
}
