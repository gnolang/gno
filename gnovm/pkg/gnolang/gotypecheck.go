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

	"go.uber.org/multierr"
	"golang.org/x/tools/go/ast/astutil"

	"github.com/gnolang/gno/tm2/pkg/std"
)

/*
	Type-checking (using go/types).
	Refer to the [Lint and Transpile ADR](./adr/pr4264_lint_transpile.md).
	XXX move to pkg/gnolang/importer.go.
*/

// While makeGnoBuiltins() returns a *std.MemFile to inject into each package,
// they may need to import a central package if they declare any types,
// otherwise each .gnobuiltins.gno would be declaring their own types.
var gnoBuiltinsCache = make(map[string]*std.MemPackage) // pkgPath -> mpkg or nil.

func gnoBuiltinsMemPackage(pkgPath string) *std.MemPackage {
	if !strings.HasPrefix(pkgPath, "gnobuiltins/") {
		panic("expected pkgPath to start with gnobuiltins/")
	}
	mpkg, ok := gnoBuiltinsCache[pkgPath]
	if ok {
		return mpkg
	}
	switch pkgPath {
	case "gnobuiltins/gno0p9": // 0.9
		mpkg = &std.MemPackage{Type: MPStdlibProd, Name: "gno0p9", Path: "gnobuiltins/gno0p9"}
		mpkg.SetFile("gno0p9.gno", `package gno0p9
type realm interface {
    Address() address
    PkgPath() string
    Coins() gnocoins
    Send(coins gnocoins, to address) error
    Previous() realm
    Origin() realm
    String() string
}
type Realm = realm

type address string
func (a address) String() string { return string(a) }
func (a address) IsValid() bool { return false } // shim
type Address = address

type gnocoins []gnocoin
func (cz gnocoins) String() string { return "" } // shim

type Gnocoins = gnocoins

type gnocoin struct {
    Denom string
    Amount int64
}
func (c gnocoin) String() string { return "" } // shim

type Gnocoin = gnocoin
`)
	default:
		panic("unrecognized gnobuiltins pkgpath")
	}
	gnoBuiltinsCache[pkgPath] = mpkg
	return mpkg
}

func makeGnoBuiltins(pkgName string, gnoVersion string) *std.MemFile {
	gnoBuiltins := ""
	switch gnoVersion {
	case GnoVerLatest: // 0.9
		gnoBuiltins = `package %s
import "gnobuiltins/gno0p9"

func istypednil(x any) bool { return false } // shim
var cross realm // shim
func revive[F any](fn F) any { return nil } // shim
type realm = gno0p9.Realm
type address = gno0p9.Address
type gnocoins = gno0p9.Gnocoins
type gnocoin = gno0p9.Gnocoin
`
	case GnoVerMissing: // 0.0
		gnoBuiltins = `package %s

func istypednil(x any) bool { return false } // shim
func crossing() { } // shim
func cross[F any](fn F) F { return fn } // shim XXX: THIS MUST NOT EXIST IN .gnobuiltins.gno for 0.9!!!
func _cross_gno0p0[F any](fn F) F { return fn } // shim XXX: THIS MUST NOT EXIST IN .gnobuiltins.gno for 0.9!!!
func revive[F any](fn F) any { return nil } // shim
`
	default:
		panic("unsupported gno.mod version " + gnoVersion)
	}
	file := &std.MemFile{
		Name: ".gnobuiltins.gno", // because GoParseMemPackage expects .gno.
		Body: fmt.Sprintf(gnoBuiltins, pkgName),
	}
	return file
}

// MemPackageGetter implements the GetMemPackage() method. It is a subset of
// [Store], separated for ease of testing.
type MemPackageGetter interface {
	GetMemPackage(pkgPath string) *std.MemPackage
}

// Wrap the getter and intercept "gnobuiltins/*" paths, which is only used for
// type-checking. Also intercept mpkg namely for integration
// xxx_test *_test.gno files to import the target package.
type gimpGetterWrapper struct {
	mpkg   *std.MemPackage
	getter MemPackageGetter
}

