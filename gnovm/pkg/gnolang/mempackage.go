package gnolang

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	gofmt "go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	r "github.com/gnolang/gno/tm2/pkg/regx"
	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
)

//----------------------------------------
// Mempackage package path functions.

var (
	// NOTE: These are further restrictions upon the validation that
	// already happens by std.MemPackage.Validate().  See also
	// tm2/pkg/std/memfile.go which has more relaxed rules.
	// XXX test exhaustively balanced futureproof vs restrictive.
	//
	// Valid gnoUserPkgPaths: (user as in user-land; system pkgpaths included)
	//  - sub.domain.tld/a/any
	//  - sub.domain.tld/b/single
	//  - sub.domain.tld/c/letter
	//  - sub.domain.tld/d/works
	//  - sub.domain.tld/r/realm
	//  - sub.domain.tld/r/_realm/_path
	//  - sub.domain.tld/p/package/_path123/etc
	//
	// Further validation should be done with LETTER to determine the type of pkgPath:
	//  - /r/ for realm paths
	//  - /p/ for p package paths
	//  - /e/ for run paths, where USER is Re_address and REPO is "/run".
	Re_gnoUserPkgPath = r.N("PKGPATH",
		Re_domain,
		r.N("URLPATH", // pkgpath minus domain
			r.E(`/`), r.N("LETTER", r.C(`a-z`)), r.E(`/`), // single latter e.g. /p/ or /r/.
			r.N("USER", Re_name), // user or org name.
			r.M(r.E(`/`), r.N("REPO", Re_name, r.S(r.E(`/`), Re_name))))) // rest of path.

	// Valid gnoStdPkgPaths:
	//  - math
	//  - math/fourier123
	//  - justnodots
	//  - _nodots123
	//  - _nodots123/_subpath1/_subpath2
	Re_gnoStdPkgPath = r.N("PKGPATH",
		Re_name, r.S(r.E(`/`), Re_name)) // no dots, just name(s) with `/` delimiter.

	// Standard components of all Gno pkgpaths:
	// (All paths must be lowercase ascii alphanumeric characters)
	Re_domain = r.N("DOMAIN", // all lowercase
		r.N("SLD", r.P(r.P(r.C(`a-z0-9-`)), r.E(`.`))), // sub(level)domain, permissive w/ dashes.
		r.N("TLD", r.R(2, 63, r.C(`a-z`))))             // top level domain, 2~63 letters.
	Re_name    = r.G(r.M(`_`), r.C(`a-z`), r.S(r.C(`a-z0-9_`))) // optional leading _, start with letter, no dots!
	Re_address = r.N("ADDRESS", `g1`, r.P(r.C(`a-z0-9`)))       // starts with g1, all lowercase.

	// Compile at init to avoid runtime compilation.
	ReGnoUserPkgPath = Re_gnoUserPkgPath.Compile()
	ReGnoStdPkgPath  = Re_gnoStdPkgPath.Compile()
	ReAddress        = Re_address.Compile()
)

// IsRealmPath determines whether the given pkgpath is for a realm, and as such
// should persist the global state. It also excludes _test paths.
func IsRealmPath(pkgPath string) bool {
	match := ReGnoUserPkgPath.Match(pkgPath)
	if match == nil || match.Get("LETTER") != "r" {
		return false
	}
	if strings.HasSuffix(match.Get("REPO"), "_test") {
		return false
	}
	return true
}

// IsEphemeralPath determines whether the given pkgpath is for an ephemeral realm.
// Ephemeral realms are temporary and don't persist state between transactions.
func IsEphemeralPath(pkgPath string) bool {
	match := ReGnoUserPkgPath.Match(pkgPath)
	return match != nil && match.Get("LETTER") == "e"
}

// IsGnoRunPath returns true if it's a run (MsgRun) package path.
// DerivePkgAddress() returns the embedded address such that the run package can
// receive coins on behalf of the user.
// XXX XXX XXX XXX change DerivePkgAddress().
func IsGnoRunPath(pkgPath string) (addr string, ok bool) {
	match := ReGnoUserPkgPath.Match(pkgPath)
	if match == nil || match.Get("LETTER") != "e" || match.Get("REPO") != "run" {
		return "", false
	}
	addrmatch := ReAddress.Match(match.Get("USER"))
	if addrmatch == nil {
		return "", false
	}
	return addrmatch.Get("ADDRESS"), true
}

// IsInternalPath determines whether the given pkgPath refers to an internal
// package, that may not be called directly or imported by packages that don't
// share the same root.
//
// If isInternal is true, base will be set to the root of the internal package,
// which must also be an ancestor or the same path that imports the given
// internal package.
func IsInternalPath(pkgPath string) (base string, isInternal bool) {
	// Restrict imports to /internal packages to a package rooted at base.
	var suff string
	base, suff, isInternal = strings.Cut(pkgPath, "/internal")
	// /internal should be either at the end, or be a part: /internal/
	isInternal = isInternal && (suff == "" || suff[0] == '/')
	return
}

// IsPPackagePath determines whether the given pkgPath is for a published Gno package.
// It only considers "pure" those starting with gno.land/p/, so it returns false for
// stdlib packages, realm paths, and run paths. It also excludes _test paths.
func IsPPackagePath(pkgPath string) bool {
	match := ReGnoUserPkgPath.Match(pkgPath)
	if match == nil || match.Get("LETTER") != "p" {
		return false
	}
	if strings.HasSuffix(match.Get("REPO"), "_test") {
		return false
	}
	return true
}

// IsStdlib determines whether pkgPath is for a standard library.
// Dots are not allowed for stdlib paths.
func IsStdlib(pkgPath string) bool {
	match := ReGnoStdPkgPath.Match(pkgPath)
	return match != nil
}

// IsUserlib determines whether pkgPath is for a non-stdlib path.
// It must be of the form <domain>/<letter>/<user>(/<repo>).
func IsUserlib(pkgPath string) bool {
	match := ReGnoUserPkgPath.Match(pkgPath)
	return match != nil
}

