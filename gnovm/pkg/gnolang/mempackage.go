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

	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
)

var (
	// NOTE: These are further restrictions upon the validation that already happens by std.MemPackage.Validate().
	// sub.domain.com/a/any
	// sub.domain.com/b/single
	// sub.domain.com/c/letter
	// sub.domain.com/d/works
	// sub.domain.com/r/realm
	// sub.domain.com/r/realm/path
	// sub.domain.com/p/package/path
	// See also tm2/pkg/std/memfile.go.
	// XXX test exhaustively balanced futureproof vs restrictive.
	reGnoPkgPathURL = regexp.MustCompile(`^([a-z0-9-]+\.)*[a-z0-9-]+\.[a-z]{2,}\/(?:[a-z])(?:\/_?[a-z][a-z0-9_]*)+$`)
	reGnoPkgPathStd = regexp.MustCompile(`^([a-z][a-z0-9_]*\/)*[a-z][a-z0-9_]+$`)
)

var (
	goodFiles = []string{
		"LICENSE",
		"README.md",
		"gno.mod",
	}
	// NOTE: Xtn is easier to type than Extension due to proximity of 'e'
	// and 'x'.  Our language is thus influenced by the layout of the
	// "qwerty" keyboard, and perhaps different keyboards affect language
	// evolution differently.
	goodFileXtns = []string{
		".gno",
		".toml",
		// ".txtar", // XXX: to be considered
	}
	badFileXtns = []string{
		".gen.go",
	}
)

// When running a mempackage (and thus in knowing what to parse), a filter
// applied must be one of these declared.
//
//  * MPFTest: When running a mempackage in testing mode, use MPFTest to filter
//  out all *_filetests.gno, and filter out all *_test files whose package name
//  is of the form "xxx_test". Notice that when running a test on a package,
//  the production declarations are amended with overrides in the *_test.gno
//  files, unless its package name is declared to be of the form
//  "mypackage_test" in order to aid with testing.
//
//  * MPFProd: When running a mempackage in production mode, use MPFProd to
//  filter out all *_tests.gno and *_filetests.gno files. No test extension
//  overrides are present.

type MemPackageFilter string

const (
	MPFNone MemPackageFilter = "MPFNone" // do not filter.
	MPFProd MemPackageFilter = "MPFProd" // filter _test.gno and _filetest.gno files.
	MPFTest MemPackageFilter = "MPFTest" // filter (xxx_test) _test.gno and _filetest.gno files.
)

func (mpfilter MemPackageFilter) Validate() {
	switch mpfilter {
	case MPFNone, MPFProd, MPFTest:
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
		pname2, err := PackageNameFromFileBody(fname, fbody)
		if err != nil {
			panic(err)
		}
		return pname2 != pname
	default:
		panic("should not happen")
	}
}