func (gw gimpGetterWrapper) GetMemPackage(pkgPath string) *std.MemPackage {
	if strings.HasPrefix(pkgPath, "gnobuiltins/") {
		return gnoBuiltinsMemPackage(pkgPath)
	} else if gw.mpkg != nil && gw.mpkg.Type != MPFiletests && gw.mpkg.Path == pkgPath {
		return gw.mpkg
	} else {
		return gw.getter.GetMemPackage(pkgPath)
	}
}

// mode for both mpkg to type-check, as well as all imports.
type TypeCheckMode int

const (
	TCLatestStrict  TypeCheckMode = iota // require latest gnomod.toml gno version, forbid drafts
	TCGenesisStrict                      // require latest gnomod.toml gno version, allow drafts
	TCLatestRelaxed                      // generate latest gnomod.toml if missing
	TCGno0p0                             // when gno fix'ing from gno 0.0.
)

// RequiresLatestGnoMod returns true if the type check mode requires latest gno.mod version
func (m TypeCheckMode) RequiresLatestGnoMod() bool {
	return m == TCLatestStrict || m == TCGenesisStrict
}

// TypeCheckCache is a permanent cache for packages imported using MPFProd,
// excluding tests.
type TypeCheckCache map[string]*types.Package

type TypeCheckOptions struct {
	// Getter is the normal package import getter, without test stdlibs.
	Getter MemPackageGetter
	// TestGetter is the package import getter for test stdlibs with overrides
	// when importing from *_test.gno|*_filetest.gno.
	TestGetter MemPackageGetter
	// Mode is the [TypeCheckMode]. Refer to the type documentation.
	Mode TypeCheckMode

	// Cache is an optional permanent cache of already imported standard
	// libraries. Packages found in the Cache won't need to be type checked
	// again.
	Cache TypeCheckCache

	// Fset, if non-nil, is used for Go parsing instead of creating a new one.
	// After TypeCheckMemPackage returns, it contains the file position
	// information from the parsed package.
	Fset *token.FileSet
}

// TypeCheckMemPackage performs type validation and checking on the given
// mpkg. To retrieve dependencies, it uses getter.
//
// The syntax checking is performed entirely using Go's go/types package.
//
// Args:
//   - tcmode: TypeCheckMode, see comments above.
//   - getter: the normal package import getter without test stdlibs.
//   - tgetter: getter for test stdlibs with overrides when gimp.testing (importing from *_test.gno|*_filetest.gno).
func TypeCheckMemPackage(mpkg *std.MemPackage, opts TypeCheckOptions) (
	pkg *types.Package, errs error,
) {
	var gimp *gnoImporter
	gimp = &gnoImporter{
		pkgPath:   mpkg.Path,
		tcmode:    opts.Mode,
		testing:   false, // only true for imports from testing files.
		getter:    gimpGetterWrapper{nil, opts.Getter},
		tgetter:   gimpGetterWrapper{mpkg, opts.TestGetter},
		cache:     map[string]*gnoImporterResult{},
		permCache: opts.Cache,
		fset:      opts.Fset,
		cfg: &types.Config{
			Error: func(err error) {
				gimp.Error(err)
			},
		},
		errors: nil,
	}
	gimp.cfg.Importer = gimp

	pkg, errs = gimp.typeCheckMemPackage(mpkg, nil)
	return
}

type gnoImporterResult struct {
	pkg     *types.Package
	err     error
	pending bool // for cyclic import detection
}

// gimp.
// gimp type checks.
// gimp remembers.
// gimp.
type gnoImporter struct {
	// when importing self (from xxx_test package) include *_test.gno.
	pkgPath   string
	tcmode    TypeCheckMode
	testing   bool             // if true, use tgetter for stdlibs.
	getter    MemPackageGetter // used for stdlibs if !.testing, and everything else.
	tgetter   MemPackageGetter // used for stdlibs if .testing
	cache     map[string]*gnoImporterResult
	permCache TypeCheckCache
	fset      *token.FileSet // if non-nil, used for Go parsing instead of creating a new one.
	cfg       *types.Config
	errors    []error  // there may be many for a single import
	stack     []string // stack of pkgpaths for cyclic import detection
}

