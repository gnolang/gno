package main

import (
	"bytes"
	"fmt"
	"go/scanner"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
)

type gnoCode string

const (
	gnoUnknownError    gnoCode = "gnoUnknownError"
	gnoReadError       gnoCode = "gnoReadError"
	gnoImportError     gnoCode = "gnoImportError"
	gnoGnoModError     gnoCode = "gnoGnoModError"
	gnoPreprocessError gnoCode = "gnoPreprocessError"
	gnoParserError     gnoCode = "gnoParserError"
	gnoTypeCheckError  gnoCode = "gnoTypeCheckError"

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
// Args:
//   - onlyFiletests: true if all files are filetests. relaxed.
//     used to transpile test/files/*.gno (no _filetest.gno suffix).
//
// Results:
//   - fset: all normal and _test.go files in package excluding `package xxx_test`
//     integration *_test.gno files.
//   - _tests: `package xxx_test` integration *_test.gno files.
//   - ftests: *_filetest.gno file tests, each in their own file set.
func sourceAndTestFileset(mpkg *std.MemPackage, onlyFiletests bool) (
	all, fset *gno.FileSet, _tests *gno.FileSet, ftests []*gno.FileSet,
) {
	all = &gno.FileSet{}
	fset = &gno.FileSet{}
	_tests = &gno.FileSet{}
	for _, mfile := range mpkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // Skip non-GNO files
		}

		n := gno.MustParseFile(mfile.Name, mfile.Body)
		if n == nil {
			continue // Skip empty files
		}
		all.AddFiles(n)
		if strings.HasSuffix(mfile.Name, "_filetest.gno") || onlyFiletests {
			// A _filetest.gno is a package of its own.
			ftset := &gno.FileSet{}
			ftset.AddFiles(n)
			ftests = append(ftests, ftset)
		} else if strings.HasSuffix(mfile.Name, "_test.gno") &&
			strings.HasSuffix(string(n.PkgName), "_test") {
			// A xxx_file integration test is a package of its own.
			_tests.AddFiles(n)
		} else {
			// All normal package files and,
			// _test.gno files that aren't xxx_test.
			fset.AddFiles(n)
		}
	}
	return
}

func parsePkgPathDirective(body string, defaultPkgPath string) (string, error) {
	dirs, err := test.ParseDirectives(bytes.NewReader([]byte(body)))
	if err != nil {
		return "", fmt.Errorf("error parsing directives: %w", err)
	}
	return dirs.FirstDefault(test.DirectivePkgPath, defaultPkgPath), nil
}

type processedFileSet struct {
	pn   *gno.PackageNode
	fset *gno.FileSet
}

type processedPackage struct {
	dir    string             // dirctory
	mpkg   *std.MemPackage    // includes all files
	normal processedFileSet   // includes all prod (and some *_test.gno) files
	_tests processedFileSet   // includes all xxx_test *_test.gno integration files
	ftests []processedFileSet // includes all *_filetest.gno filetest files
}

func (ppkg *processedPackage) AddNormal(pn *gno.PackageNode, fset *gno.FileSet) {
	if ppkg.normal != (processedFileSet{}) {
		panic("normal processed fileset already set")
	}
	ppkg.normal = processedFileSet{pn, fset}
}

func (ppkg *processedPackage) AddUnderscoreTests(pn *gno.PackageNode, fset *gno.FileSet) {
	if ppkg._tests != (processedFileSet{}) {
		panic("_test processed fileset already set")
	}
	ppkg._tests = processedFileSet{pn, fset}
}

func (ppkg *processedPackage) AddFileTest(pn *gno.PackageNode, fset *gno.FileSet) {
	if len(fset.Files) != 1 {
		panic("filetests must have filesets of length 1")
	}
	fname := fset.Files[0].FileName
	/* NOTE: filetests in tests/files do not end with _filetest.gno.
	if !strings.HasSuffix(string(fname), "_filetest.gno") {
		panic(fmt.Sprintf("expected *_filetest.gno but got %q", fname))
	}
	*/
	for _, ftest := range ppkg.ftests {
		if ftest.fset.Files[0].FileName == fname {
			panic(fmt.Sprintf("fileetest with name %q already exists", fname))
		}
	}
	ppkg.ftests = append(ppkg.ftests, processedFileSet{pn, fset})
}

func (ppkg *processedPackage) GetFileTest(fname string) processedFileSet {
	if !strings.HasSuffix(fname, "_filetest.gno") {
		panic(fmt.Sprintf("expected *_filetest.gno but got %q", fname))
	}
	for _, ftest := range ppkg.ftests {
		if ftest.fset.Files[0].FileName == fname {
			return ftest
		}
	}
	panic(fmt.Sprintf("processedFileSet for filetest %q not found", fname))
}

