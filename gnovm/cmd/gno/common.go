package main

import (
	"fmt"
	"go/scanner"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
)

type gnoCode string

const (
	gnoUnknownError    gnoCode = "gnoUnknownError"
	gnoReadError               = "gnoReadError"
	gnoImportError             = "gnoImportError"
	gnoGnoModError             = "gnoGnoModError"
	gnoPreprocessError         = "gnoPreprocessError"
	gnoParserError             = "gnoParserError"
	gnoTypeCheckError          = "gnoTypeCheckError"

	// TODO: add new gno codes here.
)

type gnoIssue struct {
	Code       gnoCode
	Msg        string
	Confidence float64 // 1 is 100%
	Location   string  // file:line, or equivalent
	// TODO: consider writing fix suggestions
}

func (i gnoIssue) String() string {
	// TODO: consider crafting a doc URL based on Code.
	return fmt.Sprintf("%s: %s (code=%s)", i.Location, i.Msg, i.Code)
}

// Gno parses and sorts mpkg files into the following filesets:
//   - fset: all normal and _test.go files in package excluding `package xxx_test`
//     integration *_test.gno files.
//   - _tests: `package xxx_test` integration *_test.gno files, each in their
//     own file set.
//   - ftests: *_filetest.gno file tests, each in their own file set.
func sourceAndTestFileset(mpkg *std.MemPackage) (
	all, fset *gno.FileSet, _tests, ftests []*gno.FileSet,
) {
	all = &gno.FileSet{}
	fset = &gno.FileSet{}
	for _, mfile := range mpkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // Skip non-GNO files
		}

		n := gno.MustParseFile(mfile.Name, mfile.Body)
		if n == nil {
			continue // Skip empty files
		}
		all.AddFiles(n)
		if string(n.PkgName) == string(mpkg.Name)+"_test" {
			// A xxx_file integration test is a package of its own.
			fset := &gno.FileSet{}
			fset.AddFiles(n)
			_tests = append(_tests, fset)
		} else if strings.HasSuffix(mfile.Name, "_filetest.gno") {
			// A _filetest.gno is a package of its own.
			fset := &gno.FileSet{}
			fset.AddFiles(n)
			ftests = append(ftests, fset)
		} else {
			// All normal package files and,
			// _test.gno files that aren't xxx_test.
			fset.AddFiles(n)
		}
	}
	return
}

func guessSourcePath(pkgPath, fname string) string {
	if info, err := os.Stat(pkgPath); !os.IsNotExist(err) && !info.IsDir() {
		pkgPath = filepath.Dir(pkgPath)
	}

	fnameJoin := filepath.Join(pkgPath, fname)
	if _, err := os.Stat(fnameJoin); !os.IsNotExist(err) {
		return filepath.Clean(fnameJoin)
	}

	if _, err := os.Stat(fname); !os.IsNotExist(err) {
		return filepath.Clean(fname)
	}

	return filepath.Clean(pkgPath)
}

// reParseRecover is a regex designed to parse error details from a string.
// It extracts the file location, line number, and error message from a
// formatted error string.
// XXX: Ideally, error handling should encapsulate location details within a
// dedicated error type.
var reParseRecover = regexp.MustCompile(`^([^:]+)((?::(?:\d+)){1,2}):? *(.*)$`)

func printError(w io.WriteCloser, dir, pkgPath string, err error) {
	switch err := err.(type) {
	case *gno.PreprocessError:
		err2 := err.Unwrap()
		fmt.Fprintln(w, issueFromError(
			dir, pkgPath, err2, gnoPreprocessError).String())
	case gno.ImportError:
		fmt.Fprintln(w, issueFromError(
			dir, pkgPath, err, gnoImportError).String())
	case scanner.ErrorList:
		for _, err := range err {
			fmt.Fprintln(w, issueFromError(
				dir,
				pkgPath,
				err,
				gnoParserError,
			).String())
		}
	default: // error type
		errors := multierr.Errors(err)
		if len(errors) == 1 {
			fmt.Fprintln(w, issueFromError(
				dir,
				pkgPath,
				err,
				gnoUnknownError,
			).String())
			return
		}
		for _, err := range errors {
			printError(w, dir, pkgPath, err)
		}
	}
}

func catchPanic(dir, pkgPath string, stderr io.WriteCloser, action func()) (didPanic bool) {
	defer func() {
		// Errors catched here mostly come from:
		// gnovm/pkg/gnolang/preprocess.go
		r := recover()
		if r == nil {
			return
		}
		didPanic = true
		if err, ok := r.(error); ok {
			printError(stderr, dir, pkgPath, err)
		} else {
			panic(r)
		}
	}()

	action()
	return
}

func issueFromError(dir, pkgPath string, err error, code gnoCode) gnoIssue {
	var issue gnoIssue
	issue.Confidence = 1
	issue.Code = code

	parsedError := strings.TrimSpace(err.Error())
	parsedError = replaceWithDirPath(parsedError, pkgPath, dir)
	parsedError = strings.TrimPrefix(parsedError, pkgPath+"/")

	matches := reParseRecover.FindStringSubmatch(parsedError)
	if len(matches) > 0 {
		sourcepath := guessSourcePath(pkgPath, matches[1])
		issue.Location = sourcepath + matches[2]
		issue.Msg = strings.TrimSpace(matches[3])
	} else {
		issue.Location = fmt.Sprintf("%s:0", filepath.Clean(pkgPath))
		issue.Msg = err.Error()
	}
	return issue
}

func replaceWithDirPath(s, pkgPath, dir string) string {
	if strings.HasPrefix(s, pkgPath) {
		return filepath.Clean(dir + s[len(pkgPath):])
	}
	return s
}