// Unused, but satisfies the Importer interface.
func (gimp *gnoImporter) Import(path string) (*types.Package, error) {
	return gimp.ImportFrom(path, "", 0)
}

// Pass through to cfg.Error for collecting all type-checking errors.
func (gimp *gnoImporter) Error(err error) {
	gimp.errors = append(gimp.errors, err)
}

func cacheKey(pkgPath string, testing bool) string {
	if testing {
		return pkgPath + ":testing"
	} else {
		return pkgPath
	}
}

// ImportFrom returns the imported package for the given import
// pkgPath when imported by a package file located in dir.
func (gimp *gnoImporter) ImportFrom(pkgPath, _ string, _ types.ImportMode) (gopkg *types.Package, err error) {
	ck := cacheKey(pkgPath, gimp.testing)
	if result, ok := gimp.cache[ck]; ok {
		if result.pending {
			idx := slices.Index(gimp.stack, pkgPath)
			cycle := gimp.stack[idx:]
			err := ImportCycleError{Cycle: cycle}
			// NOTE: see comment below for ImportNotFoundError.
			// gimp.importErrors = append(gimp.importErrors, err)
			result.err = err
			return nil, err
		} else {
			return result.pkg, result.err
		}
	}
	result := &gnoImporterResult{pending: true}
	gimp.cache[ck] = result
	gimp.stack = append(gimp.stack, pkgPath)
	defer func() {
		gimp.stack = gimp.stack[:len(gimp.stack)-1]
	}()
	// In a vast majority of cases, we can use the permCache if it is set.
	canPerm := gimp.permCache != nil &&
		((!gimp.testing && pkgPath != gimp.pkgPath) || (IsStdlib(pkgPath) && !IsStdlib(gimp.pkgPath)))
	if canPerm {
		pkg := gimp.permCache[ck]
		if pkg != nil {
			result.pkg = pkg
			result.err = nil
			result.pending = false
			return pkg, nil
		}
	}
	var mpkg *std.MemPackage
	if gimp.testing && (IsStdlib(pkgPath) || pkgPath == gimp.pkgPath) {
		mpkg = gimp.tgetter.GetMemPackage(pkgPath)
	} else {
		mpkg = gimp.getter.GetMemPackage(pkgPath)
	}
	if gimp.pkgPath == pkgPath {
		if gimp.testing {
			// xxx_test importing xxx for integration testing
			mpkg = MPFTest.FilterMemPackage(mpkg)
		} else {
			// This happens when type-checking from
			// pkg/test.runFiletest().  Normally when
			// gno.TypeCheckMemPackage() is called for a normal
			// user package which happens to include *_filetest.go
			// file tests gimp.testing will be set to true, but
			// when running the filetests each filetest is run as
			// its own mempackage (with no other files).
			// Furthermore, gnovm internal test/files filetests are
			// type-checked individually (since these testfiles are
			// not part of any package) each as a prod file, so
			// gimp.testing is false, but with gimp.getter and
			// gimp.tgetter set to the same teststore except only
			// tgetter is injected with mpkg being tested.  In
			// order to allow testfiles to have stdlib package
			// paths without overwriting existing stdlibs, in this
			// case we simply fetch it from gimp.getter above.
			mpkg = MPFProd.FilterMemPackage(mpkg)
		}
	} else {
		mpkg = MPFProd.FilterMemPackage(mpkg)
	}
	if mpkg == nil {
		err := ImportNotFoundError{PkgPath: pkgPath}
		// NOTE: When returning an err, Go will strip type information.
		// When panic'd, the type information will be preserved, but
		// the file location information will be lost.  Therefore,
		// return the error but later in printError() parse the message
		// and recast to a gnoImportError.
		// TODO: For completeness we could append to a separate slice
		// and check presence in gimp.importErrors before converting.
		// gimp.importErrors = append(gimp.importErrors, err)
		result.err = err
		result.pending = false
		return nil, err
	}
	// ensure import is not a draft package.
	mod, err := ParseCheckGnoMod(mpkg)
	if err != nil {
		result.err = err
	}
	if gimp.tcmode == TCLatestStrict && mod != nil && mod.Draft {
		// cannot import draft packages after genesis.
		// NOTE: see comment below for ImportNotFoundError.
		err = ImportDraftError{PkgPath: pkgPath}
		result.err = err
		result.pending = false
		return nil, err
	}
	if mod != nil && mod.Private {
		// If the package is private, we cannot import it.
		err := ImportPrivateError{PkgPath: pkgPath}
		// NOTE: see comment above for ImportNotFoundError.
		result.err = err
		result.pending = false
		return nil, err
	}
	wtests := gimp.testing && gimp.pkgPath == pkgPath
	pkg, errs := gimp.typeCheckMemPackage(mpkg, &wtests)
	if errs != nil {
		result.err = errs
		result.pending = false
		return nil, errs
	}
	if canPerm {
		gimp.permCache[ck] = pkg
	}
	result.pkg = pkg
	result.err = nil
	result.pending = false
	return pkg, errs
}

