package main

import (

	//	"io/ioutil"
	"os"

	//	"strings"

	//	"github.com/gnolang/gno/pkgs/amino"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys/client"
	//	"github.com/gnolang/gno/pkgs/sdk/vm"
	//	"github.com/gnolang/gno/pkgs/std"
)

func main() {
	cmd := command.NewStdCommand()

	// set default options.

	// customize call to command.
	// insert args and options here.
	// TODO: use flags or */pflags.

	exec := os.Args[0]
	args := os.Args[1:]

	client.AddApp(backupKeyApp, "bkkey", "create a backup key to a backup keybase", client.DefaultBaseOptions)
	client.AddApp(signBkApp, "signbk", "sign a transaction with the primary key and backup key", client.DefaultSignOptions)
	client.AddApp(listBkApp, "listbk", "list all know keys including back up keys", client.DefaultListOptions)

	err := client.RunMain(cmd, exec, args)
	if err != nil {
		cmd.ErrPrintfln("%s", err.Error())
		//cmd.ErrPrintfln("%#v", err)
		return // exit
	}
}
