package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

type lintCfg struct {
	verbose       bool
	rootDir       string
	setExitStatus int
	// min_confidence: minimum confidence of a problem to pirnt it (default 0.8)
	// auto-fix: apply suggested fixes automatically.
}

func newLintCmd(io *commands.IO) *commands.Command {
	cfg := &lintCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "lint",
			ShortUsage: "lint [flags] <package> [<package>...]",
			ShortHelp:  "Runs the linter for the specified packages",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execLint(cfg, args, io)
		},
	)
}

func (c *lintCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.verbose, "verbose", false, "verbose output when lintning")
	fs.StringVar(&c.rootDir, "root-dir", "", "clone location of github.com/gnolang/gno (gnodev tries to guess it)")
	fs.IntVar(&c.setExitStatus, "set_exit_status", 1, "set exit status to 1 if any issues are found")
}

func execLint(cfg *lintCfg, args []string, io *commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	var (
		verbose = cfg.verbose
		rootDir = cfg.rootDir
	)
	if rootDir == "" {
		rootDir = guessRootDir()
	}

	pkgPaths, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	hasError := false
	addIssue := func(issue lintIssue) {
		hasError = true
		fmt.Fprint(io.Err, issue.String()+"\n")
	}

	for _, pkgPath := range pkgPaths {
		if verbose {
			fmt.Fprintf(io.Err, "Linting %q...\n", pkgPath)
		}

		// 'gno.mod' exists?
		gnoModPath := filepath.Join(pkgPath, "gno.mod")
		if !osm.FileExists(gnoModPath) {
			addIssue(lintIssue{
				Code:       lintNoGnoMod,
				Confidence: 1,
				Location:   pkgPath,
				Msg:        "missing 'gno.mod' file",
			})
		}

		// TODO: add more checkers
	}

	if hasError && cfg.setExitStatus != 0 {
		os.Exit(cfg.setExitStatus)
	}
	return nil
}

type lintCode int

const (
	lintUnknown  lintCode = 0
	lintNoGnoMod lintCode = iota
	// TODO: add new linter codes here.
)

type lintIssue struct {
	Code       lintCode
	Msg        string
	Confidence float64 // 1 is 100%
	Location   string  // file:line, or equivalent
	// TODO: consider writing fix suggestions
}

func (i lintIssue) String() string {
	// TODO: consider crafting a doc URL based on Code.
	return fmt.Sprintf("%s: %s (code=%d).", i.Location, i.Msg, i.Code)
}
