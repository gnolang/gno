// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/rogpeppe/go-internal/goproxytest"
	"github.com/rogpeppe/go-internal/gotooltest"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/rogpeppe/go-internal/txtar"
)

const (
	// goModProxyDir is the special subdirectory in a txtar script's supporting files
	// within which we expect to find github.com/rogpeppe/go-internal/goproxytest
	// directories.
	goModProxyDir = ".gomodproxy"
)

type envVarsFlag struct {
	vals []string
}

func (e *envVarsFlag) String() string {
	return fmt.Sprintf("%v", e.vals)
}

func (e *envVarsFlag) Set(v string) error {
	e.vals = append(e.vals, v)
	return nil
}

func main() {
	os.Exit(main1())
}

func main1() int {
	switch err := mainerr(); err {
	case nil:
		return 0
	case flag.ErrHelp:
		return 2
	default:
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
}

func mainerr() (retErr error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.Usage = func() {
		mainUsage(os.Stderr)
	}
	var envVars envVarsFlag
	fUpdate := fs.Bool("u", false, "update archive file if a cmp fails")
	fWork := fs.Bool("work", false, "print temporary work directory and do not remove when done")
	fContinue := fs.Bool("continue", false, "continue running the script if an error occurs")
	fVerbose := fs.Bool("v", false, "run tests verbosely")
	fs.Var(&envVars, "e", "pass through environment variable to script (can appear multiple times)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	td, err := ioutil.TempDir("", "testscript")
	if err != nil {
		return fmt.Errorf("unable to create temp dir: %v", err)
	}
	if *fWork {
		fmt.Fprintf(os.Stderr, "temporary work directory: %v\n", td)
	} else {
		defer os.RemoveAll(td)
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	// If we are only reading from stdin, -u cannot be specified. It seems a bit
	// bizarre to invoke testscript with '-' and a regular file, but hey. In
	// that case the -u flag will only apply to the regular file and we assume
	// the user knows it.
	onlyReadFromStdin := true
	for _, f := range files {
		if f != "-" {
			onlyReadFromStdin = false
		}
	}
	if onlyReadFromStdin && *fUpdate {
		return fmt.Errorf("cannot use -u when reading from stdin")
	}

	tr := testRunner{
		update:          *fUpdate,
		continueOnError: *fContinue,
		verbose:         *fVerbose,
		env:             envVars.vals,
		testWork:        *fWork,
	}

	dirNames := make(map[string]int)
	for _, filename := range files {
		// TODO make running files concurrent by default? If we do, note we'll need to do
		// something smarter with the runner stdout and stderr below

		// Derive a name for the directory from the basename of file, making
		// uniq by adding a numeric suffix in the case we otherwise end
		// up with multiple files with the same basename
		dirName := filepath.Base(filename)
		count := dirNames[dirName]
		dirNames[dirName] = count + 1
		if count != 0 {
			dirName = fmt.Sprintf("%s%d", dirName, count)
		}

		runDir := filepath.Join(td, dirName)
		if err := os.Mkdir(runDir, 0o777); err != nil {
			return fmt.Errorf("failed to create a run directory within %v for %v: %v", td, renderFilename(filename), err)
		}
		if err := tr.run(runDir, filename); err != nil {
			return err
		}
	}

	return nil
}

type testRunner struct {
	// update denotes that the source testscript archive filename should be
	// updated in the case of any cmp failures.
	update bool

	// continueOnError indicates that T.FailNow should not panic, allowing the
	// test script to continue running. Note that T is still marked as failed.
	continueOnError bool

	// verbose indicates the running of the script should be noisy.
	verbose bool

	// env is the environment that should be set on top of the base
	// testscript-defined minimal environment.
	env []string

	// testWork indicates whether or not temporary working directory trees
	// should be left behind. Corresponds exactly to the
	// testscript.Params.TestWork field.
	testWork bool
}

// run runs the testscript archive located at the path filename, within the
// working directory runDir. filename could be "-" in the case of stdin
func (tr *testRunner) run(runDir, filename string) error {
	var ar *txtar.Archive
	var err error

	mods := filepath.Join(runDir, goModProxyDir)

	if err := os.MkdirAll(mods, 0o777); err != nil {
		return fmt.Errorf("failed to create goModProxy dir: %v", err)
	}

	if filename == "-" {
		byts, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %v", err)
		}
		ar = txtar.Parse(byts)
	} else {
		ar, err = txtar.ParseFile(filename)
	}

	if err != nil {
		return fmt.Errorf("failed to txtar parse %v: %v", renderFilename(filename), err)
	}

	var script, gomodProxy txtar.Archive
	script.Comment = ar.Comment

	for _, f := range ar.Files {
		fp := filepath.Clean(filepath.FromSlash(f.Name))
		parts := strings.Split(fp, string(os.PathSeparator))

		if len(parts) > 1 && parts[0] == goModProxyDir {
			gomodProxy.Files = append(gomodProxy.Files, f)
		} else {
			script.Files = append(script.Files, f)
		}
	}

	if txtar.Write(&gomodProxy, runDir); err != nil {
		return fmt.Errorf("failed to write .gomodproxy files: %v", err)
	}

	scriptFile := filepath.Join(runDir, "script.txtar")

	if err := ioutil.WriteFile(scriptFile, txtar.Format(&script), 0o666); err != nil {
		return fmt.Errorf("failed to write script for %v: %v", renderFilename(filename), err)
	}

	p := testscript.Params{
		Dir:             runDir,
		UpdateScripts:   tr.update,
		ContinueOnError: tr.continueOnError,
	}

	if _, err := exec.LookPath("go"); err == nil {
		if err := gotooltest.Setup(&p); err != nil {
			return fmt.Errorf("failed to setup go tool for %v run: %v", renderFilename(filename), err)
		}
	}

	addSetup := func(f func(env *testscript.Env) error) {
		origSetup := p.Setup
		p.Setup = func(env *testscript.Env) error {
			if origSetup != nil {
				if err := origSetup(env); err != nil {
					return err
				}
			}
			return f(env)
		}
	}

	if tr.testWork {
		addSetup(func(env *testscript.Env) error {
			fmt.Fprintf(os.Stderr, "temporary work directory for %s: %s\n", renderFilename(filename), env.WorkDir)
			return nil
		})
	}

	if len(gomodProxy.Files) > 0 {
		srv, err := goproxytest.NewServer(mods, "")
		if err != nil {
			return fmt.Errorf("cannot start proxy for %v: %v", renderFilename(filename), err)
		}
		defer srv.Close()

		addSetup(func(env *testscript.Env) error {
			// Add GOPROXY after calling the original setup
			// so that it overrides any GOPROXY set there.
			env.Vars = append(env.Vars,
				"GOPROXY="+srv.URL,
				"GONOSUMDB=*",
			)
			return nil
		})
	}

	if len(tr.env) > 0 {
		addSetup(func(env *testscript.Env) error {
			for _, v := range tr.env {
				varName := v
				if i := strings.Index(v, "="); i >= 0 {
					varName = v[:i]
				} else {
					v = fmt.Sprintf("%s=%s", v, os.Getenv(v))
				}
				switch varName {
				case "":
					return fmt.Errorf("invalid variable name %q", varName)
				case "WORK":
					return fmt.Errorf("cannot override WORK variable")
				}
				env.Vars = append(env.Vars, v)
			}
			return nil
		})
	}

	r := &runT{
		verbose: tr.verbose,
	}

	func() {
		defer func() {
			switch recover() {
			case nil, skipRun:
			case failedRun:
				err = failedRun
			default:
				panic(fmt.Errorf("unexpected panic: %v [%T]", err, err))
			}
		}()
		testscript.RunT(r, p)

		// When continueOnError is true, FailNow does not call panic(failedRun).
		// We still want err to be set, as the script resulted in a failure.
		if r.Failed() {
			err = failedRun
		}
	}()

	if err != nil {
		return fmt.Errorf("error running %v in %v\n", renderFilename(filename), runDir)
	}

	if tr.update && filename != "-" {
		// Parse the (potentially) updated scriptFile as an archive, then merge
		// with the original archive, retaining order.  Then write the archive
		// back to the source file
		source, err := ioutil.ReadFile(scriptFile)
		if err != nil {
			return fmt.Errorf("failed to read from script file %v for -update: %v", scriptFile, err)
		}
		updatedAr := txtar.Parse(source)
		updatedFiles := make(map[string]txtar.File)
		for _, f := range updatedAr.Files {
			updatedFiles[f.Name] = f
		}
		for i, f := range ar.Files {
			if newF, ok := updatedFiles[f.Name]; ok {
				ar.Files[i] = newF
			}
		}
		if err := ioutil.WriteFile(filename, txtar.Format(ar), 0o666); err != nil {
			return fmt.Errorf("failed to write script back to %v for -update: %v", renderFilename(filename), err)
		}
	}

	return nil
}

var (
	failedRun = errors.New("failed run")
	skipRun   = errors.New("skip")
)

// renderFilename renders filename in error messages, taking into account
// the filename could be the special "-" (stdin)
func renderFilename(filename string) string {
	if filename == "-" {
		return "<stdin>"
	}
	return filename
}

// runT implements testscript.T and is used in the call to testscript.Run
type runT struct {
	verbose bool
	failed  int32
}

func (r *runT) Skip(is ...interface{}) {
	panic(skipRun)
}

func (r *runT) Fatal(is ...interface{}) {
	r.Log(is...)
	r.FailNow()
}

func (r *runT) Parallel() {
	// No-op for now; we are currently only running a single script in a
	// testscript instance.
}

func (r *runT) Log(is ...interface{}) {
	fmt.Print(is...)
}

func (r *runT) FailNow() {
	atomic.StoreInt32(&r.failed, 1)
	panic(failedRun)
}

func (r *runT) Failed() bool {
	return atomic.LoadInt32(&r.failed) != 0
}

func (r *runT) Run(n string, f func(t testscript.T)) {
	// For now we we don't top/tail the run of a subtest. We are currently only
	// running a single script in a testscript instance, which means that we
	// will only have a single subtest.
	f(r)
}

func (r *runT) Verbose() bool {
	return r.verbose
}