func IsTestFile(file string) bool {
	return strings.HasSuffix(file, "_test.gno") || strings.HasSuffix(file, "_filetest.gno")
}

//----------------------------------------
// Package name and path validation helpers.
// See https://github.com/gnolang/gno/issues/1571

// reVersionSuffix matches version suffixes (v1, v2, v3, v10, v11, ...).
// Note: Go convention says v1 should not appear in paths, but Gno allows it
// for backwards compatibility with existing versioned packages.
var reVersionSuffix = regexp.MustCompile(`^v([1-9][0-9]*)$`)

// isVersionSuffix returns true if s is a version suffix (v1, v2, v3, ...).
func isVersionSuffix(s string) bool {
	return reVersionSuffix.MatchString(s)
}

// LastPathElement extracts the last meaningful element from a package path.
// For versioned paths like "gno.land/r/foo/v2" or "gno.land/r/foo/v1", it
// returns "foo" since version suffixes (v1, v2, ...) are skipped.
func LastPathElement(pkgPath string) string {
	parts := strings.Split(pkgPath, "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]

	if isVersionSuffix(last) && len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return last
}

// ValidatePkgNameMatchesPath ensures the declared package name matches the last path element.
// This prevents confusion where a package at "gno.land/r/foo" declares "package bar".
// For versioned paths (v1, v2, ...), the package name must match the element before
// the version suffix (e.g., "gno.land/r/foo/v2" expects "package foo").
func ValidatePkgNameMatchesPath(pkgName Name, pkgPath string) error {
	expectedName := LastPathElement(pkgPath)
	if expectedName == "" {
		return nil
	}
	if string(pkgName) != expectedName {
		return fmt.Errorf("package name %q does not match path element %q", pkgName, expectedName)
	}
	return nil
}

//----------------------------------------
// Mempackage basic file filters.

var (
	goodFiles = []string{
		"license",
		"license.txt",
		"licence",
		"licence.txt",
		"gno.mod",
	}
	// NOTE: Xtn is easier to type than Extension due to proximity of 'e'
	// and 'x'.  Our language is thus influenced by the layout of the
	// "qwerty" keyboard, and perhaps different keyboards affect language
	// evolution differently.
	goodFileXtns = []string{
		".gno",
		".toml",
		".md",
		// ".txtar", // XXX: to be considered
	}
	badFileXtns = []string{
		".gen.go",
	}
)

// When running a mempackage (and thus in knowing what to parse), a filter
// applied must be one of these declared.
//
//  * MPFNone: Without any filtering, a package of type MP*Any will include
//  test files including xxx_test package test files, as well as *_filetest.gno
//  filetests.
//
//  * MPFProd: When running a mempackage in production mode, use MPFProd to
//  filter out all *_tests.gno and *_filetests.gno files. No test extension
//  overrides are present.
//
//  * MPFTest: When running a mempackage in testing mode, use MPFTest to filter
//  out all *_filetests.gno, and filter out all *_test files whose package name
//  is of the form "xxx_test". Notice that when running a test on a package,
//  the production declarations are amended with overrides in the *_test.gno
//  files, unless its package name is declared to be of the form
//  "mypackage_test" in order to aid with testing.
//
//  * MPFIntegration: When running tests declared in *_test.gno files with
//  package name xxx_test, these files are first collected into a new
//  mempackage with just these files of type MP*Integration, with package path
//  and package name ending in "_test". When these files import the original
//  package (with path minus "_test") they are of the type MP*Test (after
//  filtering with MPFTest).

type MemPackageFilter string

const (
	MPFNone        MemPackageFilter = "MPFNone"        // do not filter.
	MPFProd        MemPackageFilter = "MPFProd"        // filter _test.gno and _filetest.gno files.
	MPFTest        MemPackageFilter = "MPFTest"        // filter (xxx_test) _test.gno and _filetest.gno files.
	MPFIntegration MemPackageFilter = "MPFIntegration" // filter everything but xxx_test files.
)

func (mpfilter MemPackageFilter) Validate() {
	switch mpfilter {
	case MPFNone, MPFProd, MPFTest, MPFIntegration:
		// fine.
	default:
		panic(fmt.Sprintf("invalid mem package filter type %q", mpfilter))
	}
}

func (mpfilter MemPackageFilter) FilterGno(mfile *std.MemFile, pname Name) bool {
	fname := mfile.Name
	fbody := mfile.Body
	if !strings.HasSuffix(fname, ".gno") {
		panic("should not happen")
	}
	switch mpfilter {
	case MPFNone:
		return false
	case MPFProd:
		return endsWithAny(fname, []string{"_test.gno", "_filetest.gno"})
	case MPFTest:
		if endsWithAny(fname, []string{"_filetest.gno"}) {
			return true
		}
		return isIntegrationTestFile(pname, fname, fbody)
	case MPFIntegration:
		if !endsWithAny(fname, []string{"_test.gno"}) {
			return true
		}
		return !isIntegrationTestFile(pname, fname, fbody)
	default:
		panic("should not happen")
	}
}

func isIntegrationTestFile(pname Name, fname, fbody string) bool {
	pname2, err := PackageNameFromFileBody(fname, fbody)
	if err != nil {
		panic(err)
	}
	switch pname2 {
	case pname:
		return false
	case pname + "_test":
		return true
	default:
		panic(fmt.Sprintf("unexpected package name %q in package with name %q", pname2, pname))
	}
}

func (mpfilter MemPackageFilter) FilterType(mptype MemPackageType) MemPackageType {
	switch mpfilter {
	case MPFNone:
		return mptype
	case MPFProd:
		switch mptype {
		case MPAnyAll, MPAnyTest, MPAnyProd, MPAnyIntegration:
			panic("should not happen (undecided MPAny*)")
		case MPUserAll, MPUserTest, MPUserProd:
			return MPUserProd
		case MPStdlibAll, MPStdlibTest, MPStdlibProd:
			return MPStdlibProd
		case MPUserIntegration, MPStdlibIntegration:
			panic("MP*Integration packages have no prod files")
		case MPFiletests:
			panic("should not happen")
		}
	case MPFTest:
		switch mptype {
		case MPAnyAll, MPAnyTest, MPAnyProd, MPAnyIntegration:
			panic("should not happen (undecided MPAny*)")
		case MPUserAll, MPUserTest:
			return MPUserTest
		case MPStdlibAll, MPStdlibTest:
			return MPStdlibTest
		case MPUserProd, MPStdlibProd:
			panic("MP*Prod packages have no test files")
		case MPUserIntegration, MPStdlibIntegration:
			panic("MP*Integration packages have no prod/test files")
		case MPFiletests:
			panic("should not happen")
		}
	case MPFIntegration:
		switch mptype {
		case MPAnyAll, MPAnyTest, MPAnyProd, MPAnyIntegration:
			panic("should not happen (undecided MPAny*)")
		case MPUserAll:
			return MPUserIntegration
		case MPStdlibAll:
			return MPStdlibIntegration
		case MPUserProd, MPUserTest, MPStdlibProd, MPStdlibTest:
			panic("MP*Prod and MP*Test packages have no integration test files")
		case MPFiletests:
			panic("should not happen")
		}
	default:
		panic("should not happen")
	}
	panic("should not happen")
}

// NOTE: only filters .gno files.
func (mpfilter MemPackageFilter) FilterMemPackage(mpkg *std.MemPackage) *std.MemPackage {
	if mpkg == nil {
		return nil
	}
	mpkg2 := &std.MemPackage{
		Name:  mpkg.Name,
		Path:  mpkg.Path,
		Files: nil,
		Type:  mpfilter.FilterType(mpkg.Type.(MemPackageType)),
		Info:  mpkg.Info,
	}
	for _, mfile := range mpkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			// just copy non-gno files.
			mpkg2.Files = append(mpkg2.Files, mfile.Copy())
		} else if mpfilter.FilterGno(mfile, Name(mpkg.Name)) {
			continue
		} else {
			mpkg2.Files = append(mpkg2.Files, mfile.Copy())
		}
	}
	return mpkg2
}

