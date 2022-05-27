package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	RootDir string `flag:"root-dir" help:"clone location of github.com/gnolang/gno (gnodev tries to guess it)"`
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

	// guess opts.RootDir
	if opts.RootDir == "" {
		cmd := exec.Command("go", "list", "-m", "-mod=mod", "-f", "{{.Dir}}", "github.com/gnolang/gno")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal("can't guess --root-dir, please fill it manually.")
		}
		rootDir := strings.TrimSpace(string(out))
		opts.RootDir = rootDir
	}

	pkgPaths, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	errCount := 0
	for _, pkgPath := range pkgPaths {
		testFiles := []string{}
		fileSystem := os.DirFS(pkgPath)
		fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Fatal(err)
			}
			if d.IsDir() {
				return nil
			}
			/*
				if strings.HasSuffix(path, "_test.gno") {
					panic("*_test.gno files are not yet supported by gnodev")
				}
			*/
			if strings.HasSuffix(path, "_filetest.gno") {
				testFiles = append(testFiles, path)
			}
			return nil
		})
		sort.Strings(testFiles)

		if len(testFiles) > 0 {
			startedAt := time.Now()
			err = gnoTestPkg(pkgPath, testFiles, opts)
			duration := time.Since(startedAt)

			if err != nil {
				err = fmt.Errorf("%s: test pkg: %w", pkgPath, err)
				cmd.ErrPrintfln("FAIL")
				cmd.ErrPrintfln("FAIL    %s \t%v", pkgPath, duration)
				cmd.ErrPrintfln("FAIL")
				errCount++
			} else {
				cmd.ErrPrintfln("ok      %s \t%v", pkgPath, duration)
			}
		} else {
			cmd.ErrPrintfln("?       %s \t[no test files]", pkgPath)
		}

		// testing with *_test.gno
		{
			fs, err := filepath.Glob(filepath.Join(pkgPath, "*_test.gno"))
			if err != nil {
				log.Fatal(err)
			}
			if len(fs) <= 0 {
				cmd.ErrPrintfln("?       %s \t[no test files]", pkgPath)
				continue
			}

			testStore := newTestStore(opts.RootDir, os.Stdin, os.Stdout, os.Stderr)
			startedAt := time.Now()
			ok := runTest(testStore, pkgPath, opts.Verbose)
			duration := time.Since(startedAt)
			if !ok {
				cmd.ErrPrintfln("FAIL    %s \t%v", pkgPath, duration)
			} else {
				cmd.Printfln("ok      %s \t%v", pkgPath, duration)
			}
		}
	}
	if errCount > 0 {
		cmd.ErrPrintfln("FAIL")
		return fmt.Errorf("FAIL: %d go test errors", errCount)
	}

	return nil
}

func gnoTestPkg(pkgPath string, testFiles []string, opts testOptions) error {
	verbose := opts.Verbose
	rootDir := opts.RootDir
	// FIXME support test-based, examples, huband full packages
	// FIXME update Makefile and CI

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
