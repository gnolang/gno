package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/gnolang/gnomod"
)

type modCfg struct {
	verbose bool
}

func newModCmd() *commands.Command {
	cfg := &modCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "mod",
			ShortUsage: "mod [flags] <command>",
			ShortHelp:  "Manage gno.mod",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMod(cfg, args)
		},
	)
}

func (c *modCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"verbose",
		false,
		"verbose output",
	)
}

func execMod(cfg *modCfg, args []string) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	switch args[0] {
	case "download":
		if err := runModDownload(cfg); err != nil {
			return fmt.Errorf("mod download: %w", err)
		}
	default:
		return fmt.Errorf("invalid command: %s", args[0])
	}

	return nil
}

func runModDownload(cfg *modCfg) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	modPath := filepath.Join(path, "gno.mod")
	if !isFileExist(modPath) {
		return errors.New("gno.mod not found")
	}

	// read gno.mod
	data, err := os.ReadFile(modPath)
	if err != nil {
		return fmt.Errorf("readfile %q: %w", modPath, err)
	}

	// parse gno.mod
	gnoMod, err := gnomod.Parse(modPath, data)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	// validate gno.mod
	if err := gnoMod.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	// fetch dependencies
	if err := gnoMod.FetchDeps(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	if err := gnomod.Sanitize(gnoMod); err != nil {
		return fmt.Errorf("sanitize: %w", err)
	}

	// write go.mod file
	err = gnoMod.WriteToPath(path)
	if err != nil {
		return fmt.Errorf("write go.mod file: %w", err)
	}

	return nil
}