// While std.MemPackage can contain any data, gnolang/mempackage.go expects
// these to be of a certain form. Except for MPAny*, which is not a valid
// mempackage type but a convenience argument value that must resolve either to
// MPStdlib*, MPUser*; the mempackage types represent
// different classes of mempackages.
//
//  * MPUserAll: mpkg is a non-stdlib library of the form <domain>/<letter>/...
//  No filter was applied, and even *_filetest.gno files are present, so the
//  package is suitable for saving, but not running.
//
//  * MPUserTest: mpkg is a non-stdlib library of the form
//  <domain>/<letter>/...  MPFTest filter was already applied, and no
//  *_filetest.gno are present.  Validation will fail if any unexpected files
//  are present. *_test.gno files must declare themselves to be of the same
//  package name as non-test files and not end with "_test".  package name may
//  not end with _test. MPUserTest test files run on the package with all test
//  overrides applied. MPUserIntegration when importing a package import
//  MPUserTest type packages; likewise for *_filetest.gno filetests in
//  MPUserAll.
//
//  * MPUserIntegration: mpkg is a non-stdlib library of the form
//  <domain>/<letter>/..._test (must end with _test). MPFIntegration filter was
//  already applied, and no prod files, non-xxx_test *test.gno files, nor
//  *_filetest.gno are present.  Validation will fail if any unexpected files
//  are present.  *_test.gno files must have package name ending with "_test",
//  and are referred to as "xxx_test" test files or "integration tests".  When
//  these files import a package of the same package path (minus _test), they
//  import MPUserTest type packages.
//
//  * MPUserProd: mpkg is a non-stdlib library of the form
//  <domain>/<letter>/...  MPFProd filter was already applied, and no tests
//  (*_test.gno or *_filetest.gno) are present. Validation will fail if any
//  test files are present at all. This is what gets m.RunMemPackage()'d when a
//  package is imported by the `import` statement.
//
//  * MPStdlibAll: mpkg is a stdlib library. These are only handled by the
//  gnolang project and is declared as such separately. This is handled
//  separately for defensive purposes and convenience, so that validation may
//  fail if a mempackage declares itself to be IsStdlib yet a stdlib wasn't
//  expected.  Only MPStdlib* can include native .go files.
//
//  *MPStdlibTest: like MPUserTest, is for testing.
//
//  *MPStdlibIntegration: like MPUserIntegration, is for testing.
//
//  *MPStdlibProd: like MPUserTest, is for prod; for user/stdlib to import.
//
//  * MPFiletests: mpkg is a special kind of mempackage that contains only
//  filetests, and does not represent a package otherwise: it does not make
//  sense to apply filters MPFProd or MPFTest on this type of mempackage. These
//  files if they were included in a normal package would require the suffix
//  "_filetest.gno" in the file name, but that rule does not apply to files in
//  this type of mempackage. Unlike *_testfiles.gno in MP*All, when filetests
//  in MPFiletests import a package, the imported package is of MP*Prod type.
//
//  *MPAnyAll: is not a valid type but can be used in liue of MPStdlibAll or
//  MPUserAll for Read and Validate commands. It is only recommended for
//  testing purposes; in production logic you wouldn't want to validate and
//  save with this value because user package submission should not share the
//  same code path as for stdlib package registrations.
//
//  *MPAnyProd: similar to MPAnyAll, but for MPStdlibProd or MPUserProd.
//  *MPAnyTest: similar to MPAnyAll, but for MPStdlibTest or MPUserTest.
//
// All of the above concrete types except for MPFiletests must have consistent
// package names in all of the .gno files, otherwise the package fails
// validation.

type MemPackageType string