// Minimal AST mutation(s) for Go.
// For gno 0.0 there was nothing to do besides including .gnobuiltins.gno.
// For gno 0.9 we need to support init(cur realm), main(cur realm) by
// removing them and instead setting `cur := cross`; hacky but good enough.
func prepareGoGno0p9(f *ast.File) (err error) {
	astutil.Apply(f, nil, func(c *astutil.Cursor) bool { // leaving...
		switch gon := c.Node().(type) {
		case *ast.FuncDecl:
			name := gon.Name.String()
			if gon.Recv == nil && (name == "main" || name == "init") {
				if len(gon.Type.Params.List) == 1 { // `cur realm`
					gon.Type.Params.List = nil
				} else {
					return true
				}
				// This assignment is not valid in gno.
				// `as1` declares cur and `as2` "uses" it.
				as1 := &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("cur")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{ast.NewIdent("cross")},
				}
				as2 := &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("cross")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent("cur")},
				}
				// Not sure if this does anything,
				// but we want the line numbers to not change.
				insert := gon.Type.End()
				as1.Lhs[0].(*ast.Ident).NamePos = insert
				as1.TokPos = insert
				as1.Rhs[0].(*ast.Ident).NamePos = insert
				as2.Lhs[0].(*ast.Ident).NamePos = insert
				as2.TokPos = insert
				as2.Rhs[0].(*ast.Ident).NamePos = insert
				// Prepend define and use of `cur`.
				gon.Body.List = append([]ast.Stmt{as1, as2},
					gon.Body.List...)
			}
		}
		return true
	})
	return err
}

