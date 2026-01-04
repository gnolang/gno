// Command genstd is a code-generator to create meta-information for the Gno
// standard libraries.
//
// All the packages in the standard libraries are parsed and relevant
// information is collected; the file is then generated followingthe template
// available in ./template.tmpl
//
// genstd is responsible for linking natively bound functions, a FFI from Gno
// functions to Go implementations, and calculating the initialization order
// of the standard libraries.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	_ "embed"
)

var skipInitOrder = flag.Bool("skip-init-order", false, "skip generating packages initialization order.")

func main() {
	flag.Parse()
	path := "."
	if a := flag.Arg(0); a != "" {
		path = a
	}
	if err := _main(path); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

const outputFile = "generated.go"

func _main(stdlibsPath string) error {
	stdlibsPath = filepath.Clean(stdlibsPath)
	if s, err := os.Stat(stdlibsPath); err != nil {
		return err
	} else if !s.IsDir() {
		return fmt.Errorf("not a directory: %q", stdlibsPath)
	}

	// Gather data about each package, getting functions of interest
	// (gno bodyless + go exported).
	pkgs, err := walkStdlibs(stdlibsPath)
	if err != nil {
		return err
	}

	// Link up each Gno function with its matching Go function.
	mappings := linkFunctions(pkgs)
	var initOrder []string
	if !*skipInitOrder {
		initOrder = sortPackages(pkgs)
	}

	// Create generated file.
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create "+outputFile+": %w", err)
	}
	defer f.Close()

	// Execute template.
	td := &tplData{
		Mappings:  mappings,
		InitOrder: initOrder,
	}
	if err := tpl.Execute(f, td); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	if err := f.Close(); err != nil {
		return err
	}

	// gofumpt doesn't do "import fixing" like goimports:
	// https://github.com/mvdan/gofumpt#frequently-asked-questions
	if err := runTool("golang.org/x/tools/cmd/goimports"); err != nil {
		return err
	}
	return runTool("mvdan.cc/gofumpt")
}

type pkgData struct {
	importPath string
	fsDir      string

	// for matching native functions
	gnoBodyless []funcDecl
	goExported  []funcDecl

	// for determining initialization order
	imports map[string]struct{}

	// whether there are gno files in this package; if not, it's not a valid gno
	// package and should be ignored ie. in the initialization order.
	hasGno bool
}

type funcDecl struct {
	*ast.FuncDecl
	imports []*ast.ImportSpec
}

func addImports(fds []*ast.FuncDecl, imports []*ast.ImportSpec) []funcDecl {
	r := make([]funcDecl, len(fds))
	for i, fd := range fds {
		r[i] = funcDecl{fd, imports}
	}
	return r
}

// walkStdlibs does walks through the given directory, expected to be a
// "stdlib" directory, parsing and keeping track of Go and Gno functions of
// interest.
func walkStdlibs(stdlibsPath string) ([]*pkgData, error) {
	pkgs := make([]*pkgData, 0, 64)
	err := WalkDir(stdlibsPath, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// skip dirs and top-level directory.
		if d.IsDir() || filepath.Dir(fpath) == stdlibsPath {
			return nil
		}

		// skip non-source and test files.
		ext := filepath.Ext(fpath)
		noExt := fpath[:len(fpath)-len(ext)]
		if (ext != ".go" && ext != ".gno") ||
			strings.HasSuffix(noExt, "_test") ||
			strings.HasSuffix(fpath, ".gen.go") {
			return nil
		}

		dir := filepath.Dir(fpath)
		var pkg *pkgData
		// if we've already been in this directory in a previous file, it must
		// be in the last entry of pkgs, as all files in a directory are
		// processed together.
		if len(pkgs) == 0 || pkgs[len(pkgs)-1].fsDir != dir {
			pkg = &pkgData{
				importPath: strings.TrimPrefix(strings.ReplaceAll(dir, string(filepath.Separator), "/"), stdlibsPath+"/"),
				fsDir:      dir,
				imports:    make(map[string]struct{}),
			}
			pkgs = append(pkgs, pkg)
		} else {
			pkg = pkgs[len(pkgs)-1]
		}
		fs := token.NewFileSet()
		f, err := parser.ParseFile(fs, fpath, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}

		if ext == ".go" {
			// keep track of exported function declarations.
			// warn about all exported type, const and var declarations.
			if exp := filterExported(f); len(exp) > 0 {
				pkg.goExported = append(pkg.goExported, addImports(exp, f.Imports)...)
			}
			return nil
		}

		// this is a gno file; ensure to mark that there are gno files in this
		// package.
		pkg.hasGno = true

		if bd := filterBodylessFuncDecls(f); len(bd) > 0 {
			// gno file -- keep track of function declarations without body.
			pkg.gnoBodyless = append(pkg.gnoBodyless, addImports(bd, f.Imports)...)
		}
		for _, imp := range f.Imports {
			impVal := mustUnquote(imp.Path.Value)
			pkg.imports[impVal] = struct{}{}
		}

		return nil
	})
	// Remove packages which don't have gno files within.
	pkgs = slices.DeleteFunc(pkgs, func(p *pkgData) bool {
		return !p.hasGno
	})
	return pkgs, err
}

// filterBodylessFuncDecls returns the function declarations in the given file
// which don't contain a body.
func filterBodylessFuncDecls(f *ast.File) (bodyless []*ast.FuncDecl) {
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body != nil {
			continue
		}
		bodyless = append(bodyless, fd)
	}
	return
}

// filterExported returns the exported function declarations of the given file.
func filterExported(f *ast.File) (exported []*ast.FuncDecl) {
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			// TODO: complain if there are exported types/vars/consts
			continue
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				exported = append(exported, d)
			}
		}
	}
	return
}

//go:embed template.tmpl
var templateText string

var tpl = template.Must(template.New("").Parse(templateText))

// tplData is the data passed to the template.
type tplData struct {
	Mappings  []mapping
	InitOrder []string
}

type tplImport struct{ Name, Path string }

// Imports returns the packages that the resulting generated files should import.
func (t tplData) Imports() (res []tplImport) {
	add := func(path string) {
		for _, v := range res {
			if v.Path == path {
				return
			}
		}
		res = append(res, tplImport{Name: pkgNameFromPath(path), Path: path})
	}
	for _, m := range t.Mappings {
		add(m.GoImportPath)
		// There might be a bit more than we need - but we run goimports to fix that.
		for _, v := range m.goImports {
			add(mustUnquote(v.Path.Value))
		}
	}
	return
}

func (tplData) PkgName(path string) string { return pkgNameFromPath(path) }
