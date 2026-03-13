package std

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// XXX rename to mempackage.go

const (
	fileNameLimit = 256
	pkgNameLimit  = 256
	pkgPathLimit  = 256
)

var (
	// See also gnovm/pkg/gnolang/mempackage.go.
	// NOTE: DO NOT MODIFY without a pre/post ADR and discussions with core GnoVM and gno.land teams.
	reFileName   = regexp.MustCompile(`^(([a-z0-9_\-]+|[A-Z0-9_\-]+)(\.[a-z0-9_]+)*\.[a-z0-9_]{1,7}|LICENSE|license|LICENCE|licence|README)$`)
	rePkgName    = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	rePkgPathURL = regexp.MustCompile(`^([a-z0-9-]+\.)*[a-z0-9-]+\.[a-z]{2,}(\/[a-z0-9\-_]+)+$`)
	rePkgPathStd = regexp.MustCompile(`^([a-z][a-z0-9_]*\/)*[a-z][a-z0-9_]+$`)
)

//----------------------------------------
// MemFile

// A MemFile is the simplest representation of a "file".
//
// File Name must contain a single dot and extension.
// File Name may be ALLCAPS.xxx or lowercase.xxx; extensions lowercase.
// e.g. OK:     README.md, README.md, readme.txt, READ.me
// e.g. NOT OK: Readme.md, readme.MD, README, .readme
// File Body can be any string.
//
// NOTE: It doesn't have owners or timestamps. Keep this as is for portability.
// Not even date created, ownership, or other attributes.  Just a name, and a
// body.  This keeps things portable, easy to hash (otherwise need to manage
// e.g. the encoding and portability of timestamps).
type MemFile struct {
	Name string `json:"name" yaml:"name"`
	Body string `json:"body" yaml:"body"`
}

