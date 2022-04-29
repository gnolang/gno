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
	{buildApp, "build", "build a gno package", DefaultBuildOptions},
	{precompileApp, "precompile", "precompile .gno to .go", DefaultPrecompileOptions},
	{testApp, "test", "test a gno package", DefaultTestOptions},

	// fmt
	// publish
	// generate
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
