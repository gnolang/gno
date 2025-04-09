package typecheck

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"slices"
	"strings"

	"github.com/gnolang/gno/gnovm"
	"go.uber.org/multierr"
)

// type checking (using go/types)

// MemPackageGetter implements the GetMemPackage() method. It is a subset of
// [Store], separated for ease of testing.
type MemPackageGetter interface {
	GetMemPackage(path string) *gnovm.MemPackage
}

func CheckMemPackage(mempkg *gnovm.MemPackage, getter MemPackageGetter, testing, format bool) error {
	var errs error
	imp := &GnoImporter{
		getter: getter,
		cache:  map[string]GnoImporterResult{},
		cfg: &types.Config{
			Error: func(err error) {
				errs = multierr.Append(errs, err)
			},
		},
		allowRedefinitions: testing,
	}
	imp.cfg.Importer = imp

	_, err := imp.ParseCheckMemPackage(mempkg, format)
	// prefer to return errs instead of err:
	// err will generally contain only the first error encountered.
	if errs != nil {
		return errs
	}
	return err
}

type GnoImporterResult struct {
	pkg *types.Package
	err error
}

type GnoImporter struct {
	getter MemPackageGetter
	cache  map[string]GnoImporterResult
	cfg    *types.Config

	// allow symbol redefinitions? (test standard libraries)
	allowRedefinitions bool
}

// Unused, but satisfies the Importer interface.
func (g *GnoImporter) Import(path string) (*types.Package, error) {
	return g.ImportFrom(path, "", 0)
}

type importNotFoundError string

func (e importNotFoundError) Error() string { return "import not found: " + string(e) }

// ImportFrom returns the imported package for the given import
// path when imported by a package file located in dir.
func (g *GnoImporter) ImportFrom(path, _ string, _ types.ImportMode) (*types.Package, error) {
	if pkg, ok := g.cache[path]; ok {
		return pkg.pkg, pkg.err
	}
	mpkg := g.getter.GetMemPackage(path)
	if mpkg == nil {
		err := importNotFoundError(path)
		g.cache[path] = GnoImporterResult{err: err}
		return nil, err
	}
	fmt := false
	result, err := g.ParseCheckMemPackage(mpkg, fmt)
	g.cache[path] = GnoImporterResult{pkg: result, err: err}
	return result, err
}

func (g *GnoImporter) ParseCheckMemPackage(mpkg *gnovm.MemPackage, fmt bool) (*types.Package, error) {
	// This map is used to allow for function re-definitions, which are allowed
	// in Gno (testing context) but not in Go.
	// This map links each function identifier with a closure to remove its
	// associated declaration.
	var delFunc map[string]func()
	if g.allowRedefinitions {
		delFunc = make(map[string]func())
	}

	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(mpkg.Files))
	var errs error
	for _, file := range mpkg.Files {
		// Ignore non-gno files.
		// TODO: support filetest type checking. (should probably handle as each its
		// own separate pkg, which should also be typechecked)
		if !strings.HasSuffix(file.Name, ".gno") ||
			strings.HasSuffix(file.Name, "_test.gno") ||
			strings.HasSuffix(file.Name, "_filetest.gno") {
			continue
		}

		const parseOpts = parser.ParseComments | parser.DeclarationErrors | parser.SkipObjectResolution
		f, err := parser.ParseFile(fset, path.Join(mpkg.Path, file.Name), file.Body, parseOpts)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}

		if delFunc != nil {
			deleteOldIdents(delFunc, f)
		}

		// enforce formatting
		if fmt {
			var buf bytes.Buffer
			err = format.Node(&buf, fset, f)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			file.Body = buf.String()
		}

		files = append(files, f)
	}
	if errs != nil {
		return nil, errs
	}

	return g.cfg.Check(mpkg.Path, fset, files, nil)
}

func deleteOldIdents(idents map[string]func(), f *ast.File) {
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv != nil { // ignore methods
			continue
		}
		if del := idents[fd.Name.Name]; del != nil {
			del()
		}
		decl := decl
		idents[fd.Name.Name] = func() {
			// NOTE: cannot use the index as a file may contain multiple decls to be removed,
			// so removing one would make all "later" indexes wrong.
			f.Decls = slices.DeleteFunc(f.Decls, func(d ast.Decl) bool { return decl == d })
		}
	}
}
