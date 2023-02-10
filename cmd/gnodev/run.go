package main

import (
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/tests"
)

type runOptions struct {
	Verbose bool   `flag:"verbose" help:"verbose"`
	RootDir string `flag:"root-dir" help:"clone location of github.com/gnolang/gno (gnodev tries to guess it)"`
	// Timeout time.Duration `flag:"timeout" help:"max execution time"`
	// VM Options
	// UseNativeLibs bool // experimental, but could be useful for advanced developer needs
}

var defaultRunOptions = runOptions{
	Verbose: false,
	RootDir: "",
}

func runApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(runOptions)

	if len(args) == 0 {
		cmd.ErrPrintfln("Usage: run [flags] file.gno [file2.gno...]")

		return errors.New("invalid args")
	}

	if opts.RootDir == "" {
		opts.RootDir = guessRootDir()
	}

	fnames := args

	return runRun(cmd, opts.RootDir, opts.Verbose, fnames)
}

func runRun(cmd *command.Command, rootDir string, verbose bool, fnames []string) error {
	stdin := cmd.In
	stdout := cmd.Out
	stderr := cmd.Err

	// init store and machine
	testStore := tests.TestStore(rootDir,
		"", stdin, stdout, stderr,
		tests.ImportModeStdlibsPreferred)
	if verbose {
		testStore.SetLogStoreOps(true)
	}

	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "main",
		Output:  stdout,
		Store:   testStore,
	})

	// read files
	files := make([]*gno.FileNode, len(fnames))
	for i, fname := range fnames {
		files[i] = gno.MustReadFile(fname)
	}

	// run files
	m.RunFiles(files...)
	m.RunMain()

	return nil
}