func printError(w io.WriteCloser, dir, pkgPath string, err error) {
	switch err := err.(type) {
	case *gno.PreprocessError:
		err2 := err.Unwrap()
		// XXX probably no need for guessing, replace with exact issue.
		fmt.Fprintln(w, guessIssueFromError(
			dir, pkgPath, err2, gnoPreprocessError).String())
	case gno.ImportError:
		// NOTE: gnovm/pkg/test.LoadImport will return a
		// ImportNotFoundError with format "<loc>: unknown import path:
		// <path>", while gimp.ImportFrom() doesn't know <loc> so
		// returns a ImportNotFoundError with format "unknown import
		// path: <path>"; but Go .Check ends up returning a types.Error
		// instead, as seen in the hack in the next clause.  So
		// test.LoadImport needs this and guessing isn't needed.
		fmt.Fprintln(w, gnoIssue{
			Code:       gnoImportError,
			Msg:        err.GetMsg(),
			Confidence: 1,
			Location:   err.GetLocation(),
		})
	case types.Error:
		loc := err.Fset.Position(err.Pos).String()
		loc = guessFilePathLoc(loc, pkgPath, dir)
		code := gnoTypeCheckError
		if strings.Contains(err.Msg, "(unknown import path \"") {
			// NOTE: This is a bit of a hack.
			// See gimp.ImportFrom() comment on ImportNotFoundError
			// on why this is necessary, and how to make it less hacky.
			code = gnoImportError
		}
		fmt.Fprintln(w, gnoIssue{
			Code:       code,
			Msg:        err.Msg,
			Confidence: 1,
			Location:   loc,
		})
	case scanner.ErrorList:
		for _, err := range err {
			loc := err.Pos.String()
			loc = guessFilePathLoc(loc, pkgPath, dir)
			fmt.Fprintln(w, gnoIssue{
				Code:       gnoParserError,
				Msg:        err.Msg,
				Confidence: 1,
				Location:   loc,
			})
		}
	case scanner.Error:
		loc := err.Pos.String()
		loc = guessFilePathLoc(loc, pkgPath, dir)
		fmt.Fprintln(w, gnoIssue{
			Code:       gnoParserError,
			Msg:        err.Msg,
			Confidence: 1,
			Location:   loc,
		})
	default: // error type
		errors := multierr.Errors(err)
		if len(errors) == 1 {
			fmt.Fprintln(w, guessIssueFromError(
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
	// If this gets out of hand (e.g. with nested catchPanic with need for
	// selective catching) then pass in a bool instead.
	// See also pkg/test/imports.go.
	if os.Getenv("DEBUG_PANIC") == "1" {
		fmt.Println("DEBUG_PANIC=1 (will not recover)")
	} else {
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
	}

	action()
	return
}

func guessIssueFromError(dir, pkgPath string, err error, code gnoCode) gnoIssue {
	var issue gnoIssue
	issue.Confidence = 1
	issue.Code = code

	parsedError := strings.TrimSpace(err.Error())
	match := gno.Re_errorLine.Match(parsedError)
	if match == nil {
		issue.Location = fmt.Sprintf("%s:0", filepath.Clean(pkgPath))
		issue.Msg = err.Error()
	} else {
		errPath := match.Get("PATH")
		errLoc := match.Get("LOC")
		errMsg := match.Get("MSG")
		errPath = guessFilePathLoc(errPath, pkgPath, dir)
		errPath = filepath.Clean(errPath)
		issue.Location = errPath + ":" + errLoc
		issue.Msg = strings.TrimSpace(errMsg)
	}
	return issue
}

// Takes a location string `s` and tries to convert to a path based on `dir`.
// NOTE: s may not be in pkgPath (e.g. for type-check errors on imports).
// Do not make a transformation unless the answer is highly unlikely to be incorrect.
// Otherwise debugging may be painful. Better to return s as is.
func guessFilePathLoc(s, pkgPath, dir string) string {
	if !dirExists(dir) {
		panic(fmt.Sprintf("dir %q does not exist", dir))
	}
	s = filepath.Clean(s)
	pkgPath = filepath.Clean(pkgPath)
	dir = filepath.Clean(dir)
	// s already in dir.
	if strings.HasPrefix(s, dir) {
		return s
	}
	// s in pkgPath.
	if strings.HasPrefix(s, pkgPath+"/") {
		fname := s[len(pkgPath+"/"):]
		fpath := filepath.Join(dir, fname)
		return fpath
	}
	// "GNOROOT".
	if strings.HasSuffix(dir, pkgPath) {
		gnoRoot := dir[len(dir)-len(pkgPath):]
		// s is maybe <pkgPath>/<filename>
		if strings.Contains(s, "/") {
			fpath := gnoRoot + s
			if fileExists(fpath) {
				return fpath
			}
		}
	}
	// s is a filename.
	if !strings.Contains(s, "/") {
		fpath := filepath.Join(dir, s)
		if fileExists(fpath) {
			return fpath
		}
	}
	// dunno.
	return s
}

func dirExists(dir string) bool {
	info, err := os.Stat(dir)
	return !os.IsNotExist(err) && info.IsDir()
}

func fileExists(fpath string) bool {
	info, err := os.Stat(fpath)
	return !os.IsNotExist(err) && !info.IsDir()
}