// Assumes that the code is Gno 0.9.
// If not, first use `gno lint` to transpile the code.
// Returns parsed *types.Package, *token.FileSet, []*ast.File.
//   - wtests: if nil, type check all, including filetests; otherwise returns early.
func (gimp *gnoImporter) typeCheckMemPackage(mpkg *std.MemPackage, wtests *bool) (
	pkg *types.Package, errs error,
) {
	// See adr/pr4264_lint_transpile.md
	// STEP 2: Check gno.mod version.
	var gnoVersion string
	mod, err := ParseCheckGnoMod(mpkg)
	if err != nil {
		return nil, err
	}
	if gimp.tcmode.RequiresLatestGnoMod() {
		if mod == nil {
			panic(fmt.Sprintf("gnomod.toml not found for package %q", mpkg.Path))
		}
		if mod.GetGno() != GnoVerLatest {
			panic(fmt.Sprintf("expected gnomod.toml gno version %v but got %v",
				GnoVerLatest, mod.GetGno()))
		}
		gnoVersion = mod.GetGno()
	} else {
		if mod == nil {
			// cannot be stdlib; ParseCheckGnoMod will generate a
			// gno.mod with version latest. Sanity check:
			if IsStdlib(mpkg.Path) {
				panic("expected ParseCheckGnoMod() to auto-generate a gno.mod for stdlibs")
			}
			switch gimp.tcmode {
			case TCGno0p0:
				gnoVersion = GnoVerMissing
			case TCLatestRelaxed:
				gnoVersion = GnoVerLatest
			}
		} else {
			gnoVersion = mod.GetGno()
		}
	}

	// STEP 3: Parse the mem package to Go AST.
	gofset, allgofs, gofs, _gofs, tgofs, errs := GoParseMemPackage(mpkg, gimp.fset)
	if errs != nil {
		return nil, errs
	}

	// STEP 3: Prepare for Go type-checking.
	for _, gof := range allgofs {
		err := prepareGoGno0p9(gof)
		if err != nil {
			panic(fmt.Sprintf("unexpected error: %v", err))
		}
	}

	// STEP 3: Add and Parse .gnobuiltins.go file.
	file := makeGnoBuiltins(mpkg.Name, gnoVersion)
	const parseOpts = parser.ParseComments |
		parser.DeclarationErrors |
		parser.SkipObjectResolution
	gmgof, err := parser.ParseFile(
		gofset,
		path.Join(mpkg.Path, file.Name),
		file.Body,
		parseOpts)
	if err != nil {
		panic(fmt.Errorf("error parsing gotypecheck .gnobuiltins.gno file: %w", err))
	}

	// STEP 4: Type-check Gno0.9 AST in Go (normal/production only).
	if !strings.HasPrefix(mpkg.Path, "gnobuiltins/") {
		gofs = append(gofs, gmgof)
	}
	// NOTE: .Check doesn't return an err, it appends to .errors.  also,
	// gimp.errors may already be populated. For example, even after an
	// import failure the Go type checker will continue to try to import
	// more imports, to collect more errors for the user to see.
	numErrs := len(gimp.errors)
	origTesting := gimp.testing
	defer func() { gimp.testing = origTesting }() // reset after.
	// Preserve gimp.testing, sub-imports are under the same context.
	// gimp.testing = false <-- incorrect!
	pgofs := filterTests(gofset, gofs) // prod gofs.
	pkg, _ = gimp.cfg.Check(mpkg.Path, gofset, pgofs, nil)
	// Fail early: there's no point checking the others.
	if len(gimp.errors) != numErrs {
		errs = multierr.Combine(gimp.errors[numErrs:]...)
		return
	}
	if wtests != nil && !*wtests {
		errs = multierr.Combine(gimp.errors[numErrs:]...)
		if errs != nil {
			pkg = nil
		}
		return
	}

	// STEP 4: Type-check Gno0.9 AST in Go (w/ tests, but not xxx_tests).
	if len(pgofs) < len(gofs) {
		gimp.testing = true // use tgetter for stdlibs, default to getter.
		pkg, _ = gimp.cfg.Check(mpkg.Path, gofset, gofs, nil)
		// Fail early: there's no point checking the others.
		if len(gimp.errors) != numErrs {
			errs = multierr.Combine(gimp.errors[numErrs:]...)
			return
		}
	}
	if wtests != nil { // *wtests is true.
		errs = multierr.Combine(gimp.errors[numErrs:]...)
		if errs != nil {
			pkg = nil
		}
		return
	}

	// STEP 4: Type-check Gno0.9 AST in Go (xxx_test package).
	if strings.HasSuffix(mpkg.Name, "_test") {
		// e.g. When running a filetest // PKGPATH: xxx_test.
	} else if !strings.HasPrefix(mpkg.Path, "gnobuiltins/") {
		gmgof.Name = ast.NewIdent(mpkg.Name + "_test")
		defer func() { gmgof.Name = ast.NewIdent(mpkg.Name) }() // revert
	}
	_gofs2 := _gofs
	if !strings.HasPrefix(mpkg.Path, "gnobuiltins/") {
		_gofs2 = append(_gofs, gmgof)
	}
	gimp.testing = true // use tgetter for stdlibs, default to getter.
	_, _ = gimp.cfg.Check(mpkg.Path+"_test", gofset, _gofs2, nil)
	/* NOTE: Uncomment to fail earlier.
	if len(gimp.errors) != numErrs {
		errs = multierr.Combine(gimp.errors[numErrs:]...)
		return
	}
	*/

	// STEP 4: Type-check Gno0.9 AST in Go (_filetest.gno).
	for _, tgof := range tgofs {
		// Each filetest is its own package.
		tpname := tgof.Name.String()
		gmgof.Name = ast.NewIdent(tpname)
		tgofs2 := []*ast.File{gmgof, tgof}
		gimp.testing = true // use tgetter for stdlibs, default to tgetter.
		_, _ = gimp.cfg.Check(mpkg.Path, gofset, tgofs2, nil)
		/* NOTE: Uncomment to fail earlier.
		if len(gimp.errors) != numErrs {
			errs = multierr.Combine(gimp.errors[numErrs:]...)
			return
		}
		*/
	}
	return pkg, multierr.Combine(gimp.errors[numErrs:]...)
}

