package gnovm

import (
	"fmt"
	"go/parser"
	"go/token"
	"regexp"
	"sort"
	"strings"
)

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
	Name  string     `json:"name" yaml:"name"` // package name as declared by `package`
	Path  string     `json:"path" yaml:"path"` // import path
	Files []*MemFile `json:"files" yaml:"files"`
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

const pathLengthLimit = 256

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

	if len(mempkg.Path) > pathLengthLimit {
		return fmt.Errorf("path length %d exceeds limit %d", len(mempkg.Path), pathLengthLimit)
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

	if strings.Contains(mempkg.Path, "/p/") {
		pIndex := strings.Index(mempkg.Path, "/p/")
		if pIndex < 1 || strings.Count(mempkg.Path[:pIndex], "/") > 0 {
			return fmt.Errorf("invalid pkg path %q", mempkg.Path)
		}
		for _, file := range mempkg.Files {
			// only check .gno files, can contains files like LICENSE or README that will cause parse error
			if !strings.HasSuffix(file.Name, ".gno") {
				continue
			}
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, file.Name, file.Body, parser.ImportsOnly)
			if err != nil {
				return fmt.Errorf("unable to parse %q", file.Name)
			}
			for _, imp := range astFile.Imports {
				importPath := strings.TrimPrefix(strings.TrimSuffix(imp.Path.Value, `"`), `"`)
				// ensure the pkg is a realm by checking if the path contains /r/ and no other / character before it (i.e protect from gno.land/p/demo/r/)
				rIndex := strings.Index(importPath, "/r/")
				if rIndex > 0 && strings.Count(importPath[:rIndex], "/") == 0 {
					return fmt.Errorf("package %q imports realm %q", mempkg.Path, importPath)
				}

			}
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