const (
	MPAnyAll            MemPackageType = "MPAnyAll"            // MPUserAll or MPStdlibAll.
	MPAnyProd           MemPackageType = "MPAnyProd"           // MPUserProd or MPStdlibProd.
	MPAnyTest           MemPackageType = "MPAnyTest"           // MPUserTest or MPStdlibTest.
	MPAnyIntegration    MemPackageType = "MPAnyIntegration"    // MPUserIntegration or MPStdlibIntegration.
	MPStdlibAll         MemPackageType = "MPStdlibAll"         // stdlibs only, all files.
	MPStdlibProd        MemPackageType = "MPStdlibProd"        // stdlibs only, no tests/filetests
	MPStdlibTest        MemPackageType = "MPStdlibTest"        // stdlibs only, w/ tests, w/o integration/filetests
	MPStdlibIntegration MemPackageType = "MPStdlibIntegration" // stdlibs only, only integration tests.
	MPUserAll           MemPackageType = "MPUserAll"           // no stdlibs, gno pkg path, w/ tests/filetests.
	MPUserProd          MemPackageType = "MPUserProd"          // no stdlibs, gno pkg path, no tests/filetests.
	MPUserTest          MemPackageType = "MPUserTest"          // no stdlibs, gno pkg path, w/ tests, w/o integration/filetests.
	MPUserIntegration   MemPackageType = "MPUserIntegration"   // no stdlibs, only integration tests.
	MPFiletests         MemPackageType = "MPFiletests"         // filetests only, regardless of file name (tests/files).
)

// NOTE: MPAny* are meant to be decided as MPStdlib* or MPUser*.
func (mptype MemPackageType) IsAny() bool {
	return mptype == MPAnyAll || mptype == MPAnyProd || mptype == MPAnyTest || mptype == MPAnyIntegration
}

func (mptype MemPackageType) AssertNotAny() {
	if mptype.IsAny() {
		panic(fmt.Sprintf("undefined any: %#v", mptype))
	}
}

func (mptype MemPackageType) Decide(pkgPath string) MemPackageType {
	switch mptype {
	case MPAnyAll:
		switch {
		case IsStdlib(pkgPath):
			return MPStdlibAll
		case IsUserlib(pkgPath):
			return MPUserAll
		default:
			panic(fmt.Sprintf("invalid package path %q", pkgPath))
		}
	case MPAnyProd:
		switch {
		case IsStdlib(pkgPath):
			return MPStdlibProd
		case IsUserlib(pkgPath):
			return MPUserProd
		default:
			panic(fmt.Sprintf("invalid package path %q", pkgPath))
		}
	case MPAnyIntegration:
		switch {
		case IsStdlib(pkgPath):
			return MPStdlibIntegration
		case IsUserlib(pkgPath):
			return MPUserIntegration
		default:
			panic(fmt.Sprintf("invalid package path %q", pkgPath))
		}
	case MPStdlibAll, MPStdlibProd, MPStdlibTest, MPStdlibIntegration,
		MPUserAll, MPUserProd, MPUserTest, MPUserIntegration:
		return mptype
	default:
		// e.g. doesn't make sense to decide for MPFiletests.
		panic("unexpected mptype")
	}
}

func (mptype MemPackageType) IsStdlib() bool {
	mptype.AssertNotAny()
	return mptype == MPStdlibAll || mptype == MPStdlibProd || mptype == MPStdlibTest || mptype == MPStdlibIntegration
}

func (mptype MemPackageType) IsUserlib() bool {
	mptype.AssertNotAny()
	return mptype == MPUserAll || mptype == MPUserProd || mptype == MPUserTest || mptype == MPUserIntegration
}

func (mptype MemPackageType) IsAll() bool {
	mptype.AssertNotAny()
	return mptype == MPUserAll || mptype == MPStdlibAll
}

func (mptype MemPackageType) IsProd() bool {
	mptype.AssertNotAny()
	return mptype == MPUserProd || mptype == MPStdlibProd
}

func (mptype MemPackageType) IsTest() bool {
	mptype.AssertNotAny()
	return mptype == MPUserTest || mptype == MPStdlibTest
}

func (mptype MemPackageType) IsIntegration() bool {
	mptype.AssertNotAny()
	return mptype == MPUserIntegration || mptype == MPStdlibIntegration
}

func (mptype MemPackageType) IsFiletests() bool {
	mptype.AssertNotAny()
	return mptype == MPFiletests
}

func (mptype MemPackageType) IsRunnable() bool {
	mptype.AssertNotAny()
	return mptype.IsTest() || mptype.IsProd() || mptype.IsIntegration()
}

func (mptype MemPackageType) IsStorable() bool {
	// MPAny* is not a valid mpkg type for storage,
	// e.g. mpkg.Type should never be MPAnyAll.
	mptype.AssertNotAny()
	// MP*Prod is stored by pkg/test/imports.go.
	// MP*Test is stored by pkg/test/tests.go.
	return mptype.IsAll() || mptype.IsProd() || mptype.IsTest()
	// MP*Integration has no reason to be stored.
}

func (mptype MemPackageType) AsRunnable() MemPackageType {
	// If All, demote to Prod.
	// If Test, keep as is.
	if mptype.IsAll() {
		switch mptype {
		case MPStdlibAll:
			return MPStdlibProd
		case MPUserAll:
			return MPUserProd
		default:
			panic("should not happen")
		}
	} else if mptype.IsProd() {
		return mptype
	} else if mptype.IsTest() {
		return mptype
	} else if mptype.IsIntegration() {
		return mptype
	} else {
		panic(fmt.Sprintf("mempackage type is not runnable: %v", mptype))
	}
}

