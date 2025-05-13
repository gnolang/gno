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

/*
	Type-checking (using go/types).
	Refer to the [Lint and Transpile ADR](./adr/pr4264_lint_transpile.md).
*/

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
	pkg *types.Package, gofset *token.FileSet, gofs, _gofs, tgofs []*ast.File, errs error) {
	var gimp *gnoImporter
	gimp = &gnoImporter{
		getter: getter,
		cache:  map[string]gnoImporterResult{},
		cfg: &types.Config{
			Error: func(err error) {
				gimp.Error(err)
			},
		},
		errors: nil,
	}
	gimp.cfg.Importer = gimp

	all := true    // type check all .gno files for mpkg (not for imports).
	strict := true // check gno.mod exists
	pkg, gofset, gofs, _gofs, tgofs, errs = gimp.typeCheckMemPackage(mpkg, all, strict)
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
	errors error // multierr
}

// Unused, but satisfies the Importer interface.
func (gimp *gnoImporter) Import(path string) (*types.Package, error) {
	return gimp.ImportFrom(path, "", 0)
}

// Pass through to cfg.Error for collecting all type-checking errors.
func (gimp *gnoImporter) Error(err error) {
	gimp.errors = multierr.Append(gimp.errors, err)
}

type importNotFoundError string

func (e importNotFoundError) Error() string { return "import not found: " + string(e) }

// ImportFrom returns the imported package for the given import
// path when imported by a package file located in dir.
func (gimp *gnoImporter) ImportFrom(path, _ string, _ types.ImportMode) (*types.Package, error) {
	if pkg, ok := gimp.cache[path]; ok {
		return pkg.pkg, pkg.err
	}
	mpkg := gimp.getter.GetMemPackage(path)
	if mpkg == nil {
		err := importNotFoundError(path)
		gimp.cache[path] = gnoImporterResult{err: err}
		return nil, err
	}
	all := false    // don't parse test files for imports.
	strict := false // don't check for gno.mod for imports.
	pkg, _, _, _, _, errs := gimp.typeCheckMemPackage(mpkg, all, strict)
	gimp.cache[path] = gnoImporterResult{pkg: pkg, err: errs}
	return pkg, errs
}

// Assumes that the code is Gno 0.9.
// If not, first use `gno lint` to transpile the code.
// Returns parsed *types.Package, *token.FileSet, []*ast.File.
//
// Args:
//   - all: If true add all *_test.gno and *_testfile.gno files.
//     Generally should be set to false when importing because
//     tests cannot be imported and used anyways.
//   - strict: If true errors on gno.mod version mismatch.
func (gimp *gnoImporter) typeCheckMemPackage(mpkg *std.MemPackage, all bool, strict bool) (
	pkg *types.Package, gofset *token.FileSet, gofs, _gofs, tgofs []*ast.File, errs error) {

	// See adr/pr4264_lint_transpile.md
	// STEP 2: Check gno.mod version.
	if strict {
		_, err := ParseCheckGnoMod(mpkg)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}

	// STEP 3: Parse the mem package to Go AST.
	gofset, gofs, _gofs, tgofs, errs = GoParseMemPackage(mpkg, all)
	if errs != nil {
		return nil, nil, nil, nil, nil, errs
	}
	if !all && (len(_gofs) > 0 || len(tgofs) > 0) {
		panic("unexpected test files from GoParseMemPackage()")
	}

	// STEP 3: Add .gnobuiltins.go file.
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

	// STEP 3: Parse .gnobuiltins.go file.
	const parseOpts = parser.ParseComments |
		parser.DeclarationErrors |
		parser.SkipObjectResolution
	var gmgof, err = parser.ParseFile(
		gofset,
		path.Join(mpkg.Path, file.Name),
		file.Body,
		parseOpts)
	if err != nil {
		panic("error parsing gotypecheck gnobuiltins.go file")
	}

	// STEP 4: Type-check Gno0.9 AST in Go (normal and _test.gno if all).
	gofs = append(gofs, gmgof)
	pkg, _ = gimp.cfg.Check(mpkg.Path, gofset, gofs, nil)

	// STEP 4: Type-check Gno0.9 AST in Go (xxx_test package if all).
	// Each integration test is its own package.
	for _, _gof := range _gofs {
		gmgof.Name = _gof.Name // copy _test package name to gno.mod
		gofs2 := []*ast.File{gmgof, _gof}
		_, _ = gimp.cfg.Check(mpkg.Path, gofset, gofs2, nil)
	}

	// STEP 4: Type-check Gno0.9 AST in Go (_filetest.gno if all).
	// Each filetest is its own package.
	for _, tgof := range tgofs {
		gmgof.Name = tgof.Name // copy _filetest.gno package name to gno.mod
		gofs2 := []*ast.File{gmgof, tgof}
		_, _ = gimp.cfg.Check(mpkg.Path, gofset, gofs2, nil)
	}
	return pkg, gofset, gofs, _gofs, tgofs, gimp.errors
}

func deleteOldIdents(idents map[string]func(), gof *ast.File) {
	for _, decl := range gof.Decls {
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
			gof.Decls = slices.DeleteFunc(gof.Decls,
				func(d ast.Decl) bool { return decl == d })
		}
	}
}
