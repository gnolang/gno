package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/keys/client"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
)

func main() {

	cmd := command.NewStdCommand()

	// set default options.

	// customize call to command.
	// insert args and options here.
	// TODO: use flags or */pflags.

	exec := os.Args[0]
	args := os.Args[1:]

	client.AddApp(maketxApp, "maketx", "compose a tx document to sign", nil)
	err := client.RunMain(cmd, exec, args)
	if err != nil {
		cmd.ErrPrintln(err.Error())
		return // exit
	}
}

var maketxApps client.AppList = []client.AppItem{
	{makeAddPackageTxApp, "addpkg", "upload new package", DefaultAddPackageTxOptions},
	{makeEvalTxApp, "eval", "evaluate expression", nil},
}

func maketxApp(cmd *command.Command, args []string, iopts interface{}) error {
	// show help message.
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		cmd.Println("available subcommands:")
		for _, appItem := range maketxApps {
			cmd.Printf("  %s - %s\n", appItem.Name, appItem.Desc)
		}
		return nil
	}

	// switch on first argument.
	for _, appItem := range maketxApps {
		if appItem.Name == args[0] {
			err := cmd.Run(appItem.App, args[1:], appItem.Defaults)
			return err // done
		}
	}

	// unknown app subcommand!
	return errors.New("unknown subcommand " + args[0])
}

type AddPackageTxOptions struct {
	client.BaseOptions        // home,...
	PkgPath            string `flag:"pkgpath" help:"package path (required)"`
	PkgDir             string `flag:"pkgdir" help:"path to package files (required)"`
	Deposit            string `flag:"deposit" help:"deposit coins"`
}

var DefaultAddPackageTxOptions = AddPackageTxOptions{
	PkgPath: "", // must override
	PkgDir:  "", // must override
	Deposit: "",
}

func makeAddPackageTxApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(AddPackageTxOptions)
	if opts.PkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	if opts.PkgDir == "" {
		return errors.New("pkgdir not specified")
	}
	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: addpkg <keyname>")
		return errors.New("invalid args")
	}

	// read account pubkey.
	name := args[0]
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}
	info, err := kb.Get(name)
	if err != nil {
		return err
	}
	creator := info.GetAddress()
	// info.GetPubKey()

	// Parse deposit.
	deposit, err := std.ParseCoins(opts.Deposit)
	if err != nil {
		panic(err)
	}

	// read all files.
	dir, err := os.Open(opts.PkgDir)
	if err != nil {
		panic(err)
	}
	defer dir.Close()
	entries, err := dir.Readdir(0)
	if err != nil {
		panic(err)
	}

	// For each file in the directory, filter by pattern
	namedfiles := []vm.NamedFile{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".go") {
			fpath := filepath.Join(
				opts.PkgDir, name)
			body, err := os.ReadFile(fpath)
			if err != nil {
				return errors.Wrap(err, "reading gno file")
			}
			namedfiles = append(namedfiles,
				vm.NamedFile{
					Name: name,
					Body: string(body),
				})
		}
	}

	msg := vm.MsgAddPackage{
		Creator: creator,
		PkgPath: opts.PkgPath,
		Files:   namedfiles,
		Deposit: deposit,
	}
	fmt.Println(string(amino.MustMarshalJSONAny(msg)))
	return nil
}

func makeEvalTxApp(cmd *command.Command, args []string, iopts interface{}) error {
	panic("WOOT")
}
