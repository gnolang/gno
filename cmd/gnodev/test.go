package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

type testOptions struct {
	Verbose bool `flag:"verbose" help:"verbose"`
	// Run string `flag:"run" help:"test name filtering pattern"`
	// Timeout time.Duration `flag:"timeout" help:"max execution time"`
	// VM Options
	// A flag about if we should download the production realms
}

var DefaultTestOptions = testOptions{
	Verbose: false,
}

func testApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(testOptions)
	if len(args) < 1 {
		cmd.ErrPrintfln("Usage: test [test flags] [packages]")
		return errors.New("invalid args")
	}

	pkgs, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	errCount := 0
	for _, pkg := range pkgs {
		startedAt := time.Now()
		err = gnoTestPkg(pkg, opts)
		if err != nil {
			err = fmt.Errorf("%s: test pkg: %w", pkg, err)
			cmd.ErrPrintfln("%s", err)
			errCount++
			continue
		}
		duration := time.Since(startedAt)
		cmd.ErrPrintfln("ok      %s (%v)", pkg, duration)
	}
	if errCount > 0 {
		return fmt.Errorf("FAIL: %d go test errors", errCount)
	}

	return nil
}

func gnoTestPkg(path string, opts testOptions) error {
	verbose := opts.Verbose
	// FIXME support test-based, examples, and full packages
	// FIXME update Makefile and CI

	// * put in mempkg, including deps
	// * setup vm
	// * iterate over files

	testNames := []string{"TestFoo", "TestBar"}
	for _, testName := range testNames {
		startedAt := time.Now()
		if verbose {
			fmt.Fprintf(os.Stderr, "=== RUN   %s\n", testName)
		}
		if verbose {
			duration := time.Since(startedAt)
			fmt.Fprintf(os.Stderr, "--- PASS: %s (%v)\n", testName, duration)
		}
	}

	return nil
}
