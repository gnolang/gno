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
	"os/exec"
	"path/filepath"
	"reflect"
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

// for now a simple call to fmt.Fprintf, but could be improved
func logWarning(format string, v ...any) {
	// TODO: add these at the top of the generated file as a comment
	// so that if there are exceptions to make these are made
	// consciously in code review.
	fmt.Fprintf(os.Stderr, "warn: "+format+"\n", v...)
}

func _main(stdlibsPath string) error {
	stdlibsPath = filepath.Clean(stdlibsPath)
	if s, err := os.Stat(stdlibsPath); err != nil {
		return err
	} else if !s.IsDir() {
		return fmt.Errorf("not a directory: %q", stdlibsPath)
	}

	pkgs, err := walkStdlibs(stdlibsPath)
	if err != nil {
		return err
	}

	// Create mappings.
	mappings := linkFunctions(pkgs)

	// Create file.
	f, err := os.Create("native.go")
	if err != nil {
		return fmt.Errorf("create native.go: %w", err)
	}
	defer f.Close()

	// Execute template
	td := &tplData{
		Mappings: mappings,
	}
	td.generateLibnums()
	if err := tpl.Execute(f, td); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	if err := f.Close(); err != nil {
		return err
	}

	cmd := exec.Command(
		"go", "run", "-modfile", "../../misc/devdeps/go.mod",
		"mvdan.cc/gofumpt", "-w", "native.go",
	)
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("error executing gofumpt: %w", err)
	}

	return nil
}

type pkgData struct {
	importPath  string
	fsDir       string
	gnoBodyless []*ast.FuncDecl
	goExported  []*ast.FuncDecl
}

func walkStdlibs(stdlibsPath string) ([]*pkgData, error) {
	pkgs := make([]*pkgData, 0, 64)
	err := filepath.WalkDir(stdlibsPath, func(fpath string, d fs.DirEntry, err error) error {
		// skip dirs and top-level directory.
		if d.IsDir() || filepath.Dir(fpath) == stdlibsPath {
			return nil
		}

		// skip non-source files.
		ext := filepath.Ext(fpath)
		if ext != ".go" && ext != ".gno" {
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
			exp := resolveGnoMachine(f, filterExported(f))
			if len(exp) > 0 {
				pkg.goExported = append(pkg.goExported, exp...)
			}
		} else if bd := filterBodylessFuncDecls(f); len(bd) > 0 {
			// gno file -- keep track of function declarations without body.
			pkg.gnoBodyless = append(pkg.gnoBodyless, bd...)
		}
		return nil
	})
	return pkgs, err
}

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

func resolveGnoMachine(f *ast.File, fns []*ast.FuncDecl) []*ast.FuncDecl {
	iname := gnolangImportName(f)
	if iname == "" {
		return fns
	}
	for _, fn := range fns {
		if len(fn.Type.Params.List) == 0 {
			continue
		}
		first := fn.Type.Params.List[0]
		if len(first.Names) > 1 {
			continue
		}
		ind, ok := first.Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		res, ok := ind.X.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if id, ok := res.X.(*ast.Ident); ok && id.Name == iname && res.Sel.Name == "Machine" {
			id.Name = "#gnomachine"
		}
	}
	return fns
}

func gnolangImportName(f *ast.File) string {
	for _, i := range f.Imports {
		ipath, err := strconv.Unquote(i.Path.Value)
		if err != nil {
			continue
		}
		if ipath == "github.com/gnolang/gno/gnovm/pkg/gnolang" {
			if i.Name == nil {
				return "gnolang"
			}
			return i.Name.Name
		}
	}
	return ""
}

type mapping struct {
	GnoImportPath  string
	GnoMethod      string
	GnoParamTypes  []string
	GnoResultTypes []string
	GoImportPath   string
	GoFunc         string
	MachineParam   bool
}

func linkFunctions(pkgs []*pkgData) []mapping {
	var mappings []mapping
	for _, pkg := range pkgs {
		for _, gb := range pkg.gnoBodyless {
			nameWant := gb.Name.Name
			if !gb.Name.IsExported() {
				nameWant = "X_" + nameWant
			}
			fn := findFuncByName(pkg.goExported, nameWant)
			if fn == nil {
				logWarning("package %q: no matching go function declaration (%q) exists for function %q",
					pkg.importPath, nameWant, gb.Name.Name)
				continue
			}
			mp := mapping{
				GnoImportPath: pkg.importPath,
				GnoMethod:     gb.Name.Name,
				GoImportPath:  "github.com/gnolang/gno/gnovm/stdlibs/" + pkg.importPath,
				GoFunc:        fn.Name.Name,
			}
			if !mp.loadSignaturesMatch(gb, fn) {
				logWarning("package %q: signature of gno function %s doesn't match signature of go function %s",
					pkg.importPath, gb.Name.Name, fn.Name.Name)
				continue
			}
			mappings = append(mappings, mp)
		}
	}
	return mappings
}

