package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnoutil"
	"github.com/gnolang/gno/gnovm/tests"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type runCfg struct {
	verbose bool
	rootDir string
	expr    string
}

func newRunCmd(io *commands.IO) *commands.Command {
	cfg := &runCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "run",
			ShortUsage: "run [flags] <file|pkg> [<file|pkg>...]",
			ShortHelp:  "Runs the specified gno files or packages",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRun(cfg, args, io)
		},
	)
}

func (c *runCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"verbose",
		false,
		"verbose output when running",
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gnodev tries to guess it)",
	)

	fs.StringVar(
		&c.expr,
		"expr",
		"main()",
		"value of expression to evaluate. Defaults to executing function main() with no args",
	)
}

func execRun(cfg *runCfg, args []string, io *commands.IO) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	if cfg.rootDir == "" {
		cfg.rootDir = gnoutil.DefaultRootDir()
	}

	stdin := io.In
	stdout := io.Out
	stderr := io.Err

	// init store and machine
	testStore := tests.TestStore(cfg.rootDir,
		"", stdin, stdout, stderr,
		tests.ImportModeStdlibsPreferred)
	if cfg.verbose {
		testStore.SetLogStoreOps(true)
	}

	if len(args) == 0 {
		args = []string{"."}
	}

	// read files
	files, err := parseFiles(args)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errors.New("no files to run")
	}

	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: string(files[0].PkgName),
		Output:  stdout,
		Store:   testStore,
	})

	defer m.Release()

	// run files
	m.RunFiles(files...)
	runExpr(m, cfg.expr)

	return nil
}

func parseFiles(fnames []string) ([]*gno.FileNode, error) {
	gnoFnames, err := gnoutil.Match(fnames, gnoutil.MatchFiles("!*_test.gno", "!*_filetest.gno"))
	if err != nil {
		return nil, err
	}

	files := make([]*gno.FileNode, 0, len(gnoFnames))
	for _, fname := range gnoFnames {
		files = append(files, gno.MustReadFile(fname))
	}
	return files, nil
}

func runExpr(m *gno.Machine, expr string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic running expression %s: %v\n%s\n",
				expr, r, m.String())
			panic(r)
		}
	}()
	ex, err := gno.ParseExpr(expr)
	if err != nil {
		panic(fmt.Errorf("could not parse: %w", err))
	}
	m.Eval(ex)
}
