package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/tests"
	"go.uber.org/multierr"
)

type testOptions struct {
	Verbose bool   `flag:"verbose" help:"verbose"`
	RootDir string `flag:"root-dir" help:"github.com/gnolang/gno clone dir"`
	// Run string `flag:"run" help:"test name filtering pattern"`
	// Timeout time.Duration `flag:"timeout" help:"max execution time"`
	// VM Options
	// A flag about if we should download the production realms
}

var DefaultTestOptions = testOptions{
	Verbose: false,
	RootDir: "",
}

func testApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(testOptions)
	if len(args) < 1 {
		cmd.ErrPrintfln("Usage: test [test flags] [packages]")
		return errors.New("invalid args")
	}

	// FIXME: guess opts.RootDir by walking parent dirs.

	pkgDirs, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	errCount := 0
	for _, pkgDir := range pkgDirs {
		testFiles := []string{}
		hasPackageTests := false
		fileSystem := os.DirFS(pkgDir)
		fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Fatal(err)
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, "_test.gno") {
				hasPackageTests = true
			}
			if strings.HasSuffix(path, "_filetest.gno") {
				testFiles = append(testFiles, path)
			}
			return nil
		})
		// FIXME: perform tests per package instead of per test kind
		sort.Strings(testFiles)

		if len(testFiles) == 0 && !hasPackageTests {
			cmd.ErrPrintfln("?       %s \t[no test files]", pkgDir)
			continue
		}

		startedAt := time.Now()
		// run _test.gno tests
		if hasPackageTests {
			err = gnoTestPkg(pkgDir, opts)
			if err != nil {
				duration := time.Since(startedAt)
				err = fmt.Errorf("%s: test pkg: %w", pkgDir, err)
				cmd.ErrPrintfln("FAIL")
				cmd.ErrPrintfln("FAIL    %s \t%v", pkgDir, duration)
				cmd.ErrPrintfln("FAIL")
				errCount++
				continue
			}
		}

		// run _filetest.gno tests
		if len(testFiles) > 0 {
			err = gnoTestFiles(pkgDir, testFiles, opts)
			if err != nil {
				duration := time.Since(startedAt)
				err = fmt.Errorf("%s: test pkg: %w", pkgDir, err)
				cmd.ErrPrintfln("FAIL")
				cmd.ErrPrintfln("FAIL    %s \t%v", pkgDir, duration)
				cmd.ErrPrintfln("FAIL")
				errCount++
				continue
			}
		}

		duration := time.Since(startedAt)
		cmd.ErrPrintfln("ok      %s \t%v", pkgDir, duration)
	}
	if errCount > 0 {
		cmd.ErrPrintfln("FAIL")
		return fmt.Errorf("FAIL: %d go test errors", errCount)
	}

	return nil
}

func gnoTestFiles(pkgDir string, testFiles []string, opts testOptions) error {
	verbose := opts.Verbose
	rootDir := opts.RootDir

	var errs error
	for _, testFile := range testFiles {
		testName := "file/" + testFile
		startedAt := time.Now()
		if verbose {
			fmt.Fprintf(os.Stderr, "=== RUN   %s\n", testName)
		}

		var closer func() string
		if !verbose {
			var err error
			closer, err = captureStdoutAndStderr()
			if err != nil {
				panic(err)
			}
		}

		testFilePath := filepath.Join(pkgDir, testFile)
		err := tests.RunFileTest(rootDir, testFilePath, false, nil)
		duration := time.Since(startedAt)
		if err != nil {
			errs = multierr.Append(errs, err)
			fmt.Fprintf(os.Stderr, "--- FAIL: %s (%v)\n", testName, duration)
			if !verbose {
				stdouterr := closer()
				fmt.Fprintln(os.Stderr, stdouterr)
			}
			continue
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "--- PASS: %s (%v)\n", testName, duration)
		}
	}

	return errs
}

func gnoTestPkg(pkgDir string, opts testOptions) error {
	rootDir := opts.RootDir
	exampleDir := filepath.Join(rootDir, "examples")
	pkgPath, err := filepath.Rel(exampleDir, pkgDir)
	if err != nil {
		return fmt.Errorf("failed to guess package path")
	}

	err = tests.RunPackageTest(nil, pkgDir, pkgPath)
	fmt.Println(err)

	// BLAHBLAH ../stdlibs/bufio bufio

	/*
		verbose := opts.Verbose

		var errs error
		for _, testFile := range testFiles {
			testName := "file/" + testFile
			startedAt := time.Now()
			if verbose {
				fmt.Fprintf(os.Stderr, "=== RUN   %s\n", testName)
			}

			var closer func() string
			if !verbose {
				var err error
				closer, err = captureStdoutAndStderr()
				if err != nil {
					panic(err)
				}
			}

			testFilePath := filepath.Join(pkgPath, testFile)
			err := tests.RunFileTest(rootDir, testFilePath, false, nil)
			duration := time.Since(startedAt)
			if err != nil {
				errs = multierr.Append(errs, err)
				fmt.Fprintf(os.Stderr, "--- FAIL: %s (%v)\n", testName, duration)
				if !verbose {
					stdouterr := closer()
					fmt.Fprintln(os.Stderr, stdouterr)
				}
				continue
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "--- PASS: %s (%v)\n", testName, duration)
			}
		}

		return errs
	*/
	return nil
}

// CaptureStdoutAndStderr temporarily pipes os.Stdout and os.Stderr into a buffer.
// Imported from https://github.com/moul/u/blob/master/io.go.
func captureStdoutAndStderr() (func() string, error) {
	oldErr := os.Stderr
	oldOut := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	os.Stderr = w
	os.Stdout = w

	closer := func() string {
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stderr = oldErr
		os.Stdout = oldOut
		return string(out)
	}
	return closer, nil
}
