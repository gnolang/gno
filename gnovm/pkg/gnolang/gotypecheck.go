package gnolang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"
	"path"
	"slices"

	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
)

// type checking (using go/types)

// MemPackageGetter implements the GetMemPackage() method. It is a subset of
// [Store], separated for ease of testing.
type MemPackageGetter interface {
	GetMemPackage(path string) *std.MemPackage
}

// TypeCheckMemPackage performs type validation and checking on the given
// mpkg. To retrieve dependencies, it uses getter.
//
// The syntax checking is performed entirely using Go's go/types package.
func TypeCheckMemPackage(mpkg *std.MemPackage, getter MemPackageGetter) error {
	return typeCheckMemPackage(mpkg, getter, false)
}

// TypeCheckMemPackageTest performs the same type checks as
// [TypeCheckMemPackage], but allows re-declarations.
//
// NOTE: like TypeCheckMemPackage, this function ignores tests and filetests.
func TypeCheckMemPackageTest(mpkg *std.MemPackage, getter MemPackageGetter) error {
	return typeCheckMemPackage(mpkg, getter, true)
}

func typeCheckMemPackage(
	mpkg *std.MemPackage,
	getter MemPackageGetter,
	testing bool) error {

	var errs error
	imp := &gnoImporter{
		getter: getter,
		cache:  map[string]gnoImporterResult{},
		cfg: &types.Config{
			Error: func(err error) {
				errs = multierr.Append(errs, err)
			},
		},
		allowRedefinitions: testing,
	}
	imp.cfg.Importer = imp

	_, err := imp.parseCheckMemPackage(mpkg)
	// prefer to return errs instead of err:
	// err will generally contain only the first error encountered.
	if errs != nil {
		return errs
	}
	return err
}

type gnoImporterResult struct {
	pkg *types.Package
	err error
}

type gnoImporter struct {
	getter MemPackageGetter
	cache  map[string]gnoImporterResult
	cfg    *types.Config

	// allow symbol redefinitions? (test standard libraries)
	allowRedefinitions bool
}

// Unused, but satisfies the Importer interface.
func (g *gnoImporter) Import(path string) (*types.Package, error) {
	return g.ImportFrom(path, "", 0)
}

type importNotFoundError string

func (e importNotFoundError) Error() string { return "import not found: " + string(e) }

// ImportFrom returns the imported package for the given import
// path when imported by a package file located in dir.
func (g *gnoImporter) ImportFrom(path, _ string, _ types.ImportMode) (*types.Package, error) {
	if pkg, ok := g.cache[path]; ok {
		return pkg.pkg, pkg.err
	}
	mpkg := g.getter.GetMemPackage(path)
	if mpkg == nil {
		err := importNotFoundError(path)
		g.cache[path] = gnoImporterResult{err: err}
		return nil, err
	}
	result, err := g.parseCheckMemPackage(mpkg)
	g.cache[path] = gnoImporterResult{pkg: result, err: err}
	return result, err
}

func (g *gnoImporter) parseCheckMemPackage(mpkg *std.MemPackage) (*types.Package, error) {

	/* NOTE: Don't pre-transpile here; `gno lint` and fix the source first.
	fset, astfs, err := PretranspileToGno0p9(mpkg, false)
	if err != nil {
		return nil, err
	}
	*/

	// Add builtins file.
	file := &std.MemFile{
		Name: ".gnobuiltins.go",
		Body: fmt.Sprintf(`package %s

func istypednil(x any) bool { return false } // shim
func crossing() { } // shim
func cross[F any](fn F) F { return fn } // shim
func revive[F any](fn F) any { return nil } // shim
type realm interface{} // shim
`, mpkg.Name),
	}
	const parseOpts = parser.ParseComments | parser.DeclarationErrors | parser.SkipObjectResolution
	astf, err := parser.ParseFile(fset, path.Join(mpkg.Path, file.Name), file.Body, parseOpts)
	if err != nil {
		panic("error parsing gotypecheck gnobuiltins.go file")
	}
	astfs = append(astfs, astf)

	// Type-check pre-transpiled Gno0.9 AST in Go.
	// We don't (post)-transpile because the linter
	// is supposed to be used to write to the files.
	// (No need to support JIT transpiling for imports)
	//
	// XXX The pre pre thing for @cross.
	pkg, err := g.cfg.Check(mpkg.Path, fset, astfs, nil)
	return pkg, err
}

func deleteOldIdents(idents map[string]func(), f *ast.File) {
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		// ignore methods and init functions
		//nolint:goconst
		if !ok ||
			fd.Recv != nil ||
			fd.Name.Name == "init" {
			continue
		}
		if del := idents[fd.Name.Name]; del != nil {
			del()
		}
		decl := decl
		idents[fd.Name.Name] = func() {
			// NOTE: cannot use the index as a file may contain
			// multiple decls to be removed, so removing one would
			// make all "later" indexes wrong.
			f.Decls = slices.DeleteFunc(f.Decls,
				func(d ast.Decl) bool { return decl == d })
		}
	}
}
