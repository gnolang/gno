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

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/std"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
)

type ParseError error

// Package contains the memory package and directory path.
type Package struct {
	std.MemPackage
	Dir string
}

type Processor struct {
	stdlibs map[string][]*Package
	extlibs map[string][]*Package
}

func NewProcessor() *Processor {
	return &Processor{
		stdlibs: map[string][]*Package{},
		extlibs: map[string][]*Package{},
	}
}

func (p *Processor) LoadStdPackages(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		files, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		var gnofiles []string
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".gno" {
				gnofiles = append(gnofiles, filepath.Join(path, file.Name()))
			}
		}
		if len(gnofiles) == 0 {
			return nil
		}

		pkgname, ok := strings.CutPrefix(path, root)
		if !ok {
			return nil
		}
		memPkg := gnolang.ReadMemPackageFromList(gnofiles, strings.TrimPrefix(pkgname, "/"))

		p.stdlibs[memPkg.Name] = append(p.stdlibs[memPkg.Name], &Package{
			MemPackage: *memPkg,
			Dir:        path,
		})
		return nil
	})
}

func (p *Processor) LoadPackages(root string) error {
	mods, err := gnomod.ListPkgs(root)
	if err != nil {
		return fmt.Errorf("unable to resolve example folder: %w", err)
	}

	sorted, err := mods.Sort()
	if err != nil {
		return fmt.Errorf("unable to sort pkgs: %w", err)
	}

	for _, modPkg := range sorted.GetNonDraftPkgs() {
		memPkg := gnolang.ReadMemPackage(modPkg.Dir, modPkg.Name)
		if memPkg.Validate() != nil {
			continue
		}

		p.extlibs[memPkg.Name] = append(p.extlibs[memPkg.Name], &Package{
			MemPackage: *memPkg,
			Dir:        modPkg.Dir,
		})
	}

	return nil
}

// FormatImports processes a single Gno file and adds necessary imports.
func (p *Processor) FormatImports(filep string) ([]byte, error) {
	fset := token.NewFileSet()

	pkgDecls := make(map[string]*ast.File)
	_, err := processPackageFiles(fset, filepath.Dir(filep), pkgDecls)
	if err != nil {
		return nil, fmt.Errorf("unable to process package: %w", err)
	}

	node, ok := pkgDecls[filepath.Base(filep)]
	if !ok {
		return nil, fmt.Errorf("not a valid gno file: %s", filep)
	}

	topDecls := make(map[*ast.Object]ast.Decl)
	collectTopDeclarations(pkgDecls, topDecls)
	p.resolveAndUpdateImports(fset, node, topDecls)

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return nil, fmt.Errorf("unable to format file: %w", err)
	}

	// we let `go/imports` managing the sort of the imports
	ret, err := imports.Process(filep, buf.Bytes(), &imports.Options{
		TabWidth:   8,
		Comments:   true,
		TabIndent:  true,
		FormatOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to format import: %w", err)
	}

	return ret, nil
}

// resolveAndUpdateImports resolves and updates imports.
func (p *Processor) resolveAndUpdateImports(fset *token.FileSet, node *ast.File, topDecls map[*ast.Object]ast.Decl) {
	unresolved := collectUnresolved(node, topDecls)
	cleanupPreviousImports(fset, node, topDecls, unresolved)
	resolve(fset, node, unresolved, p.stdlibs) // first resolve stdlibs
	resolve(fset, node, unresolved, p.extlibs)
}

// processPackageFiles processes Gno package files and collects top-level declarations.
func processPackageFiles(fset *token.FileSet, root string, filesNode map[string]*ast.File) (map[*ast.Object]ast.Decl, error) {
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

		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.AllErrors)
		if err != nil {
			return fmt.Errorf("unable to process file %q: %w", path, ParseError(err))
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
			if d.Recv == nil && d.Name != nil && d.Name.Obj != nil { // Check if it's a top-level function
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
func cleanupPreviousImports(fset *token.FileSet, node *ast.File, topDecls map[*ast.Object]ast.Decl, unresolved map[string]map[string]bool) {
	imports := astutil.Imports(fset, node)

	for _, imps := range imports {
		for _, imp := range imps {
			pkgpath := imp.Path.Value[1 : len(imp.Path.Value)-1] // unquote the value
			name := filepath.Base(pkgpath)
			isNamedImport := imp.Name != nil
			if isNamedImport {
				name = imp.Name.Name
			}

			if _, ok := unresolved[name]; ok {
				delete(unresolved, name)
				continue
			}

			if isNamedImport {
				astutil.DeleteNamedImport(fset, node, name, pkgpath)
			} else {
				astutil.DeleteImport(fset, node, pkgpath)
			}
		}
	}

	for obj := range topDecls {
		delete(unresolved, obj.Name)
	}
}

// resolve tries to resolve unresolved package based on a list of pkgs.
func resolve(
	fset *token.FileSet,
	node *ast.File,
	unresolved map[string]map[string]bool,
	pkgs map[string][]*Package,
) {
	for decl, sels := range unresolved {
		if listPkgs, ok := pkgs[decl]; ok {
			for _, pkg := range listPkgs {
				if !hasDeclExposed(fset, sels, pkg.Dir) {
					continue
				}

				astutil.AddImport(fset, node, pkg.Path)
				delete(unresolved, decl)
				break
			}
		}
	}
}

// hasDeclExposed checks if declarations are exposed in the specified path.
func hasDeclExposed(fset *token.FileSet, decls map[string]bool, path string) bool {
	filesNode := make(map[string]*ast.File)
	exposed, err := processPackageFiles(fset, path, filesNode)
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