// Validates that mptype is a valid MemPackageType; includes MPAny*, MPFiletests, etc.
// and that mptype is compatible with pkgPath.
func (mptype MemPackageType) Validate(pkgPath string) {
	// "_test" suffix is only allowed for integration.
	if mptype.IsIntegration() || mptype == MPAnyIntegration {
		if !strings.HasSuffix(pkgPath, "_test") {
			panic(fmt.Sprintf("integration package path must end with %q but got %q", "_test", pkgPath))
		}
	} else if strings.HasSuffix(pkgPath, "_test") {
		panic(fmt.Sprintf("only integration package types may end with %q but got %q", "_test", pkgPath))
	}
	// Check if MPUser*.
	switch {
	case mptype.IsUserlib():
		if !IsUserlib(pkgPath) {
			panic(fmt.Sprintf("expected user package path for %q but got %q", mptype, pkgPath))
		}
	case mptype.IsStdlib():
		if !IsStdlib(pkgPath) {
			panic(fmt.Sprintf("expected stdlib package path for %q but got %q", mptype, pkgPath))
		}
	default:
		panic("should not happen")
	}
	switch mptype {
	case MPAnyAll, MPAnyProd, MPAnyTest:
	case MPAnyIntegration:
	case MPUserAll, MPUserProd, MPUserTest:
	case MPUserIntegration:
	case MPStdlibAll, MPStdlibProd, MPStdlibTest:
	case MPStdlibIntegration:
	case MPFiletests:
	default:
		panic(fmt.Sprintf("invalid mem package type %q", mptype))
	}
}

// Given mpkg.Type (mptype), ensures that pkgPath matches mptype, and also that
// mptype2 is a compatible sub-type, or any.  This function is used by
// ParseMemPackageAsType to parse a subset of files in mpkg.
//   - if mpkg.Type .IsFiletests(), mptype2 MUST be MPFiletests.
//   - If mpkg.Type .IsAll(), mptype2 can be .IsAll(), .IsTest(), or .IsProd().
//   - if mpkg.Type .IsProd(), mptype2 MUST be .IsProd().
//   - if mpkg.Type .IsTest(), mptype2 can be .IsTest() or .IsProd().
//   - if mpkg.Type .IsIntegration(), mptype2 MUST be .IsIntegration(), and pkgPath suffix "_test".
//   - if mpkg.Type .IsStdlib(), mptype2 MUST be .IsAny() or .IsStdlib(), and pkgPath too.
//   - if mpkg.Type .IsUserlib(), mptype2 MUST be .IsAny() or .IsUserlib(), and pkgPath too.
func (mptype MemPackageType) AssertCompatible(pkgPath string, mptype2 MemPackageType) {
	mptype.Validate(pkgPath)
	mptype2.Validate(pkgPath)
	mptype2 = mptype2.Decide(pkgPath)
	if mptype.IsFiletests() && !mptype2.IsFiletests() ||
		mptype.IsAll() && !mptype2.IsAll() && !mptype2.IsTest() && !mptype2.IsProd() ||
		mptype.IsProd() && !mptype2.IsProd() ||
		mptype.IsTest() && !mptype2.IsTest() && !mptype2.IsProd() ||
		mptype.IsIntegration() && !mptype2.IsIntegration() ||
		mptype.IsStdlib() && !mptype2.IsStdlib() ||
		mptype.IsUserlib() && !mptype2.IsUserlib() {
		panic(fmt.Sprintf("%v does not match %v", mptype, mptype2))
	}
}

// fname: the file name.
// pname: the pname as declared in the file.
func (mptype MemPackageType) ExcludeGno(fname string, pname Name) bool {
	if !strings.HasSuffix(fname, ".gno") {
		panic("should not happen")
	}
	switch mptype {
	case MPAnyAll, MPAnyProd, MPAnyTest, MPAnyIntegration:
		panic("unresolved MPAny*")
	case MPStdlibAll, MPUserAll, MPFiletests:
		// include all files.
		return false
	case MPStdlibProd, MPUserProd:
		// exclude all test files.
		return endsWithAny(fname, []string{"_test.gno", "_filetest.gno"})
	case MPStdlibTest, MPUserTest:
		// exclude filetest files, and xxx_test package names.
		return endsWithAny(fname, []string{"_filetest.gno"}) ||
			endsWithAny(string(pname), []string{"_test"})
	case MPStdlibIntegration, MPUserIntegration:
		// only xxx_test *_test.gno files.
		return endsWithAny(fname, []string{"_filetest.gno"}) ||
			!endsWithAny(fname, []string{"_test.gno"}) ||
			!endsWithAny(string(pname), []string{"_test"})
	default:
		panic("should not happen")
	}
}

// ReadMemPackage initializes a new MemPackage by reading the OS directory at
// dir, and saving it with the given pkgPath (import path).  The resulting
// MemPackage will contain the names and content of all *.gno files, and
// additionally LICENSE, *.md and *.toml .
//
// ReadMemPackage only reads good file extensions or whitelisted good files,
// and ignores bad file extensions. Validation will fail if any bad extensions
// are found, but otherwise new files may be added by various logic. It also
// ignores and does not include files that wouldn't pass validation before any
// any filters applied. Unless MPFiletests, the package name declared in each
// file must be consistent with others, or nil and an error is returned.
//
// Filtering, parsing, and validation is performed separately.
func ReadMemPackage(dir string, pkgPath string, mptype MemPackageType) (*std.MemPackage, error) {
	mptype = mptype.Decide(pkgPath)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	// Shadow defense.
	goodFiles := goodFiles
	// Stdlib pkgpath validation.
	if !mptype.IsStdlib() && IsStdlib(pkgPath) {
		panic(fmt.Sprintf("unexpected stdlib package path %q for mempackage type %q", pkgPath, mptype))
	} else if mptype.IsStdlib() && !IsStdlib(pkgPath) {
		panic(fmt.Sprintf("unexpected non-stdlib package path %q", pkgPath))
	}
	// Integration pkgpath validation.
	if mptype.IsIntegration() && !strings.HasSuffix(pkgPath, "_test") {
		panic(fmt.Sprintf("unexpected package path %q for mempackage type %q (expected suffix %q)", pkgPath, mptype, "_test"))
	}
	// Allows transpilation to work on stdlibs with native fns.
	if IsStdlib(pkgPath) {
		goodFiles = append(goodFiles, ".go")
	}
	// Construct list of files to add to mpkg.
	list := make([]string, 0, len(files))
	for _, file := range files {
		// Ignore directories and hidden files, only include allowed files & extensions,
		// then exclude files that are of the bad extensions.
		// We do case ignore to check goodFiles. MemFile ValidateBasic will enforce case rules.
		if file.IsDir() ||
			strings.HasPrefix(file.Name(), ".") ||
			(!endsWithAny(file.Name(), goodFileXtns) &&
				!slices.Contains(goodFiles, strings.ToLower(file.Name()))) ||
			endsWithAny(file.Name(), badFileXtns) {
			continue
		}
		list = append(list, filepath.Join(dir, file.Name()))
	}
	return ReadMemPackageFromList(list, pkgPath, mptype)
}