func findFuncByName(fns []*ast.FuncDecl, name string) *ast.FuncDecl {
	for _, fn := range fns {
		if fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

func (m *mapping) loadSignaturesMatch(gnof, gof *ast.FuncDecl) bool {
	if gnof.Type.TypeParams != nil || gof.Type.TypeParams != nil {
		panic("type parameters not supported")
	}
	// Ideally, signatures match when they accept the same types,
	// or aliases.
	gnop, gnor := fieldListToTypes(gnof.Type.Params), fieldListToTypes(gnof.Type.Results)
	// store gno params and results in mapping
	m.GnoParamTypes = gnop
	m.GnoResultTypes = gnor
	gop, gor := fieldListToTypes(gof.Type.Params), fieldListToTypes(gof.Type.Results)
	if len(gop) > 0 && gop[0] == "*gno.Machine" {
		m.MachineParam = true
		gop = gop[1:]
	}
	return reflect.DeepEqual(gnop, gop) && reflect.DeepEqual(gnor, gor)
}

var builtinTypes = [...]string{
	"bool",
	"string",
	"int",
	"int8",
	"int16",
	"rune",
	"int32",
	"int64",
	"uint",
	"byte",
	"uint8",
	"uint16",
	"uint32",
	"uint64",
	"bigint",
	"float32",
	"float64",
	"error",
	"any",
}

func validIdent(name string) bool {
	for _, t := range builtinTypes {
		if name == t {
			return true
		}
	}
	return false
}

func exprToString(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.Ident:
		if !validIdent(e.Name) {
			panic(fmt.Sprintf("ident is not builtin: %q", e.Name))
		}
		return e.Name
	case *ast.Ellipsis:
		return "..." + exprToString(e.Elt)
	case *ast.SelectorExpr:
		if x, ok := e.X.(*ast.Ident); ok && x.Name == "#gnomachine" {
			return "gno.Machine"
		}
		panic("SelectorExpr not supported")
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		var ls string
		if e.Len != nil {
			switch e.Len.(type) {
			case *ast.Ellipsis:
				ls = "..."
			}
		}
		return "[" + ls + "]" + exprToString(e.Elt)
	case *ast.StructType:
		if len(e.Fields.List) > 0 {
			panic("structs with values not supported yet")
		}
		return "struct{}"
	case *ast.FuncType:
		return "func(" + strings.Join(fieldListToTypes(e.Params), ", ") + ")" + strings.Join(fieldListToTypes(e.Results), ", ")
	case *ast.InterfaceType:
		if len(e.Methods.List) > 0 {
			panic("interfaces with methods not supported yet")
		}
		return "interface{}"
	case *ast.MapType:
		return "map[" + exprToString(e.Key) + "]" + exprToString(e.Value)
	default:
		panic(fmt.Sprintf("invalid expression as func param/return type: %T", e))
	}
}

func fieldListToTypes(fl *ast.FieldList) []string {
	if fl == nil {
		return nil
	}
	r := make([]string, 0, len(fl.List))
	for _, f := range fl.List {
		ts := exprToString(f.Type)
		times := len(f.Names)
		if times == 0 {
			// case of unnamed params; such as return values (often)
			times = 1
		}
		for i := 0; i < times; i++ {
			r = append(r, ts)
		}
	}
	return r
}

type tplData struct {
	Mappings []mapping
	LibNums  []string
}

//go:embed template.tmpl
var templateText string

var tpl = template.Must(template.New("").Parse(templateText))

func (t tplData) FindLibNum(s string) (int, error) {
	for i, v := range t.LibNums {
		if v == s {
			return i, nil
		}
	}
	return -1, fmt.Errorf("could not find lib: %q", s)
}

func (t *tplData) generateLibnums() {
	var last string
	for _, m := range t.Mappings {
		if m.GoImportPath != last {
			t.LibNums = append(t.LibNums, m.GoImportPath)
			last = m.GoImportPath
		}
	}
}