// Ensure uniqueness of declarations,
// e.g. test/stdlibs overriding stdlibs.
func uniqueDecls(decls map[string]struct{}, gof *ast.File) {
	dupes := []ast.Decl{}
	for _, decl := range gof.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		// ignore methods and init functions
		if !ok ||
			fd.Recv != nil ||
			fd.Name.Name == "init" {
			continue
		}
		// if declaration is duplicate, delete this one.
		_, exists := decls[fd.Name.Name]
		if exists {
			// delete this one. doesn't matter which one (whether
			// Go native or gno) for type-checking.
			dupes = append(dupes, decl)
		} else {
			decls[fd.Name.Name] = struct{}{}
		}
	}
	// actually delete.
	gof.Decls = slices.DeleteFunc(gof.Decls,
		func(d ast.Decl) bool { return slices.Contains(dupes, d) })
}

// ========================================
// Go parse the Gno source in mpkg to Go's *token.FileSet and
// []ast.File with `go/parser`.
//
// Results:
//   - gofs: all normal .gno files (and _test.gno files if wtests).
//   - _gofs: all xxx_test package _test.gno files if wtests.
//   - tgofs: all _testfile.gno test files.
func GoParseMemPackage(mpkg *std.MemPackage, fset *token.FileSet) (
	gofset *token.FileSet, allgofs, gofs, _gofs, tgofs []*ast.File, errs error,
) {
	if fset != nil {
		gofset = fset
	} else {
		gofset = token.NewFileSet()
	}

	// This map is used to allow for Go native overrides/redeclarations.
	decls := make(map[string]struct{}) // (func) decl name

	// Go parse and collect files from mpkg.
	for _, file := range mpkg.Files {
		// Ignore non-gno files.
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}
		// Ignore _test/_filetest.gno files depending.
		switch mpkg.Type {
		case MPAnyAll:
			panic("undefined MPAnyAll")
		case MPUserAll, MPStdlibAll, MPFiletests:
			// parse all.
		case MPUserProd, MPStdlibProd:
			// ignore test files.
			if strings.HasSuffix(file.Name, "_test.gno") ||
				strings.HasSuffix(file.Name, "_filetest.gno") {
				continue
			}
		case MPUserTest, MPStdlibTest:
			// TODO: rename to MPIntegration?
			if strings.HasSuffix(file.Name, "_filetest.gno") {
				continue
			}
		case MPUserIntegration, MPStdlibIntegration:
			// parse all integration test files
		default:
			panic("should not happen")
		}

		// Go parse file.
		const parseOpts = parser.ParseComments |
			parser.DeclarationErrors |
			parser.SkipObjectResolution
		gof, err := parser.ParseFile(
			gofset, path.Join(mpkg.Path, file.Name),
			file.Body,
			parseOpts)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		// The *ast.File passed all filters.
		if strings.HasSuffix(file.Name, "_filetest.gno") ||
			mpkg.Type == MPFiletests {
			tgofs = append(tgofs, gof)
			allgofs = append(allgofs, gof)
		} else if strings.HasSuffix(file.Name, "_test.gno") &&
			strings.HasSuffix(gof.Name.String(), "_test") {
			switch mpkg.Type {
			case MPAnyAll:
				panic("undefined MPAnyAll")
			case MPUserProd, MPStdlibProd:
				panic("should not happen")
			case MPUserTest, MPStdlibTest:
				// Do not include xxx_test.
				// xxx_test imports normal and only
				// non-xxx_test *test.go files from xxx.
			case MPUserAll, MPStdlibAll, MPFiletests:
				_gofs = append(_gofs, gof)
				allgofs = append(allgofs, gof)
			default:
				panic("should not happen")
			}
		} else if strings.HasSuffix(file.Name, "_test.gno") {
			// !strings.HasSuffix(gof.Name.String(), "_test")
			// + non-xxx_test *_test.gno here for integration testing.
			gofs = append(gofs, gof)
			allgofs = append(allgofs, gof)
		} else { // prod files
			uniqueDecls(decls, gof)
			gofs = append(gofs, gof)
			allgofs = append(allgofs, gof)
		}
	}
	if errs != nil {
		return gofset, allgofs, gofs, _gofs, tgofs, errs
	}
	// END processing all files.
	// Sanity check before returning.
	if mpkg.Type.(MemPackageType).IsProd() && (len(_gofs) > 0 || len(tgofs) > 0) {
		panic("unexpected test files from GoParseMemPackage()")
	}
	if mpkg.Type.(MemPackageType).IsTest() && (len(_gofs) > 0 || len(tgofs) > 0) {
		// same as above, because the non-xxx_test *_test.gno files are
		// part of gofs, not _gofs; for testing purposes those test
		// files extend the original package when imported by xxx_test
		// *_test.gno files.
		panic("unexpected xxx_test and *_filetest.gno tests")
	}
	return
}

