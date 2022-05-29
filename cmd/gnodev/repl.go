package main

import (
	"fmt"
	"os"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/tests"
)

type replOptions struct {
	Verbose bool   `flag:"verbose" help:"verbose"`
	RootDir string `flag:"root-dir" help:"clone location of github.com/gnolang/gno (gnodev tries to guess it)"`
	// Run string `flag:"run" help:"test name filtering pattern"`
	// Timeout time.Duration `flag:"timeout" help:"max execution time"`
	// VM Options
	// A flag about if we should download the production realms
	// UseNativeLibs bool // experimental, but could be useful for advanced developer needs
	// AutoImport bool
	// ImportPkgs...
}

var DefaultReplOptions = replOptions{
	Verbose: false,
	RootDir: "",
}

func replApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(replOptions)
	if len(args) > 0 {
		cmd.ErrPrintfln("Usage: repl [flags]")
		return errors.New("invalid args")
	}

	if opts.RootDir == "" {
		opts.RootDir = guessRootDir()
	}

	return runRepl(opts.RootDir, opts.Verbose)
}

func runRepl(rootDir string, verbose bool) error {
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr
	useNativeLibs := false

	testStore := tests.TestStore(rootDir, "", stdin, stdout, stderr, useNativeLibs)
	if verbose {
		testStore.SetLogStoreOps(true)
	}

	m := tests.TestMachine(testStore, stdout, "main")

	input := `package main
func main() {
	println("hello")
}
`
	n := gno.MustParseFile("main.go", input)
	m.RunFiles(n)
	m.RunMain()

	fmt.Println(m)

	return nil
}
