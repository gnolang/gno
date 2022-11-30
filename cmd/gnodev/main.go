package main

import (
	"os"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

func main() {
	cmd := command.NewStdCommand()
	exec := os.Args[0]
	args := os.Args[1:]
	err := runMain(cmd, exec, args)
	if err != nil {
		cmd.ErrPrintfln("%s", err.Error())
		// cmd.ErrPrintfln("%#v", err)
		os.Exit(1)
	}
}

type (
	AppItem = command.AppItem
	AppList = command.AppList
)

var mainApps AppList = []AppItem{
	{
		App:      buildApp,
		Name:     "build",
		Desc:     "build a gno package",
		Defaults: defaultBuildOptions,
	},
	{
		App:      precompileApp,
		Name:     "precompile",
		Desc:     "precompile .gno to .go",
		Defaults: defaultPrecompileOptions,
	},
	{
		App:      testApp,
		Name:     "test",
		Desc:     "test a gno package",
		Defaults: defaultTestOptions,
	},
	{
		App:      replApp,
		Name:     "repl",
		Desc:     "start a GnoVM REPL",
		Defaults: defaultReplOptions,
	},

	// fmt -- gofmt
	// clean
	// graph
	// vendor -- download deps from the chain in vendor/
	// list -- list packages
	// run -- call render(), or maybe create a new main?
	// publish/release
	// generate
	// doc -- godoc
	// "vm" -- starts an in-memory chain that can be interacted with?
	// bug -- start a bug report
	// version -- show gnodev, golang versions
}

func runMain(cmd *command.Command, exec string, args []string) error {
	// show help message.
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		cmd.Println("available subcommands:")
		for _, appItem := range mainApps {
			cmd.Printf("  %s - %s\n", appItem.Name, appItem.Desc)
		}
		return nil
	}

	// switch on first argument.
	for _, appItem := range mainApps {
		if appItem.Name == args[0] {
			err := cmd.Run(appItem.App, args[1:], appItem.Defaults)
			return err // done
		}
	}

	// unknown app command!
	return errors.New("unknown command " + args[0])
}
