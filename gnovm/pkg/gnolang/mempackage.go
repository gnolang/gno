package gnolang

import (
	"fmt"
	"path"
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
	allowedMemPackageFiles = []string{
		"LICENSE",
		"README.md",
		"gno.mod",
	}
	allowedMemPackageFileExtensions = []string{
		".gno",
	}
	badMemPackageFileExtensions = []string{
		".gen.go",
	}
)

type MemPackageType string

const (
	MemPackageTypeAny       MemPackageType = "MemPackageTypeAny"       // anything but not filetests only
	MemPackageTypeStdlib    MemPackageType = "MemPackageTypeStdlib"    // stdlibs only
	MemPackageTypeNormal    MemPackageType = "MemPackageTypeNormal"    // no stdlibs, gno pkg path, may include filetests
	MemPackageTypeFiletests MemPackageType = "MemPackageTypeFiletests" // filetests only
)

type ValidateMemPackageOptions struct {
	Type MemPackageType
}

// Validates a non-stdlib mempackage.
func ValidateMemPackage(mpkg *std.MemPackage) error {
	return ValidateMemPackageWithOptions(mpkg, ValidateMemPackageOptions{
		Type: MemPackageTypeNormal, // Keep this for defensiveness.
	})
}

func ValidateMemPackageWithOptions(mpkg *std.MemPackage, opts ValidateMemPackageOptions) (errs error) {
	// Check for file sorting, string lengths, uniqueness...
	err := mpkg.ValidateBasic()
	if err != nil {
		return err
	}
	// Validate mpkg path.
	if true && // none of these match...
		!reGnoPkgPathURL.MatchString(mpkg.Path) &&
		!reGnoPkgPathStd.MatchString(mpkg.Path) &&
		opts.Type != MemPackageTypeAny { // .ValidateBasic() ensured rePkgPathRUL
		return fmt.Errorf("invalid package/realm path %q", mpkg.Path)
	}
	// Check stdlib.
	isStdlib := IsStdlib(mpkg.Path)
	if isStdlib && !(opts.Type == MemPackageTypeStdlib || opts.Type == MemPackageTypeAny) {
		return fmt.Errorf("invalid package path %q: unexpected stdlib-type path", mpkg.Path)
	}
	if !isStdlib && opts.Type == MemPackageTypeStdlib {
		return fmt.Errorf("invalid package path %q: expected stdlib-type path", mpkg.Path)
	}
	allowedMemPackageFileExtensions := allowedMemPackageFileExtensions
	if isStdlib { // Allow transpilation to work on stdlib with native functions.
		allowedMemPackageFileExtensions = append(allowedMemPackageFileExtensions, ".go")
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
		if endsWithAny(fname, badMemPackageFileExtensions) {
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
		if !endsWithAny(fname, allowedMemPackageFileExtensions) {
			if !slices.Contains(allowedMemPackageFiles, fname) {
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
			if opts.Type == MemPackageTypeFiletests || strings.HasSuffix(fname, "_filetest.gno") {
				// Any valid package name is OK for filetests.
				if pkgName == Name(mpkg.Name) {
					pkgNameFound = true
				}
			} else if strings.HasSuffix(fname, "_test.gno") {
				if pkgName == Name(mpkg.Name) || pkgName == Name(mpkg.Name)+"_test" {
					pkgNameFound = true
				} else {
					errs = multierr.Append(errs, fmt.Errorf("invalid file %q: invalid package name", pkgName))
					continue
				}
			} else {
				if pkgName == Name(mpkg.Name) {
					pkgNameFound = true
				} else if opts.Type != MemPackageTypeFiletests {
					errs = multierr.Append(errs, fmt.Errorf("invalid file %q: invalid package name", pkgName))
					continue
				}
			}
		}
	}
	if numGnoFiles == 0 {
		errs = multierr.Append(errs, fmt.Errorf("package has no .gno files"))
	}
	if (opts.Type != MemPackageTypeFiletests) && !pkgNameFound {
		errs = multierr.Append(errs, fmt.Errorf("package name %q not found in files", mpkg.Name))
	}
	return errs
}
