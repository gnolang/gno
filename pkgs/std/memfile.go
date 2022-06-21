package std

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gnolang/gno/pkgs/errors"
)

type MemFile struct {
	Name string
	Body string
}

// NOTE: in the future, a MemPackage may represent
// updates/additional-files for an existing package.
type MemPackage struct {
	Name  string
	Path  string
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

const (
	reDomainPart   = `gno\.land`
	rePathPart     = `[a-z][a-z0-9_]*`
	rePkgName      = `^[a-z][a-z0-9_]*$`
	rePkgPath      = reDomainPart + `/p/` + rePathPart + `(/` + rePathPart + `)*`
	reRlmPath      = reDomainPart + `/r/` + rePathPart
	rePkgOrRlmPath = `^(` + rePkgPath + `|` + reRlmPath + `)$`
	reFileName     = `^[a-zA-Z0-9_]*\.[a-z0-9_\.]*$`
)

// path must not contain any dots after the first domain component.
// file names must contain dots.
// NOTE: this is to prevent conflicts with nested paths.
func (mempkg *MemPackage) Validate() error {
	ok, _ := regexp.MatchString(rePkgName, mempkg.Name)
	if !ok {
		return errors.New(fmt.Sprintf("invalid package name %q", mempkg.Name))
	}
	ok, _ = regexp.MatchString(rePkgOrRlmPath, mempkg.Path)
	if !ok {
		return errors.New(fmt.Sprintf("invalid package/realm path %q", mempkg.Path))
	}
	fnames := map[string]struct{}{}
	re, err := regexp.Compile(reFileName)
	if err != nil {
		return fmt.Errorf("failed to compile regex to match file names: %w", err)
	}
	for _, memfile := range mempkg.Files {
		ok := re.MatchString(memfile.Name)
		if !ok {
			return errors.New(fmt.Sprintf("invalid file name %q", memfile.Name))
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