func endsWithAny(str string, suffixes []string) bool {
	return slices.ContainsFunc(suffixes, func(s string) bool {
		return strings.HasSuffix(str, s)
	})
}

// MustReadMemPackage is a wrapper around [ReadMemPackage] that panics on error.
func MustReadMemPackage(dir string, pkgPath string, mptype MemPackageType) *std.MemPackage {
	pkg, err := ReadMemPackage(dir, pkgPath, mptype)
	if err != nil {
		panic(err)
	}
	return pkg
}

// ReadMemPackageFromList creates a new [std.MemPackage] with the specified
// pkgPath, containing the contents of all the files provided in the list
// slice.
//
// ReadMemPackageFromList only reads good file extensions or whitelisted good
// files, and ignores bad file extensions. Validation will fail if any bad
// extensions are found, but otherwise new files may be added by various logic.
// It also ignores and does not include files that wouldn't pass validation
// before any filters applied. Unless MPFiletests, the package name declared in
// each file must be consistent with others, an err will be returned.
//
// Filtering, parsing, and validation is performed separately.
//
// NOTE: panics if package name is invalid (characters must be alphanumeric or
// _, lowercase, and must start with a letter).
func ReadMemPackageFromList(list []string, pkgPath string, mptype MemPackageType) (*std.MemPackage, error) {
	mptype.Validate(pkgPath)
	mptype = mptype.Decide(pkgPath)
	mpkg := &std.MemPackage{
		Type: mptype,
		Path: pkgPath,
	}
	var pkgName Name          // normal file pkg name
	var pkgNameDiffers bool   // normal file pkg name is inconsistent
	var pkgNameFT Name        // filetest pkg name
	var pkgNameFTDiffers bool // filetest pkg name is inconsistent
	var errs error            // all errors minus filetest pkg name errors.
	for _, fpath := range list {
		fname := filepath.Base(fpath)
		bz, err := os.ReadFile(fpath)
		if err != nil {
			return nil, err
		}
		// Check that all pkg names are the same (else package is invalid).
		// Try to derive the package name, but this is not a replacement
		// for gno.ValidateMemPackage().
		if strings.HasSuffix(fname, ".gno") {
			//--------------------------------------------------------------------------------
			// NOTE: the below is (almost) duplicated in ParseMemPackageAsType().
			// If MPProd, don't even try to read _test.gno and _filetest.gno files.
			if mptype.IsProd() &&
				endsWithAny(fname, []string{"_test.gno", "_filetest.gno"}) {
				continue
			}
			// If MPTest, don't even try to read _filetest.gno files.
			if mptype.IsTest() &&
				endsWithAny(fname, []string{"_filetest.gno"}) {
				continue
			}
			// If MPIntegration, only read _test.gno files.
			if mptype.IsIntegration() &&
				!endsWithAny(fname, []string{"_test.gno"}) {
				continue
			}
			// Read package name from file.
			var pname2 Name
			pname2, err = PackageNameFromFileBody(path.Join(pkgPath, fname), string(bz))
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			// Ignore files that aren't suitable for mem package type.
			if mptype.ExcludeGno(fname, pname2) {
				continue
			}
			// NOTE: the above is (almost) duplicated in ParseMemPackageAsType().
			//--------------------------------------------------------------------------------
			// Try to derive the mem package name from suitable files.
			if mptype.IsFiletests() || strings.HasSuffix(fname, "_filetest.gno") {
				// Filetests may have arbitrary package names.
				// pname2 (of this file) may be unrelated to
				// pkgName of the mem package.
				if pkgNameFT == "" && !pkgNameFTDiffers {
					pkgNameFT = pname2
				} else if pkgNameFT != pname2 {
					pkgNameFT = ""
					pkgNameFTDiffers = true
				}
			} else {
				if strings.HasSuffix(string(pname2), "_test") {
					pname2 = pname2[:len(pname2)-len("_test")]
				}
				if pkgName == "" {
					pkgName = pname2
				} else if pkgName != pname2 {
					pkgNameDiffers = true
					errs = multierr.Append(errs, fmt.Errorf("%s:0: expected package name %q but got %q", fpath, pkgName, pname2))
					return nil, errs
				}
			}
		}
		mpkg.Files = append(mpkg.Files,
			&std.MemFile{
				Name: fname,
				Body: string(bz),
			})
	}

	// If there were any errors so far, return error.
	if errs != nil {
		return nil, errs
	}
	// If mpkg is empty, return an error
	if mpkg.IsEmpty() {
		return nil, fmt.Errorf("package has no files")
	}
	// Verify/derive package name.
	if pkgName == "" {
		// Inconsistent package name in prod files.
		if pkgNameDiffers {
			// Not actually reachable, but for defensive purposes.
			return nil, errs
		}
		// There are no prod/test files, only possible filetests.
		if !pkgNameFTDiffers {
			// If filetest pkgnames are consistent, use that.
			pkgName = pkgNameFT
		} else {
			// Set a default one. It doesn't matter.
			pkgName = "filetests"
		}
		// NOTE: mptype may be MPFiletests or anything.  In the future
		// we may make it illegal for anything but MPFiletests to have
		// no prod/test files.
	}
	// Still no pkgName or invalid; ensure error.
	if pkgName == "" {
		return nil, fmt.Errorf("package name could be determined")
	} else if err := validatePkgName(pkgName); err != nil {
		return nil, err
	}
	// Finally, set the name.
	mpkg.Name = string(pkgName)
	// Sort files and return.
	mpkg.Sort()
	return mpkg, nil
}

