package gnolang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
	"golang.org/x/tools/go/ast/astutil"
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
		mpkg = &std.MemPackage{Name: "gno0p9", Path: "gnobuiltins/gno0p9"}
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
type Gnocoins = gnocoins

type gnocoin struct {
    Denom string
    Amount int64
}
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
// type-checking.
type gnoBuiltinsGetterWrapper struct {
	getter MemPackageGetter
}

func (gw gnoBuiltinsGetterWrapper) GetMemPackage(pkgPath string) *std.MemPackage {
	if strings.HasPrefix(pkgPath, "gnobuiltins/") {
		return gnoBuiltinsMemPackage(pkgPath)
	} else {
		return gw.getter.GetMemPackage(pkgPath)
	}
}

// mode for both mpkg to type-check, as well as all imports.
type TypeCheckMode int

const (
	TCLatestStrict  TypeCheckMode = iota // require latest gno.mod gno version.
	TCLatestRelaxed                      // generate latest gno.mod if missing; for testing
	TCGno0p0                             // when gno fix'ing from gno 0.0.
)

// TypeCheckMemPackage performs type validation and checking on the given
// mpkg. To retrieve dependencies, it uses getter.
//
// The syntax checking is performed entirely using Go's go/types package.
//
// Args:

// - tcmode: TypeCheckMode, see comments above.
func TypeCheckMemPackage(mpkg *std.MemPackage, getter MemPackageGetter, pmode ParseMode, tcmode TypeCheckMode) (
	pkg *types.Package, errs error,
) {
	return TypeCheckMemPackageWithOptions(mpkg, getter, TypeCheckOptions{
		ParseMode: pmode,
		Mode:      tcmode,
	})
}

type TypeCheckCache map[string]*gnoImporterResult

// TypeCheckOptions allows to set custom options in [TypeCheckMemPackageWithOptions].
type TypeCheckOptions struct {
	Mode      TypeCheckMode
	ParseMode ParseMode
	// custom cache, for retaining results across several runs of the type
	// checker when the packages themselves won't change.
	Cache TypeCheckCache
}

// TypeCheckMemPackageWithOptions checks the given mpkg, configured using opts.
func TypeCheckMemPackageWithOptions(mpkg *std.MemPackage, getter MemPackageGetter, opts TypeCheckOptions) (
	pkg *types.Package, errs error,
) {
	if opts.Cache == nil {
		opts.Cache = TypeCheckCache{}
	}
	var gimp *gnoImporter
	gimp = &gnoImporter{
		pkgPath: mpkg.Path,
		tcmode:  opts.Mode,
		getter:  gnoBuiltinsGetterWrapper{getter},
		cache:   opts.Cache,
		cfg: &types.Config{
			Error: func(err error) {
				gimp.Error(err)
			},
		},
		errors: nil,
	}
	gimp.cfg.Importer = gimp
	pkg, errs = gimp.typeCheckMemPackage(mpkg, opts.ParseMode)
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
	pkgPath string
	tcmode  TypeCheckMode
	getter  MemPackageGetter
	cache   TypeCheckCache
	cfg     *types.Config
	errors  []error  // there may be many for a single import
	stack   []string // stack of pkgpaths for cyclic import detection
}

// Unused, but satisfies the Importer interface.
func (gimp *gnoImporter) Import(path string) (*types.Package, error) {
	return gimp.ImportFrom(path, "", 0)
}

// Pass through to cfg.Error for collecting all type-checking errors.
func (gimp *gnoImporter) Error(err error) {
	gimp.errors = append(gimp.errors, err)
}