func filterTests(gofset *token.FileSet, gofs []*ast.File) []*ast.File {
	pgofs := make([]*ast.File, 0, len(gofs))
	for _, gof := range gofs {
		gofname := gofset.File(gof.Pos()).Name()
		if strings.HasSuffix(gofname, "_test.gno") {
			continue
		} else {
			pgofs = append(pgofs, gof)
		}
	}
	return pgofs
}

//----------------------------------------
// Errors

// ImportError is an interface type.
type ImportError interface {
	assertImportError()
	error
	GetLocation() string
	GetMsg() string
}

func (e ImportNotFoundError) assertImportError() {}
func (e ImportPrivateError) assertImportError()  {}
func (e ImportDraftError) assertImportError()    {}
func (e ImportCycleError) assertImportError()    {}

var (
	_ ImportError = ImportNotFoundError{}
	_ ImportError = ImportPrivateError{}
	_ ImportError = ImportDraftError{}
	_ ImportError = ImportCycleError{}
)

// ImportNotFoundError implements ImportError
type ImportNotFoundError struct {
	Location string
	PkgPath  string
}

func (e ImportNotFoundError) GetLocation() string { return e.Location }

func (e ImportNotFoundError) GetMsg() string { return fmt.Sprintf("unknown import path %q", e.PkgPath) }

func (e ImportNotFoundError) Error() string { return importErrorString(e) }

// ImportDraftError implements ImportError
type ImportDraftError struct {
	Location string
	PkgPath  string
}

func (e ImportDraftError) GetLocation() string { return e.Location }

func (e ImportDraftError) GetMsg() string {
	return fmt.Sprintf("import path %q is a draft package and can only be imported at genesis", e.PkgPath)
}

func (e ImportDraftError) Error() string { return importErrorString(e) }

// ImportPrivateError implements ImportError
type ImportPrivateError struct {
	Location string
	PkgPath  string
}

func (e ImportPrivateError) GetLocation() string { return e.Location }

func (e ImportPrivateError) GetMsg() string {
	return fmt.Sprintf("import path %q is private and cannot be imported", e.PkgPath)
}

func (e ImportPrivateError) Error() string { return importErrorString(e) }

// ImportCycleError implements ImportError
type ImportCycleError struct {
	Location string
	Cycle    []string
}

func (e ImportCycleError) GetLocation() string { return e.Location }

func (e ImportCycleError) GetMsg() string {
	return fmt.Sprintf("cyclic import detected: %s -> %s", strings.Join(e.Cycle, " -> "), e.Cycle[0])
}

func (e ImportCycleError) Error() string { return importErrorString(e) }

// helper
func importErrorString(err ImportError) string {
	loc := err.GetLocation()
	msg := err.GetMsg()
	if loc != "" {
		return loc + ": " + msg
	} else {
		return msg
	}
}