// MustReadMemPackageFromList is a wrapper around [ReadMemPackageFromList] that panics on error.
func MustReadMemPackageFromList(list []string, pkgPath string, mptype MemPackageType) *std.MemPackage {
	pkg, err := ReadMemPackageFromList(list, pkgPath, mptype)
	if err != nil {
		panic(err)
	}
	return pkg
}

// ParseMemPackage executes [ParseFile] on each file of the mpkg.
//
// If one of the files has a different package name than mpkg.Name,
// or [ParseFile] returns an error, ParseMemPackageAsType panics.
func (m *Machine) ParseMemPackage(mpkg *std.MemPackage) (fset *FileSet) {
	return m.ParseMemPackageAsType(mpkg, mpkg.Type.(MemPackageType))
}

// ParseMemPackageAsType executes [ParseFile] on each file of the mpkg, with
// files filtered based on mptype, which must match mpkg.Type. See also
// MemPackageType.Matches() for details.
//
// If one of the files has a different package name than mpkg.Name,
// or [ParseFile] returns an error, ParseMemPackageAsType panics.
func (m *Machine) ParseMemPackageAsType(mpkg *std.MemPackage, mptype MemPackageType) (fset *FileSet) {
	pkgPath := mpkg.Path
	mptype.Validate(pkgPath)
	mpkg.Type.(MemPackageType).AssertCompatible(mpkg.Path, mptype)
	fset = &FileSet{}
	var errs error
	for _, mfile := range mpkg.Files {
		fname := mfile.Name
		// Can't parse non-gno files.
		if !strings.HasSuffix(fname, ".gno") ||
			mfile.Name == "gno.mod" {
			continue // skip spurious or test or gno.mod file.
		}
		//--------------------------------------------------------------------------------
		// NOTE: the below is (almost) duplicated in ReadMemPackageFromList().
		// If MP*Prod, don't even try to read _test.gno and _filetest.gno files.
		if mptype.IsProd() &&
			endsWithAny(fname, []string{"_test.gno", "_filetest.gno"}) {
			continue
		}
		// If MP*Test, don't even try to read _filetest.gno files.
		if mptype.IsTest() &&
			endsWithAny(fname, []string{"_filetest.gno"}) {
			continue
		}
		// Read package name from file.
		pname2, err := PackageNameFromFileBody(path.Join(pkgPath, fname), mfile.Body)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		// Ignore files that aren't suitable for mem package type.
		if mptype.ExcludeGno(fname, pname2) {
			continue
		}
		// NOTE: the above is (almost) duplicated in ReadMemPackageFromList().
		//--------------------------------------------------------------------------------
		// Parse the file.
		n, err := m.ParseFile(mfile.Name, mfile.Body)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		// Package name *must* be consistent.
		if mpkg.Name != string(n.PkgName) {
			panic(fmt.Sprintf(
				"expected package name [%s] but got [%s]",
				mpkg.Name, n.PkgName))
		}
		// add package file.
		fset.AddFiles(n)
	}
	if errs != nil {
		panic(errs)
	}
	return fset
}

// ParseMemPackageTests parses test files (skipping filetests) in the mpkg and splits
// the files into categories for testing.
func (m *Machine) ParseMemPackageTests(mpkg *std.MemPackage) (tset, itset *FileSet, itfiles, ftfiles []*std.MemFile) {
	tset = &FileSet{}
	itset = &FileSet{}
	var errs error
	for _, mfile := range mpkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // skip this file.
		}

		n, err := m.ParseFile(mfile.Name, mfile.Body)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		if n == nil {
			panic("should not happen")
		}
		switch {
		case strings.HasSuffix(mfile.Name, "_filetest.gno"):
			ftfiles = append(ftfiles, mfile)
		case strings.HasSuffix(mfile.Name, "_test.gno") && mpkg.Name == string(n.PkgName):
			tset.AddFiles(n)
		case strings.HasSuffix(mfile.Name, "_test.gno") && mpkg.Name+"_test" == string(n.PkgName):
			itset.AddFiles(n)
			itfiles = append(itfiles, mfile)
		case mpkg.Name == string(n.PkgName):
			// normal package file
		default:
			panic(fmt.Sprintf(
				"expected package name [%s] or [%s_test] but got [%s] file [%s]",
				mpkg.Name, mpkg.Name, n.PkgName, mfile))
		}
	}
	if errs != nil {
		panic(errs)
	}
	return
}

// Validates a non-stdlib production mempackage with no tests.
func ValidateMemPackage(mpkg *std.MemPackage) error {
	mptype := mpkg.Type.(MemPackageType)
	mptype.Validate(mpkg.Path)
	if mptype.IsAny() {
		return errors.New("undecided mptype")
	}
	if !mptype.IsProd() {
		return errors.New("expected prod")
	}
	if mptype.IsStdlib() {
		return errors.New("unexpected stdlib")
	}
	return ValidateMemPackageAny(mpkg)
}

