package std

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// A MemFile is the simplest representation of a "file".
//
// Notice it doesn't have owners or timestamps. Keep this as is for
// portability.  Not even date created, ownership, or other attributes.  Just a
// name, and a body.  This keeps things portable, easy to hash (otherwise need
// to manage e.g. the encoding and portability of timestamps).
type MemFile struct {
	Name string `json:"name" yaml:"name"`
	Body string `json:"body" yaml:"body"`
}

// MemPackage represents the information and files of a package which will be
// stored in memory. It will generally be initialized by package gnolang's
// ReadMemPackage.
//
// NOTE: in the future, a MemPackage may represent
// updates/additional-files for an existing package.
type MemPackage struct {
	Name  string     `json:"name" yaml:"name"`   // package name as declared by `package`
	Path  string     `json:"path" yaml:"path"`   // import path
	Files []*MemFile `json:"files" yaml:"files"` // plain file system files.
}

const pathLengthLimit = 256

var (
	rePkgName      = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	rePkgOrRlmPath = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}\/(?:p|r)(?:\/_?[a-z]+[a-z0-9_]*)+$`)
	reFileName     = regexp.MustCompile(`^([a-zA-Z0-9_]*\.[a-z0-9_\.]*|LICENSE|README)$`)
)

// path must not contain any dots after the first domain component.
// file names must contain dots.
// NOTE: this is to prevent conflicts with nested paths.
func (mpkg *MemPackage) Validate() error {
	// add assertion that MemPkg contains at least 1 file
	if len(mpkg.Files) <= 0 {
		return fmt.Errorf("no files found within package %q", mpkg.Name)
	}

	if len(mpkg.Path) > pathLengthLimit {
		return fmt.Errorf("path length %d exceeds limit %d", len(mpkg.Path), pathLengthLimit)
	}

	if !rePkgName.MatchString(mpkg.Name) {
		return fmt.Errorf("invalid package name %q, failed to match %q", mpkg.Name, rePkgName)
	}

	if !rePkgOrRlmPath.MatchString(mpkg.Path) {
		return fmt.Errorf("invalid package/realm path %q, failed to match %q", mpkg.Path, rePkgOrRlmPath)
	}
	// enforce sorting files based on Go conventions for predictability
	sorted := sort.SliceIsSorted(
		mpkg.Files,
		func(i, j int) bool {
			return mpkg.Files[i].Name < mpkg.Files[j].Name
		},
	)
	if !sorted {
		return fmt.Errorf("mempackage %q has unsorted files", mpkg.Path)
	}

	var prev string
	for i, file := range mpkg.Files {
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

func (mpkg *MemPackage) GetFile(name string) *MemFile {
	for _, memFile := range mpkg.Files {
		if memFile.Name == name {
			return memFile
		}
	}
	return nil
}

func (mpkg *MemPackage) IsEmpty() bool {
	return mpkg.Name == "" || len(mpkg.Files) == 0
}

func (mpkg *MemPackage) WriteTo(dirPath string) error {
	for _, file := range mpkg.Files {
		fmt.Printf("MemPackage.WriteTo(%s) (%d) bytes written\n", file.Name, len(file.Body))
		fpath := filepath.Join(dirPath, file.Name)
		err := ioutil.WriteFile(fpath, []byte(file.Body), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

const licenseName = "LICENSE"

// Splits a path into the dirpath and filename.
func SplitFilepath(filepath string) (dirpath string, filename string) {
	parts := strings.Split(filepath, "/")
	if len(parts) == 1 {
		return parts[0], ""
	}

	switch last := parts[len(parts)-1]; {
	case strings.Contains(last, "."):
		return strings.Join(parts[:len(parts)-1], "/"), last
	case last == "":
		return strings.Join(parts[:len(parts)-1], "/"), ""
	case last == licenseName:
		return strings.Join(parts[:len(parts)-1], "/"), licenseName
	}

	return strings.Join(parts, "/"), ""
}