func (mfile *MemFile) ValidateBasic() error {
	if len(mfile.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if len(mfile.Name) > fileNameLimit {
		return fmt.Errorf("name length %d exceeds limit %d", len(mfile.Name), fileNameLimit)
	}
	if !reFileName.MatchString(mfile.Name) {
		return fmt.Errorf("invalid file name %q", mfile.Name)
	}
	return nil
}

// Print file to stdout.
func (mfile *MemFile) Print() error {
	if mfile == nil {
		return fmt.Errorf("file not found")
	}
	fmt.Printf("MemFile[%q]:\n", mfile.Name)
	fmt.Println(mfile.Body)
	return nil
}

// Creates a new copy.
func (mfile *MemFile) Copy() *MemFile {
	mfile2 := *mfile
	return &mfile2
}

//----------------------------------------
// MemPackage

// MemPackage represents the information and files of a package which will be
// stored in memory. It will generally be initialized by package gnolang's
// ReadMemPackage.
// Note: a package does not support subfolders.
//
// NOTE: in the future, a MemPackage may represent updates/additional-files for
// an existing package.
type MemPackage struct {
	Name  string     `json:"name" yaml:"name"`           // package name as declared by `package`
	Path  string     `json:"path" yaml:"path"`           // import path
	Files []*MemFile `json:"files" yaml:"files"`         // plain file system files.
	Type  any        `json:"type,omitempty" yaml:"type"` // (user defined) package type.
	Info  any        `json:"info,omitempty" yaml:"info"` // (user defined) extra information.
}

// Package Name must be lower_case, can have digits & underscores.
// Package Path must be "a.valid.url/path" or a "simple/path".
// An empty MemPackager is invalid.
func (mpkg *MemPackage) ValidateBasic() error {
	// add assertion that MemPkg contains at least 1 file
	if len(mpkg.Files) <= 0 {
		return fmt.Errorf("no files found within package %q", mpkg.Name)
	}
	if len(mpkg.Name) > pkgNameLimit {
		return fmt.Errorf("name length %d exceeds limit %d", len(mpkg.Name), pkgNameLimit)
	}
	if len(mpkg.Path) > pkgPathLimit {
		return fmt.Errorf("path length %d exceeds limit %d", len(mpkg.Path), pkgPathLimit)
	}
	if !rePkgName.MatchString(mpkg.Name) {
		return fmt.Errorf("invalid package name %q", mpkg.Name)
	}
	if true && // none of these match...
		!rePkgPathURL.MatchString(mpkg.Path) &&
		!rePkgPathStd.MatchString(mpkg.Path) {
		return fmt.Errorf("invalid package path %q", mpkg.Path)
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

	// unique filenames
	if err := mpkg.Uniq(); err != nil {
		return err
	}

	// validate files
	for _, mfile := range mpkg.Files {
		if err := mfile.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid file in package: %w", err)
		}
		if !reFileName.MatchString(mfile.Name) {
			return fmt.Errorf("invalid file name %q, failed to match %q", mfile.Name, reFileName)
		}
	}
	return nil
}

// Returns an error if lowercase(file.Name) are not unique.
func (mpkg *MemPackage) Uniq() error {
	uniq := make(map[string]struct{}, len(mpkg.Files))
	for _, mfile := range mpkg.Files {
		lname := strings.ToLower(mfile.Name)
		if _, exists := uniq[lname]; exists {
			return fmt.Errorf("duplicate file name %q", lname)
		}
		uniq[lname] = struct{}{}
	}
	return nil
}

// Sort files; a MemPackage with unordered files is in valid.
func (mpkg *MemPackage) Sort() {
	slices.SortFunc(mpkg.Files, func(a, b *MemFile) int {
		return strings.Compare(a.Name, b.Name)
	})
}

// Return the named file or none if it doesn't exist.
func (mpkg *MemPackage) GetFile(name string) *MemFile {
	for _, mfile := range mpkg.Files {
		if mfile.Name == name {
			return mfile
		}
	}
	return nil
}

// Adds a file to the package without validation.
func (mpkg *MemPackage) AddFile(mfile *MemFile) {
	mpkg.Files = append(mpkg.Files, mfile)
}

// Creates a new MemFile and adds without validation.
func (mpkg *MemPackage) NewFile(name string, body string) (mfile *MemFile) {
	mfile = &MemFile{
		Name: name,
		Body: body,
	}
	mpkg.AddFile(mfile)
	return
}

// Writes to existing file or creates a new one.
func (mpkg *MemPackage) SetFile(name string, body string) *MemFile {
	for _, mfile := range mpkg.Files {
		if mfile.Name == name {
			mfile.Body = body
			return mfile
		}
	}
	return mpkg.NewFile(name, body)
}

// Removes an existing file and returns it or nil.
func (mpkg *MemPackage) DeleteFile(name string) *MemFile {
	for i, mfile := range mpkg.Files {
		if mfile.Name == name {
			mpkg.Files = append(mpkg.Files[:i], mpkg.Files[i+1:]...)
			return mfile
		}
	}
	return nil
}

// Returns true if it has no files.
func (mpkg *MemPackage) IsEmpty() bool {
	return mpkg.IsEmptyOf(".gno")
}

// Returns true if it has no files ending in `xtn`.  xtn should start with a
// dot to check extensions, but need not start with one, e.g. to test for
// _test.gno.
func (mpkg *MemPackage) IsEmptyOf(xtn string) bool {
	for _, mfile := range mpkg.Files {
		if strings.HasSuffix(mfile.Name, xtn) {
			return false
		}
	}
	return true
}

// Returns true if zero.
func (mpkg *MemPackage) IsZero() bool {
	return mpkg.Name == "" && len(mpkg.Files) == 0
}

// Write all files into dir.
func (mpkg *MemPackage) WriteTo(dir string) error {
	// fmt.Printf("writing mempackage to %q:\n", dir)
	for _, mfile := range mpkg.Files {
		// fmt.Printf(" - %s (%d bytes)\n", mfile.Name, len(mfile.Body))
		fpath := filepath.Join(dir, mfile.Name)
		err := os.WriteFile(fpath, []byte(mfile.Body), 0o644)
		if err != nil {
			return err
		}
	}
	return nil
}

// Print all files to stdout.
func (mpkg *MemPackage) Print() error {
	fmt.Printf("MemPackage[%q %s]:\n", mpkg.Path, mpkg.Name)
	for _, mfile := range mpkg.Files {
		mfile.Print()
	}
	return nil
}

// Return a list of all file names.
func (mpkg *MemPackage) FileNames() (fnames []string) {
	for _, mfile := range mpkg.Files {
		fnames = append(fnames, mfile.Name)
	}
	return
}

// Splits a path into the dir and filename.
func SplitFilepath(fpath string) (dir string, filename string) {
	dir, filename = path.Split(fpath)
	if dir == "" {
		// assume that filename is actually a directory
		return filename, ""
	}

	var (
		isFileWithExtension = strings.Contains(filename, ".")
		isSpecialFile       = filename == "LICENSE" || filename == "README"
		noFileSpecified     = filename == ""
	)
	if isFileWithExtension || isSpecialFile || noFileSpecified {
		dir = strings.TrimRight(dir, "/") // gno.land/r/path//a.gno -> dir=gno.land/r/path filename=a.gno
		return
	}

	return dir + filename, ""
}
