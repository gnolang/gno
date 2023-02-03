package main

import (
	"os"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

type (
	AppItem = command.AppItem
	AppList = command.AppList
)

var mainApps AppList = []AppItem{
	{txExportApp, "export", "export txs from node", defaultTxExportOptions},
	{txImportApp, "import", "import txs to node", defaultTxImportOptions},
}

func main() {
	cmd := command.NewStdCommand()
	args := os.Args[1:]

	// show help message.
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		cmd.Println("available subcommands:")
		for _, appItem := range mainApps {
			cmd.Printf("  %s - %s\n", appItem.Name, appItem.Desc)
		}

		return
	}

	// switch on first argument.
	for _, appItem := range mainApps {
		if appItem.Name == args[0] {
			err := cmd.Run(appItem.App, args[1:], appItem.Defaults)
			if err != nil {
				panic(err)
			}

			return
		}
	}

	// unknown app command!
	panic(errors.New("unknown command " + args[0]))
}
