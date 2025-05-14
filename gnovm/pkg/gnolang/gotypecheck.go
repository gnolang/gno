package gnolang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"slices"
	"strings"

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
func TypeCheckMemPackage(mpkg *std.MemPackage, getter MemPackageGetter) (
	pkg *types.Package, gofset *token.FileSet, gofs, _gofs, tgofs []*ast.File, errs error) {
	var gimp *gnoImporter
	gimp = &gnoImporter{
		pkgPath: mpkg.Path,
		getter:  getter,
		cache:   map[string]gnoImporterResult{},
		cfg: &types.Config{
			Error: func(err error) {
				gimp.Error(err)
			},
		},
		errors: nil,
	}
	gimp.cfg.Importer = gimp

	pmode := ParseModeAll // type check all .gno files
	strict := true        // check gno.mod exists
	pkg, gofset, gofs, _gofs, tgofs, errs = gimp.typeCheckMemPackage(mpkg, pmode, strict)
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
	// when importing self (from xxx_test package) include *_test.gno.
	pkgPath string
	getter  MemPackageGetter
	cache   map[string]gnoImporterResult
	cfg     *types.Config
	errors  error // multierr
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
// pkgPath when imported by a package file located in dir.
func (gimp *gnoImporter) ImportFrom(pkgPath, _ string, _ types.ImportMode) (*types.Package, error) {
	if pkg, ok := gimp.cache[pkgPath]; ok {
		return pkg.pkg, pkg.err
	}
	mpkg := gimp.getter.GetMemPackage(pkgPath)
	if mpkg == nil {
		err := importNotFoundError(pkgPath)
		gimp.cache[pkgPath] = gnoImporterResult{err: err}
		return nil, err
	}
	var pmode = ParseModeProduction // don't parse test files for imports...
	if gimp.pkgPath == pkgPath {
		// ...unless importing self from a *_test.gno
		// file with package name xxx_test.
		pmode = ParseModeIntegration
	}
	strict := false // don't check for gno.mod for imports.
	pkg, _, _, _, _, errs := gimp.typeCheckMemPackage(mpkg, pmode, strict)
	if errs != nil {
		// NOTE:
		// Returning an error doesn't abort the type-checker.
		// Panic instead to quit quickly.
		panic(errs)
	}
	gimp.cache[pkgPath] = gnoImporterResult{pkg: pkg, err: errs}
	return pkg, errs
}

// Assumes that the code is Gno 0.9.
// If not, first use `gno lint` to transpile the code.
// Returns parsed *types.Package, *token.FileSet, []*ast.File.
//
// Args:
//   - pmode: ParseModeAll for type-checking all files.
//     ParseModeProduction when type-checking imports.
//   - strict: If true errors on gno.mod version mismatch.
func (gimp *gnoImporter) typeCheckMemPackage(mpkg *std.MemPackage, pmode ParseMode, strict bool) (
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
	gofset, gofs, _gofs, tgofs, errs = GoParseMemPackage(mpkg, pmode)
	if errs != nil {
		return nil, nil, nil, nil, nil, errs
	}
	if pmode == ParseModeProduction && (len(_gofs) > 0 || len(tgofs) > 0) {
		panic("unexpected test files from GoParseMemPackage()")
	}
	if pmode == ParseModeIntegration && (len(_gofs) > 0 || len(tgofs) > 0) {
		panic("unexpected xxx_test and *_filetest.gno tests")
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

	// STEP 4: Type-check Gno0.9 AST in Go (normal, and _test.gno if ParseModeIntegration).
	gofs = append(gofs, gmgof)
	pkg, _ = gimp.cfg.Check(mpkg.Path, gofset, gofs, nil)
	if gimp.errors != nil {
		errs = gimp.errors
		return
	}

	// STEP 4: Type-check Gno0.9 AST in Go (xxx_test package if ParseModeAll).
	if strings.HasSuffix(mpkg.Name, "_test") {
		// e.g. When running a filetest // PKGPATH: xxx_test.
	} else {
		gmgof.Name = ast.NewIdent(mpkg.Name + "_test")
		defer func() { gmgof.Name = ast.NewIdent(mpkg.Name) }() // revert
	}
	_gofs2 := append(_gofs, gmgof)
	_, _ = gimp.cfg.Check(mpkg.Path, gofset, _gofs2, nil)
	if gimp.errors != nil {
		errs = gimp.errors
		return
	}

	// STEP 4: Type-check Gno0.9 AST in Go (_filetest.gno if ParseModeAll).
	// Each filetest is its own package.
	defer func() { gmgof.Name = ast.NewIdent(mpkg.Name) }() // revert
	for _, tgof := range tgofs {
		gmgof.Name = tgof.Name // may be anything.
		tgof2 := []*ast.File{gmgof, tgof}
		_, _ = gimp.cfg.Check(mpkg.Path, gofset, tgof2, nil)
		if gimp.errors != nil {
			errs = gimp.errors
			return
		}
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
