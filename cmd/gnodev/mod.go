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

func newModCmd() *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "mod",
			ShortUsage: "mod <command>",
			ShortHelp:  "Manage gno.mod",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execMod(args)
		},
	)
}

func execMod(args []string) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	switch args[0] {
	case "download":
		if err := runModDownload(); err != nil {
			return fmt.Errorf("mod download: %w", err)
		}
	default:
		return fmt.Errorf("invalid command: %s", args[0])
	}

	return nil
}

func runModDownload() error {
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
	// sanitize gno.mod
	gnoMod.Sanitize()

	// validate gno.mod
	if err := gnoMod.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	// fetch dependencies
	if err := gnoMod.FetchDeps(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	gomod, err := gnomod.GnoToGoMod(*gnoMod)
	if err != nil {
		return fmt.Errorf("sanitize: %w", err)
	}

	// write go.mod file
	err = gomod.WriteToPath(path)
	if err != nil {
		return fmt.Errorf("write go.mod file: %w", err)
	}

	return nil
}
