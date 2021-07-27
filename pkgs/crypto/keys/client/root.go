package client

import (
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

type AppItem struct {
	App      command.App
	Name     string      // arg name
	Desc     string      // short (single line) description of app
	Defaults interface{} // default options
	// Help string // long form help
}

type AppList []AppItem

var mainApp AppList = []AppItem{
	{addApp, "add", "add key to keybase", DefaultAddOptions},
	{deleteApp, "delete", "delete key from keybase", DefaultDeleteOptions},
	{generateApp, "generate", "generate a new private key", DefaultGenerateOptions},
	{listApp, "list", "list all known keys", DefaultListOptions},
	{signApp, "sign", "sign a document", DefaultSignOptions},
	{verifyApp, "verify", "verify a document signature", DefaultVerifyOptions},
}

func RunMain(cmd *command.Command, exec string, args []string) error {

	// show help message.
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		cmd.Println("available subcommands:")
		for _, appItem := range mainApp {
			cmd.Printf("  %s - %s\n", appItem.Name, appItem.Desc)
		}
		return nil
	}

	// switch on first argument.
	for _, appItem := range mainApp {
		if appItem.Name == args[0] {
			err := cmd.Run(appItem.App, args[1:], appItem.Defaults)
			return err // done
		}
	}

	// unknown app command!
	return errors.New("unknown command " + args[0])

}
