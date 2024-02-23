// Command genstd provides static code generation for standard library native
// bindings.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	_ "embed"
)

func main() {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	if err := _main(path); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

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

	// Create generated file.
	f, err := os.Create("native.go")
	if err != nil {
		return fmt.Errorf("create native.go: %w", err)
	}
	defer f.Close()

	// Execute template.
	td := &tplData{
		Mappings: mappings,
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
	importPath  string
	fsDir       string
	gnoBodyless []funcDecl
	goExported  []funcDecl
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

// walkStdlibs does a BFS walk through the given directory, expected to be a
// "stdlib" directory, parsing and keeping track of Go and Gno functions of
// interest.
func walkStdlibs(stdlibsPath string) ([]*pkgData, error) {
	pkgs := make([]*pkgData, 0, 64)
	err := filepath.WalkDir(stdlibsPath, func(fpath string, d fs.DirEntry, err error) error {
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
		// because of bfs, we know that if we've already been in this directory
		// in a previous file, it must be in the last entry of pkgs.
		if len(pkgs) == 0 || pkgs[len(pkgs)-1].fsDir != dir {
			pkg = &pkgData{
				importPath: strings.ReplaceAll(strings.TrimPrefix(dir, stdlibsPath+"/"), string(filepath.Separator), "/"),
				fsDir:      dir,
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
		} else if bd := filterBodylessFuncDecls(f); len(bd) > 0 {
			// gno file -- keep track of function declarations without body.
			pkg.gnoBodyless = append(pkg.gnoBodyless, addImports(bd, f.Imports)...)
		}
		return nil
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
	Mappings []mapping
}

type tplImport struct{ Name, Path string }

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
			s, err := strconv.Unquote(v.Path.Value)
			if err != nil {
				panic(fmt.Errorf("could not unquote go import string literal: %s", v.Path.Value))
			}
			add(s)
		}
	}
	return
}

func (tplData) PkgName(path string) string { return pkgNameFromPath(path) }
