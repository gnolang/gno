package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/gnolang/gnomod"
)

type modFlags struct {
	Verbose bool `flag:"verbose" help:"verbose"`
}

var defaultModFlags = modFlags{
	Verbose: false,
}

func modApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(modFlags)

	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: mod [flags] <command>")
		return errors.New("invalid command")
	}

	switch args[0] {
	case "download":
		if err := runModDownload(&opts); err != nil {
			return fmt.Errorf("mod download: %w", err)
		}
	default:
		return fmt.Errorf("invalid command: %s", args[0])
	}

	return nil
}

func runModDownload(opts *modFlags) error {
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
