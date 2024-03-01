package std

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

type MemFile struct {
	Name string
	Body string
}

// MemPackage represents the information and files of a package which will be
// stored in memory. It will generally be initialized by package gnolang's
// ReadMemPackage.
//
// NOTE: in the future, a MemPackage may represent
// updates/additional-files for an existing package.
type MemPackage struct {
	Name  string // package name as declared by `package`
	Path  string // import path
	Files []*MemFile
}

func (mempkg *MemPackage) GetFile(name string) *MemFile {
	for _, memFile := range mempkg.Files {
		if memFile.Name == name {
			return memFile
		}
	}
	return nil
}

func (mempkg *MemPackage) IsEmpty() bool {
	return len(mempkg.Files) == 0
}

const rePathPart = `[a-z][a-z0-9_]*`

var (
	rePkgName      = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	rePkgOrRlmPath = regexp.MustCompile(`gno\.land/(?:p|r)(?:/` + rePathPart + `)+`)
	reFileName     = regexp.MustCompile(`^([a-zA-Z0-9_]*\.[a-z0-9_\.]*|LICENSE|README)$`)
)

// path must not contain any dots after the first domain component.
// file names must contain dots.
// NOTE: this is to prevent conflicts with nested paths.
func (mempkg *MemPackage) Validate() error {
	if !rePkgName.MatchString(mempkg.Name) {
		return errors.New(fmt.Sprintf("invalid package name %q, failed to match %q", mempkg.Name, rePkgName))
	}
	if !rePkgOrRlmPath.MatchString(mempkg.Path) {
		return errors.New(fmt.Sprintf("invalid package/realm path %q, failed to match %q", mempkg.Path, rePkgOrRlmPath))
	}
	fnames := map[string]struct{}{}
	for _, memfile := range mempkg.Files {
		if !reFileName.MatchString(memfile.Name) {
			return errors.New(fmt.Sprintf("invalid file name %q, failed to match %q", memfile.Name, reFileName))
		}
		if _, exists := fnames[memfile.Name]; exists {
			return errors.New(fmt.Sprintf("duplicate file name %q", memfile.Name))
		}
		fnames[memfile.Name] = struct{}{}
	}
	return nil
}

// Splits a path into the dirpath and filename.
func SplitFilepath(filepath string) (dirpath string, filename string) {
	parts := strings.Split(filepath, "/")
	if len(parts) == 1 {
		return parts[0], ""
	}
	last := parts[len(parts)-1]
	if strings.Contains(last, ".") {
		return strings.Join(parts[:len(parts)-1], "/"), last
	} else if last == "" {
		return strings.Join(parts[:len(parts)-1], "/"), ""
	} else {
		return strings.Join(parts, "/"), ""
	}
}