// ImportFrom returns the imported package for the given import
// pkgPath when imported by a package file located in dir.
func (gimp *gnoImporter) ImportFrom(pkgPath, _ string, _ types.ImportMode) (*types.Package, error) {
	if result, ok := gimp.cache[pkgPath]; ok {
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
	gimp.cache[pkgPath] = result
	gimp.stack = append(gimp.stack, pkgPath)
	defer func() {
		gimp.stack = gimp.stack[:len(gimp.stack)-1]
	}()
	mpkg := gimp.getter.GetMemPackage(pkgPath)
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
	pmode := ParseModeProduction // don't parse test files for imports...
	if gimp.pkgPath == pkgPath {
		// ...unless importing self from a *_test.gno
		// file with package name xxx_test.
		pmode = ParseModeIntegration
	}
	pkg, errs := gimp.typeCheckMemPackage(mpkg, pmode)
	if errs != nil {
		result.err = errs
		result.pending = false
		return nil, errs
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
//
// Args:
//   - pmode: ParseModeAll for type-checking all files.
//     ParseModeProduction when type-checking imports.
func (gimp *gnoImporter) typeCheckMemPackage(mpkg *std.MemPackage, pmode ParseMode) (
	pkg *types.Package, errs error,
) {
	// See adr/pr4264_lint_transpile.md
	// STEP 2: Check gno.mod version.
	var gnoVersion string
	mod, err := ParseCheckGnoMod(mpkg)
	if err != nil {
		return nil, err
	}
	if gimp.tcmode == TCLatestStrict {
		if mod == nil {
			panic(fmt.Sprintf("gno.mod not found for package %q", mpkg.Path))
		}
		if mod.GetGno() != GnoVerLatest {
			panic(fmt.Sprintf("expected gno.mod gno version %v but got %v",
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
	gofset, allgofs, gofs, _gofs, tgofs, errs := GoParseMemPackage(mpkg, pmode)
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

	// STEP 4: Type-check Gno0.9 AST in Go (normal, and _test.gno if
	// ParseModeIntegration).
	if !strings.HasPrefix(mpkg.Path, "gnobuiltins/") {
		gofs = append(gofs, gmgof)
	}
	// NOTE: .Check doesn't return an err, it appends to .errors.  also,
	// gimp.errors may already be populated. For example, even after an
	// import failure the Go type checker will continue to try to import
	// more imports, to collect more errors for the user to see.
	numErrs := len(gimp.errors)
	pkg, _ = gimp.cfg.Check(mpkg.Path, gofset, gofs, nil)
	/* NOTE: Uncomment to fail earlier.
	if len(gimp.errors) != numErrs {
		errs = multierr.Combine(gimp.errors...)
		return
	}
	*/

	// STEP 4: Type-check Gno0.9 AST in Go (xxx_test package if ParseModeAll).
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
	_, _ = gimp.cfg.Check(mpkg.Path+"_test", gofset, _gofs2, nil)
	/* NOTE: Uncomment to fail earlier.
	if len(gimp.errors) != numErrs {
		errs = multierr.Combine(gimp.errors...)
		return
	}
	*/

	// STEP 4: Type-check Gno0.9 AST in Go (_filetest.gno if ParseModeAll).
	for _, tgof := range tgofs {
		// Each filetest is its own package.
		// XXX If we're re-parsing the filetest anyways,
		// change GoParseMemPackage to not parse into tgofs.
		tfname := filepath.Base(gofset.File(tgof.Pos()).Name())
		tpname := tgof.Name.String()
		tfile := mpkg.GetFile(tfname)
		// XXX If filetest are having issues, consider this:
		// pkgPath := fmt.Sprintf("%s_filetest%d", mpkg.Path, i)
		pkgPath := mpkg.Path
		tmpkg := &std.MemPackage{Name: tpname, Path: pkgPath}
		tmpkg.NewFile(tfname, tfile.Body)
		// NOTE: not gnobuiltins/*; gnobuiltins/* don't have filetests.
		bfile := makeGnoBuiltins(tpname, gnoVersion)
		tmpkg.AddFile(bfile)
		gofset2, _, gofs2, _, tgofs2, _ := GoParseMemPackage(tmpkg, ParseModeAll)
		if len(gimp.errors) != numErrs {
			/* NOTE: Uncomment to fail earlier.
			errs = multierr.Combine(gimp.errors...)
			return
			*/
			continue
		}
		// gofs2 (.gnobuiltins.gno), tgofs2 (*_testfile.gno)
		gofs2 = append(gofs2, tgofs2...)
		_, _ = gimp.cfg.Check(tmpkg.Path, gofset2, gofs2, nil)
		/* NOTE: Uncomment to fail earlier.
		if len(gimp.errors) != numErrs {
			errs = multierr.Combine(gimp.errors...)
			return
		}
		*/
	}
	return pkg, multierr.Combine(gimp.errors[numErrs:]...)
}

func deleteOldIdents(idents map[string]func(), gof *ast.File) {
	for _, decl := range gof.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		// ignore methods and init functions
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

//----------------------------------------
// GoParseMemPackage

type ParseMode int

const (
	// no test files.
	ParseModeProduction ParseMode = iota
	// production and test files when xxx_test tests import xxx package.
	ParseModeIntegration
	// all files even including *_filetest.gno; for linting and testing.
	ParseModeAll
	// a directory of file tests. consider all to be filetests.
	ParseModeOnlyFiletests
)

// ========================================
// Go parse the Gno source in mpkg to Go's *token.FileSet and
// []ast.File with `go/parser`.
//
// Args:
//   - pmode: see documentation for ParseMode.
//
// Results:
//   - gofs: all normal .gno files (and _test.gno files if wtests).
//   - _gofs: all xxx_test package _test.gno files if wtests.
//   - tgofs: all _testfile.gno test files.
func GoParseMemPackage(mpkg *std.MemPackage, pmode ParseMode) (
	gofset *token.FileSet, allgofs, gofs, _gofs, tgofs []*ast.File, errs error,
) {
	gofset = token.NewFileSet()

	// This map is used to allow for function re-definitions, which are
	// allowed in Gno (testing context) but not in Go.  This map links
	// each function identifier with a closure to remove its associated
	// declaration.
	delFunc := make(map[string]func())

	// Go parse and collect files from mpkg.
	for _, file := range mpkg.Files {
		// Ignore non-gno files.
		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}
		// Ignore _test/_filetest.gno files depending.
		switch pmode {
		case ParseModeProduction:
			if strings.HasSuffix(file.Name, "_test.gno") ||
				strings.HasSuffix(file.Name, "_filetest.gno") {
				continue
			}
		case ParseModeIntegration:
			if strings.HasSuffix(file.Name, "_filetest.gno") {
				continue
			}
		case ParseModeAll, ParseModeOnlyFiletests:
			// include all
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
			pmode == ParseModeOnlyFiletests {
			tgofs = append(tgofs, gof)
			allgofs = append(allgofs, gof)
		} else if strings.HasSuffix(file.Name, "_test.gno") &&
			strings.HasSuffix(gof.Name.String(), "_test") {
			if pmode == ParseModeIntegration {
				// never wanted these gofs.
				// (we do want other *_test.gno in gofs)
			} else {
				deleteOldIdents(delFunc, gof)
				_gofs = append(_gofs, gof)
				allgofs = append(allgofs, gof)
			}
		} else { // normal *_test.gno here for integration testing.
			deleteOldIdents(delFunc, gof)
			gofs = append(gofs, gof)
			allgofs = append(allgofs, gof)
		}
	}
	if errs != nil {
		return gofset, allgofs, gofs, _gofs, tgofs, errs
	}
	// END processing all files.
	// Sanity check before returning.
	if pmode == ParseModeProduction && (len(_gofs) > 0 || len(tgofs) > 0) {
		panic("unexpected test files from GoParseMemPackage()")
	}
	if pmode == ParseModeIntegration && (len(_gofs) > 0 || len(tgofs) > 0) {
		panic("unexpected xxx_test and *_filetest.gno tests")
	}
	return
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
func (e ImportCycleError) assertImportError()    {}

var (
	_ ImportError = ImportNotFoundError{}
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
