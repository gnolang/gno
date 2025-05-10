package gnolang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"slices"

	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
)

// Type-checking (using go/types)

// MemPackageGetter implements the GetMemPackage() method. It is a subset of
// [Store], separated for ease of testing.
type MemPackageGetter interface {
	GetMemPackage(path string) *std.MemPackage
}

// TypeCheckMemPackage performs type validation and checking on the given
// mpkg. To retrieve dependencies, it uses getter.
//
// The syntax checking is performed entirely using Go's go/types package.
// TODO: rename these to GoTypeCheck*, goTypeCheck*...
func TypeCheckMemPackage(mpkg *std.MemPackage, getter MemPackageGetter) (
	pkg *types.Package, fset *token.FileSet, astfs []*ast.File, errs error) {

	return typeCheckMemPackage(mpkg, getter)
}

func typeCheckMemPackage(mpkg *std.MemPackage, getter MemPackageGetter) (
	pkg *types.Package, fset *token.FileSet, astfs []*ast.File, errs error) {

	gimp := &gnoImporter{
		getter: getter,
		cache:  map[string]gnoImporterResult{},
		cfg: &types.Config{
			Error: func(err error) {
				errs = multierr.Append(errs, err)
			},
		},
	}
	gimp.cfg.Importer = gimp

	wtests := true // type check all .gno files.
	pkg, fset, astfs, errs = gimp.typeCheckMemPackage(mpkg, wtests)
	return
}

type gnoImporterResult struct {
	pkg *types.Package
	err error
}

// gimp.
// gimp type checks.
// gimp remembers.
// gimp.
type gnoImporter struct {
	getter MemPackageGetter
	cache  map[string]gnoImporterResult
	cfg    *types.Config
}

// Unused, but satisfies the Importer interface.
func (gimp *gnoImporter) Import(path string) (*types.Package, error) {
	return gimp.ImportFrom(path, "", 0)
}

type importNotFoundError string

func (e importNotFoundError) Error() string { return "import not found: " + string(e) }

// ImportFrom returns the imported package for the given import
// path when imported by a package file located in dir.
func (g *gnoImporter) ImportFrom(path, _ string, _ types.ImportMode) (*types.Package, error) {
	if pkg, ok := g.cache[path]; ok {
		return pkg.pkg, pkg.err
	}
	// fmt.Println("GNOIMPORTER IMPORTFROM > GETMEMPACKAGE", path)
	mpkg := g.getter.GetMemPackage(path)
	if mpkg == nil {
		err := importNotFoundError(path)
		g.cache[path] = gnoImporterResult{err: err}
		return nil, err
	}
	wtests := false // don't parse test files for imports.
	pkg, _, _, errs := g.typeCheckMemPackage(mpkg, wtests)
	g.cache[path] = gnoImporterResult{pkg: pkg, err: errs}
	return pkg, errs
}

// Assumes that the code is Gno 0.9.
// If not, first use `gno lint` to transpile the code.
// Returns parsed *types.Package, *token.FileSet, []*ast.File.
//
// Args:
//   - wtests: if true, with all *_test.gno and *_testfile.gno files.
func (g *gnoImporter) typeCheckMemPackage(mpkg *std.MemPackage, wtests bool) (
	pkg *types.Package, fset *token.FileSet, astfs []*ast.File, errs error) {

	// STEP 1: Check gno.mod version.
	_, outdated := ParseGnoMod(mpkg)
	if outdated {
		return nil, nil, nil, fmt.Errorf("outdated gno version for package %s", mpkg.Path)
	}

	// STEP 2: Parse the mem package to Go AST.
	fset, astfs, errs = GoParseMemPackage(mpkg, wtests)
	if errs != nil {
		return nil, nil, nil, fmt.Errorf("go parsing mem package: %v", errs)
	}

	// STEP 2: Add .gnobuiltins.go file.
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

	// STEP 2: Parse .gnobuiltins.go file.
	const parseOpts = parser.ParseComments |
		parser.DeclarationErrors |
		parser.SkipObjectResolution
	var astf, err = parser.ParseFile(
		fset,
		path.Join(mpkg.Path, file.Name),
		file.Body,
		parseOpts)
	if err != nil {
		panic("error parsing gotypecheck gnobuiltins.go file")
	}
	astfs = append(astfs, astf)

	// STEP 3: Type-check Gno0.9 AST in Go.
	pkg, errs = g.cfg.Check(mpkg.Path, fset, astfs, nil)
	return pkg, fset, astfs, errs
}

func deleteOldIdents(idents map[string]func(), astf *ast.File) {
	for _, decl := range astf.Decls {
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
			astf.Decls = slices.DeleteFunc(astf.Decls,
				func(d ast.Decl) bool { return decl == d })
		}
	}
}