// Validates everything about mpkg, including that all files are within the
// scope of its type.  It does not validate whether mpkg is runnable or
// storable.
func ValidateMemPackageAny(mpkg *std.MemPackage) (errs error) {
	// Check for file sorting, string lengths, uniqueness...
	err := mpkg.ValidateBasic()
	if err != nil {
		return err
	}
	// Validate mpkg path.
	if true && // none of these match...
		!ReGnoUserPkgPath.Matches(mpkg.Path) &&
		!ReGnoStdPkgPath.Matches(mpkg.Path) {
		// .ValidateBasic() ensured rePkgPathURL or stdlib path,
		// but reGnoPkgPathStd is more restrictive.
		return fmt.Errorf("invalid package/realm path %q", mpkg.Path)
	}
	// Check mpkg.Type/mptype.
	mptype := mpkg.Type.(MemPackageType)
	mptype.Validate(mpkg.Path)
	// ...
	goodFileXtns := goodFileXtns
	if mptype.IsStdlib() { // Allow transpilation to work on stdlib with native functions.
		goodFileXtns = append(goodFileXtns, ".go")
	}
	// Validate package name.
	if err := validatePkgName(Name(mpkg.Name)); err != nil {
		return err
	}

	// NOTE: Package name vs path element validation is done separately in
	// keeper.go for deployment and lint.go for linting, because existing
	// packages may legitimately have different internal names than their path
	// element (e.g., filtests/extern).

	// Validate files.
	if mpkg.IsEmpty() {
		return fmt.Errorf("package has no files")
	}
	numGnoFiles := 0
	pkgNameFound := false
	for _, mfile := range mpkg.Files {
		// Validate file name.
		fname := mfile.Name
		if endsWithAny(fname, badFileXtns) {
			errs = multierr.Append(errs, fmt.Errorf("invalid file %q: illegal file extension", fname))
			continue
		}
		if strings.HasPrefix(fname, ".") {
			errs = multierr.Append(errs, fmt.Errorf("invalid file %q: file name cannot start with a dot", fname))
			continue
		}
		if strings.Contains(fname, "/") {
			errs = multierr.Append(errs, fmt.Errorf("invalid file %q: file name cannot contain a slash", fname))
			continue
		}
		if !endsWithAny(fname, goodFileXtns) {
			if !slices.Contains(goodFiles, strings.ToLower(fname)) {
				errs = multierr.Append(errs, fmt.Errorf("invalid file %q: unrecognized file type", fname))
				continue
			}
		}
		// Validate .gno package names.
		if strings.HasSuffix(fname, ".gno") {
			numGnoFiles += 1
			pkgName, err := PackageNameFromFileBody(path.Join(mpkg.Path, fname), mfile.Body)
			if err != nil {
				errs = multierr.Append(errs, err)
				continue
			}
			if pkgName != Name(mpkg.Name) { // Check validity but skip if mpkg.Name (already checked).
				if err := validatePkgName(pkgName); err != nil {
					errs = multierr.Append(errs, fmt.Errorf("invalid file %q: invalid package name", pkgName))
					continue
				}
			}
			// Validate and check package name.
			if mptype.ExcludeGno(fname, pkgName) {
				// Panic on unexpected files.
				errs = multierr.Append(errs, fmt.Errorf("invalid file %q: unexpected file given type %v", fname, mptype))
				continue
			} else if mptype.IsFiletests() || strings.HasSuffix(fname, "_filetest.gno") {
				// Any valid package name is OK for filetests.
				if pkgName == Name(mpkg.Name) {
					pkgNameFound = true
				}
			} else if strings.HasSuffix(fname, "_test.gno") {
				// Special case, xxx_test matches too.
				if pkgName == Name(mpkg.Name) || pkgName == Name(mpkg.Name)+"_test" {
					pkgNameFound = true
				} else { // since not filetest,
					errs = multierr.Append(errs, fmt.Errorf("invalid file %q: invalid package name", pkgName))
					continue
				}
			} else if pkgName == Name(mpkg.Name) {
				// General case, name found, or,
				pkgNameFound = true
				continue
			} else {
				// Doesn't belong here.
				errs = multierr.Append(errs, fmt.Errorf("invalid file %q: invalid package name", pkgName))
				continue
			}
		}
	}
	if numGnoFiles == 0 { // something else is probably wrong.
		errs = multierr.Append(errs, fmt.Errorf("package has no .gno files"))
	}
	if !mptype.IsFiletests() && !pkgNameFound { // strange.
		errs = multierr.Append(errs, fmt.Errorf("package name %q not found in files", mpkg.Name))
	}
	return errs
}

// PackageNameFromFileBody extracts the package name from the given Gno code body.
// The 'name' parameter is used for better error traces, and 'body' contains the Gno code.
func PackageNameFromFileBody(name, body string) (Name, error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, name, body, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}

	return Name(astFile.Name.Name), nil
}

// MustPackageNameFromFileBody is a wrapper around [PackageNameFromFileBody] that panics on error.
func MustPackageNameFromFileBody(name, body string) Name {
	pkgName, err := PackageNameFromFileBody(name, body)
	if err != nil {
		panic(err)
	}
	return pkgName
}

// ========================================
// WriteToMemPackage writes Go AST to a mempackage
// This is useful for preparing prior version code for the preprocessor.
func WriteToMemPackage(gofset *token.FileSet, gofs []*ast.File, mpkg *std.MemPackage, create bool) error {
	for _, gof := range gofs {
		fpath := gofset.File(gof.Pos()).Name()
		_, fname := filepath.Split(fpath)
		if strings.HasPrefix(fname, ".") {
			// Hidden files like .gnobuiltins.gno that
			// start with a dot should not get written to
			// the mempackage.
			continue
		}
		mfile := mpkg.GetFile(fname)
		if mfile == nil {
			if create {
				mfile = mpkg.NewFile(fname, "")
			} else {
				return fmt.Errorf("missing memfile %q", mfile)
			}
		}
		err := WriteToMemFile(gofset, gof, mfile)
		if err != nil {
			return fmt.Errorf("writing to mempackage %q: %w",
				mpkg.Path, err)
		}
	}
	return nil
}

func WriteToMemFile(gofset *token.FileSet, gof *ast.File, mfile *std.MemFile) error {
	var buf bytes.Buffer
	err := gofmt.Node(&buf, gofset, gof)
	if err != nil {
		return fmt.Errorf("writing to memfile %q: %w",
			mfile.Name, err)
	}
	mfile.Body = buf.String()
	return nil
}
