package std

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
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

var (
	rePkgName      = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	rePkgOrRlmPath = regexp.MustCompile(`^gno\.land\/(?:p|r)(?:\/_?[a-z]+[a-z0-9_]*)+$`)
	reFileName     = regexp.MustCompile(`^([a-zA-Z0-9_]*\.[a-z0-9_\.]*|LICENSE|README)$`)
)

// path must not contain any dots after the first domain component.
// file names must contain dots.
// NOTE: this is to prevent conflicts with nested paths.
func (mempkg *MemPackage) Validate() error {
	// add assertion that MemPkg contains at least 1 file
	if len(mempkg.Files) <= 0 {
		return fmt.Errorf("no files found within package %q", mempkg.Name)
	}

	if !rePkgName.MatchString(mempkg.Name) {
		return fmt.Errorf("invalid package name %q, failed to match %q", mempkg.Name, rePkgName)
	}
	if !rePkgOrRlmPath.MatchString(mempkg.Path) {
		return fmt.Errorf("invalid package/realm path %q, failed to match %q", mempkg.Path, rePkgOrRlmPath)
	}
	// enforce sorting files based on Go conventions for predictability
	sorted := sort.SliceIsSorted(
		mempkg.Files,
		func(i, j int) bool {
			return mempkg.Files[i].Name < mempkg.Files[j].Name
		},
	)
	if !sorted {
		return fmt.Errorf("mempackage %q has unsorted files", mempkg.Path)
	}

	var prev string
	for i, file := range mempkg.Files {
		if !reFileName.MatchString(file.Name) {
			return fmt.Errorf("invalid file name %q, failed to match %q", file.Name, reFileName)
		}
		if i > 0 && prev == file.Name {
			return fmt.Errorf("duplicate file name %q", file.Name)
		}
		prev = file.Name
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
