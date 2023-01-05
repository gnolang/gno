package client

import (
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

type (
	AppItem = command.AppItem
	AppList = command.AppList
)

var mainApps AppList = []AppItem{
	{addApp, "add", "add key to keybase", DefaultAddOptions},
	{deleteApp, "delete", "delete key from keybase", DefaultDeleteOptions},
	{generateApp, "generate", "generate a new private key", DefaultGenerateOptions},
	{exportApp, "export", "export encrypted private key armor", DefaultExportOptions},
	{listApp, "list", "list all known keys", DefaultListOptions},
	{signApp, "sign", "sign a document", DefaultSignOptions},
	{verifyApp, "verify", "verify a document signature", DefaultVerifyOptions},
	{broadcastApp, "broadcast", "broadcast a signed document", DefaultBroadcastOptions},
	{queryApp, "query", "make an ABCI query", DefaultQueryOptions},
}

// For clients that want to extend the functionality of the base client.
func AddApp(app command.App, name string, desc string, defaults interface{}) {
	mainApps = append(mainApps, AppItem{
		App:      app,
		Name:     name,
		Desc:     desc,
		Defaults: defaults,
	})
}

func RunMain(cmd *command.Command, exec string, args []string) error {
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