func (mpfilter MemPackageFilter) FilterType(mptype MemPackageType) MemPackageType {
	switch mpfilter {
	case MPFNone:
		return mptype
	case MPFProd:
		switch mptype {
		case MPAnyAll, MPAnyTest, MPAnyProd:
			panic("undecided MPAny*")
		case MPUserAll, MPUserTest, MPUserProd:
			return MPUserProd
		case MPStdlibAll, MPStdlibTest, MPStdlibProd:
			return MPStdlibProd
		case MPFiletests:
			panic("should not happen")
		}
	case MPFTest:
		switch mptype {
		case MPAnyAll, MPAnyTest, MPAnyProd:
			panic("undecided MPAny*")
		case MPUserAll, MPUserTest:
			return MPUserTest
		case MPStdlibAll, MPStdlibTest:
			return MPStdlibTest
		case MPUserProd, MPStdlibProd:
			panic("cannot filter for MPFTest on MP*Prod (no tests files)")
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
// MPStdlibAll, MPUserAll, or MPFiletesst; the mempackage types represent
// different classes of mempackages.
//
//  * MPUserAll: mpkg is a non-stdlib library of the form <domain>/<letter>/...
//  No filter was applied, and even *_filetest.gno files are present, so the
//  package is suitable for saving, but not running.
//
//  * MPUserTest: mpkg is a non-stdlib library of the form
//  <domain>/<letter>/...  MPFTest filter was already applied, and no
//  *_filetest.gno are present.  Validation will fail if any *_filetests.gno
//  files are present. *_test.gno files may declare themselves to be of the
//  same package name as non-test files, or have "_test" appended, and are
//  referred to as "xxx_test" package *_test.gno files.
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
//  *MPStdlibProd: like MPUserTest, is for prod; for user/stdlib to import.
//
//  * MPFiletests: mpkg is a special kind of mempackage that contains only
//  filetests, and does not represent a package otherwise: it does not make
//  sense to apply filters MPFProd or MPFTest on this type of mempackage. These
//  files if they were included in a normal package would require the suffix
//  "_filetest.gno" in the file name, but that rule does not apply to files in
//  this type of mempackage.
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
	MPAnyAll     MemPackageType = "MPAnyAll"     // MPUserAll or MPStdlibAll.
	MPAnyTest    MemPackageType = "MPAnyTest"    // MPUserTest or MPStdlibTest.
	MPAnyProd    MemPackageType = "MPAnyProd"    // MPUserProd or MPStdlibProd.
	MPStdlibAll  MemPackageType = "MPStdlibAll"  // stdlibs only, all files.
	MPStdlibTest MemPackageType = "MPStdlibTest" // stdlibs only, w/ tests, w/o integration/filetests
	MPStdlibProd MemPackageType = "MPStdlibProd" // stdlibs only, no tests/filetests
	MPUserAll    MemPackageType = "MPUserAll"    // no stdlibs, gno pkg path, w/ tests/filetests.
	MPUserTest   MemPackageType = "MPUserTest"   // no stdlibs, gno pkg path, w/ tests, w/o integration/filetests.
	MPUserProd   MemPackageType = "MPUserProd"   // no stdlibs, gno pkg path, no tests/filetests.
	MPFiletests  MemPackageType = "MPFiletests"  // filetests only, regardless of file name (tests/files).
)

// NOTE: MPAnyAll, MPAnyTest, MPAnyProd are meant to be decided as MPStdlib* or MPUser*.
func (mptype MemPackageType) IsAny() bool {
	return mptype == MPAnyAll || mptype == MPAnyTest || mptype == MPAnyProd
}
func (mptype MemPackageType) AssertNotAny() {
	if mptype.IsAny() {
		panic("undefined MPAny*")
	}
}
func (mptype MemPackageType) Decide(pkgPath string) MemPackageType {
	switch mptype {
	case MPAnyAll:
		if IsStdlib(pkgPath) {
			return MPStdlibAll
		} else { // XXX IsUserPath(), and default panic.
			return MPUserAll
		}
	case MPAnyTest:
		if IsStdlib(pkgPath) {
			return MPStdlibTest
		} else { // XXX IsUserPath(), and default panic.
			return MPUserTest
		}
	case MPAnyProd:
		if IsStdlib(pkgPath) {
			return MPStdlibProd
		} else { // XXX IsUserPath(), and default panic.
			return MPUserProd
		}
	case MPStdlibAll, MPStdlibTest, MPStdlibProd,
		MPUserAll, MPUserTest, MPUserProd:
		return mptype
	default:
		// e.g. doesn't make sense to decide for MPFiletests.
		panic("unexpected mptype")
	}
}
func (mptype MemPackageType) IsStdlib() bool {
	mptype.AssertNotAny()
	return mptype == MPStdlibAll || mptype == MPStdlibTest || mptype == MPStdlibProd
}
func (mptype MemPackageType) IsUser() bool {
	mptype.AssertNotAny()
	return mptype == MPUserAll || mptype == MPUserTest || mptype == MPUserProd
}
func (mptype MemPackageType) IsAll() bool {
	mptype.AssertNotAny()
	return mptype == MPUserAll || mptype == MPStdlibAll
}
func (mptype MemPackageType) IsTest() bool {
	mptype.AssertNotAny()
	return mptype == MPUserTest || mptype == MPStdlibTest
}
func (mptype MemPackageType) IsProd() bool {
	mptype.AssertNotAny()
	return mptype == MPUserProd || mptype == MPStdlibProd
}
func (mptype MemPackageType) IsFiletests() bool {
	mptype.AssertNotAny()
	return mptype == MPFiletests
}
func (mptype MemPackageType) IsRunnable() bool {
	mptype.AssertNotAny()
	return mptype.IsTest() || mptype.IsProd()
}
func (mptype MemPackageType) IsStorable() bool {
	// MPAny* is not a valid mpkg type for storage,
	// e.g. mpkg.Type should never be MPAnyAll.
	mptype.AssertNotAny()
	// MP*Prod is stored by pkg/test/imports.go.
	// MP*Test is stored by pkg/test/tests.go.
	return mptype.IsAll() || mptype.IsProd() || mptype.IsTest()
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
	} else {
		panic(fmt.Sprintf("mempackage type is not runnable: %v", mptype))
	}
}

// Validates that mptype is a valid MemPackageType; includes MPAny*, MPFiletests, etc.
func (mptype MemPackageType) Validate() {
	if !slices.Contains([]MemPackageType{
		MPAnyAll, MPAnyTest, MPAnyProd,
		MPStdlibAll, MPStdlibTest, MPStdlibProd,
		MPUserAll, MPUserTest, MPUserProd, MPFiletests,
	}, mptype) {
		panic(fmt.Sprintf("invalid mem package type %q", mptype))
	}
}

// Given mpkg.Type (mptype), ensures that pkgPath matches mptype, and also that
// mptype2 is a compatible sub-type, or any.  This function is used by
// ParseMemPackageAsType to parse a subset of files in mpkg.
//   - if mpkg.Type .IsFiletests(), mptype2 MUST be MPFiletests.
//   - If mpkg.Type .IsAll(), mptype2 can be .IsAll(), .IsTest(), or .IsProd().
//   - if mpkg.Type .IsTest(), mptype2 can be .IsTest() or .IsProd().
//   - if mpkg.Type .IsProd(), mptype2 MUST be .IsProd().
//   - if mpkg.Type .IsStdlib(), mptype2 MUST be .IsAny() or .IsStdlib(), and pkgPath too.
//   - if mpkg.Type .IsUser(), mptype2 MUST be .IsAny() or .IsUser(), and pkgPath too.
func (mptype MemPackageType) AssertCompatible(pkgPath string, mptype2 MemPackageType) {
	mptype.Validate()
	mptype2.Validate()
	if mptype.IsStdlib() && !IsStdlib(pkgPath) {
		panic(fmt.Sprintf("%v does not match non-stdlib %q", mptype, pkgPath))
	}
	if mptype.IsUser() && IsStdlib(pkgPath) {
		panic(fmt.Sprintf("%v does not match stdlib %q", mptype, pkgPath))
	}
	mptype2 = mptype2.Decide(pkgPath)
	if mptype.IsFiletests() && !mptype2.IsFiletests() ||
		mptype.IsAll() && !mptype2.IsAll() && !mptype2.IsTest() && !mptype2.IsProd() ||
		mptype.IsTest() && !mptype2.IsTest() && !mptype2.IsProd() ||
		mptype.IsProd() && !mptype2.IsProd() ||
		mptype.IsStdlib() && !mptype2.IsStdlib() ||
		mptype.IsUser() && !mptype2.IsUser() {
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
	case MPAnyAll, MPAnyTest, MPAnyProd:
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
	default:
		panic("should not happen")
	}
}

// ReadMemPackage initializes a new MemPackage by reading the OS directory at
// dir, and saving it with the given pkgPath (import path).  The resulting
// MemPackage will contain the names and content of all *.gno files, and
// additionally README.md, LICENSE.
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
	// shadow defense.
	goodFiles := goodFiles
	// Special stdlib validation.
	if !mptype.IsStdlib() && IsStdlib(pkgPath) {
		panic(fmt.Sprintf("unexpected stdlib package path %q for mempackage type %q", pkgPath, mptype))
	} else if mptype.IsStdlib() && !IsStdlib(pkgPath) {
		panic(fmt.Sprintf("unexpected non-stdlib package path %q", pkgPath))
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
		if file.IsDir() ||
			strings.HasPrefix(file.Name(), ".") ||
			(!endsWithAny(file.Name(), goodFileXtns) &&
				!slices.Contains(goodFiles, file.Name())) ||
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
	mptype.Validate()
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
		if mptype.IsFiletests() {
			// If only filetests with the same name, its package name is used.
			if !pkgNameFTDiffers {
				pkgName = pkgNameFT
			} else {
				// Otherwie, set a default one. It doesn't matter.
				pkgName = "filetests"
			}
		} else if pkgNameDiffers {
			return nil, errs
		}
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

// ParseMemPackageAsType executes [ParseFile] on each file of the mpkg, with
// files filtered based on mptype, which must match mpkg.Type. See also
// MemPackageType.Matches() for details.
//
// If one of the files has a different package name than mpkg.Name,
// or [ParseFile] returns an error, ParseMemPackageAsType panics.
func ParseMemPackageAsType(mpkg *std.MemPackage, mptype MemPackageType) (fset *FileSet) {
	pkgPath := mpkg.Path
	mptype.Validate()
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
		n, err := ParseFile(mfile.Name, mfile.Body)
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
func ParseMemPackageTests(mpkg *std.MemPackage) (tset, itset *FileSet, itfiles, ftfiles []*std.MemFile) {
	tset = &FileSet{}
	itset = &FileSet{}
	var errs error
	for _, mfile := range mpkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // skip this file.
		}

		n, err := ParseFile(mfile.Name, mfile.Body)
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
	mptype.Validate()
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
		!reGnoPkgPathURL.MatchString(mpkg.Path) &&
		!reGnoPkgPathStd.MatchString(mpkg.Path) {
		// .ValidateBasic() ensured rePkgPathURL or stdlib path,
		// but reGnoPkgPathStd is more restrictive.
		return fmt.Errorf("invalid package/realm path %q", mpkg.Path)
	}
	// Check mpkg.Type/mptype.
	mptype := mpkg.Type.(MemPackageType)
	mptype.Validate()
	// ...
	goodFileXtns := goodFileXtns
	if mptype.IsStdlib() { // Allow transpilation to work on stdlib with native functions.
		goodFileXtns = append(goodFileXtns, ".go")
	}
	// Validate package name.
	if err := validatePkgName(Name(mpkg.Name)); err != nil {
		return err
	}
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
			if !slices.Contains(goodFiles, fname) {
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
