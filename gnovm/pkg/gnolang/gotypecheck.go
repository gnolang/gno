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
)

/*
	Type-checking (using go/types).
	Refer to the [Lint and Transpile ADR](./adr/pr4264_lint_transpile.md).
	XXX move to pkg/gnolang/importer.go.
*/

func makeGnoBuiltins(pkgName string) *std.MemFile {
	file := &std.MemFile{
		Name: ".gnobuiltins.gno", // because GoParseMemPackage expects .gno.
		Body: fmt.Sprintf(`package %s

func istypednil(x any) bool { return false } // shim
func crossing() { } // shim
func cross[F any](fn F) F { return fn } // shim
func revive[F any](fn F) any { return nil } // shim
type realm interface{} // shim
`, pkgName),
	}
	return file
}

// MemPackageGetter implements the GetMemPackage() method. It is a subset of
// [Store], separated for ease of testing.
type MemPackageGetter interface {
	GetMemPackage(path string) *std.MemPackage
}

type TypeCheckFilesResult struct {
	FileSet          *token.FileSet
	SourceFiles      []*ast.File // All normal .gno files (and _test.gno files if wtests).
	TestPackageFiles []*ast.File // All files in test packages (_test.gno & _testfile.gno).
	TestFiles        []*ast.File // All standalone test files (_testfile.gno).
}

// TypeCheckMemPackage performs type validation and checking on the given
// mpkg. To retrieve dependencies, it uses getter.
//
// The syntax checking is performed entirely using Go's go/types package.
func TypeCheckMemPackage(mpkg *std.MemPackage, getter MemPackageGetter, pmode ParseMode) (
	pkg *types.Package, tfiles *TypeCheckFilesResult, errs error,
) {
	var gimp *gnoImporter
	gimp = &gnoImporter{
		pkgPath: mpkg.Path,
		getter:  getter,
		cache:   map[string]*gnoImporterResult{},
		cfg: &types.Config{
			Error: func(err error) {
				gimp.Error(err)
			},
		},
		errors: nil,
	}
	gimp.cfg.Importer = gimp

	strict := true // check gno.mod exists
	return gimp.typeCheckMemPackage(mpkg, pmode, strict)
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
	getter  MemPackageGetter
	cache   map[string]*gnoImporterResult
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
	strict := false // don't check for gno.mod for imports.
	pkg, _, errs := gimp.typeCheckMemPackage(mpkg, pmode, strict)
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

// Assumes that the code is Gno 0.9.
// If not, first use `gno lint` to transpile the code.
// Returns parsed *types.Package, *token.FileSet, []*ast.File.
//
// Args:
//   - pmode: ParseModeAll for type-checking all files.
//     ParseModeProduction when type-checking imports.
//   - strict: If true errors on gno.mod version mismatch.
func (gimp *gnoImporter) typeCheckMemPackage(mpkg *std.MemPackage, pmode ParseMode, strict bool) (
	pkg *types.Package, tfiles *TypeCheckFilesResult, errs error,
) {
	// See adr/pr4264_lint_transpile.md
	// STEP 2: Check gno.mod version.
	if strict {
		_, err := ParseCheckGnoMod(mpkg)
		if err != nil {
			return nil, nil, err
		}
	}

	// STEP 3: Parse the mem package to Go AST.
	gofset, gofs, _gofs, tgofs, errs := GoParseMemPackage(mpkg, pmode)
	if errs != nil {
		return nil, nil, errs
	}
	if pmode == ParseModeProduction && (len(_gofs) > 0 || len(tgofs) > 0) {
		panic("unexpected test files from GoParseMemPackage()")
	}
	if pmode == ParseModeIntegration && (len(_gofs) > 0 || len(tgofs) > 0) {
		panic("unexpected xxx_test and *_filetest.gno tests")
	}

	// STEP 3: Add and Parse .gnobuiltins.go file.
	file := makeGnoBuiltins(mpkg.Name)
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

	// NOTE: When returning errs from this function,

	// STEP 4: Type-check Gno0.9 AST in Go (normal, and _test.gno if ParseModeIntegration).
	gofs = append(gofs, gmgof)
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
	} else {
		gmgof.Name = ast.NewIdent(mpkg.Name + "_test")
		defer func() { gmgof.Name = ast.NewIdent(mpkg.Name) }() // revert
	}
	_gofs2 := append(_gofs, gmgof)
	_, _ = gimp.cfg.Check(mpkg.Path+"_test", gofset, _gofs2, nil)
	/* NOTE: Uncomment to fail earlier.
	if len(gimp.errors) != numErrs {
		errs = multierr.Combine(gimp.errors...)
		return
	}
	*/

	// STEP 4: Type-check Gno0.9 AST in Go (_filetest.gno if ParseModeAll).
	defer func() { gmgof.Name = ast.NewIdent(mpkg.Name) }() // revert
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
		bfile := makeGnoBuiltins(tpname)
		tmpkg.AddFile(bfile)
		gofset2, gofs2, _, tgofs2, _ := GoParseMemPackage(tmpkg, ParseModeAll)
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

	tfiles = &TypeCheckFilesResult{
		FileSet:          gofset,
		SourceFiles:      gofs,
		TestPackageFiles: _gofs,
		TestFiles:        gofs,
	}
	return pkg, tfiles, multierr.Combine(gimp.errors[numErrs:]...)
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
