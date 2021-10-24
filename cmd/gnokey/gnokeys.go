// Dedicated to my love, Lexi.
package main

import (
	"fmt"
	"io/ioutil"
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

	client.AddApp(makeTxApp, "maketx", "compose a tx document to sign", nil)
	err := client.RunMain(cmd, exec, args)
	if err != nil {
		cmd.ErrPrintfln("%s", err.Error())
		cmd.ErrPrintfln("%#v", err)
		return // exit
	}
}

var makeTxApps client.AppList = []client.AppItem{
	{makeAddPackageTxApp,
		"addpkg", "upload new package",
		defaultMakeAddPackageTxOptions},
	{makeExecTxApp,
		"exec", "execute statement",
		defaultmakeExecTxOptions},
}

func makeTxApp(cmd *command.Command, args []string, iopts interface{}) error {
	// show help message.
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		cmd.Println("available subcommands:")
		for _, appItem := range makeTxApps {
			cmd.Printf("  %s - %s\n", appItem.Name, appItem.Desc)
		}
		return nil
	}

	// switch on first argument.
	for _, appItem := range makeTxApps {
		if appItem.Name == args[0] {
			err := cmd.Run(appItem.App, args[1:], appItem.Defaults)
			return err // done
		}
	}

	// unknown app subcommand!
	return errors.New("unknown subcommand " + args[0])
}

type BaseTxOptions struct {
	GasWanted int64  `flag:"gas-wanted" help:"gas requested for tx"`
	GasFee    string `flag:"gas-fee" help:"gas payment fee"`
	Memo      string `flag:"memo" help:"any descriptive text"`
}

//----------------------------------------
// makeAddPackageTx

type makeAddPackageTxOptions struct {
	client.BaseOptions        // home,...
	BaseTxOptions             // gas-wanted, gas-fee, memo, ...
	PkgPath            string `flag:"pkgpath" help:"package path (required)"`
	PkgDir             string `flag:"pkgdir" help:"path to package files (required)"`
	Deposit            string `flag:"deposit" help:"deposit coins"`
}

var defaultMakeAddPackageTxOptions = makeAddPackageTxOptions{
	PkgPath: "", // must override
	PkgDir:  "", // must override
	Deposit: "",
}

func makeAddPackageTxApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(makeAddPackageTxOptions)
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

	// parse deposit.
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

	// parse gas wanted & fee.
	gaswanted := opts.GasWanted
	gasfee, err := std.ParseCoin(opts.GasFee)
	if err != nil {
		panic(err)
	}
	// construct msg & tx and marshal.
	msg := vm.MsgAddPackage{
		Creator: creator,
		PkgPath: opts.PkgPath,
		Files:   namedfiles,
		Deposit: deposit,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       opts.Memo,
	}
	fmt.Println(string(amino.MustMarshalJSON(tx)))
	return nil
}

//----------------------------------------
// makeExecTxApp

type makeExecTxOptions struct {
	client.BaseOptions        // home,...
	BaseTxOptions             // gas-wanted, gas-fee, memo, ...
	PkgPath            string `flag:"pkgpath" help:"package path (required)"`
	Stmt               string `flag:"stmt" help:"statement to execute" (required)"`
	StmtFile           string `flag:"stmtfile" help:"statement file instead of inline"`
	Send               string `flag:"send" help:"send coins"`
}

var defaultmakeExecTxOptions = makeExecTxOptions{
	PkgPath:  "", // must override
	Stmt:     "", // must override
	StmtFile: "", // must override
	Send:     "",
}

func makeExecTxApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(makeExecTxOptions)
	if opts.PkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	if opts.Stmt == "" && opts.StmtFile == "" {
		return errors.New("stmt (or stmtfile) not specified")
	}
	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: exec <keyname>")
		return errors.New("invalid args")
	}
	if opts.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if opts.GasFee == "" {
		return errors.New("gas-fee not specified")
	}

	// read statement.
	stmt := opts.Stmt
	if opts.StmtFile != "" {
		bz, err := ioutil.ReadFile(opts.StmtFile)
		if err != nil {
			return errors.Wrap(err, "loading statement file")
		}
		stmt = string(bz)
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
	caller := info.GetAddress()
	// info.GetPubKey()

	// Parse deposit.
	send, err := std.ParseCoins(opts.Send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}

	// parse gas wanted & fee.
	gaswanted := opts.GasWanted
	gasfee, err := std.ParseCoin(opts.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// construct msg & tx and marshal.
	msg := vm.MsgExec{
		Caller:  caller,
		PkgPath: opts.PkgPath,
		Stmt:    stmt,
		Send:    send,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       opts.Memo,
	}
	fmt.Println(string(amino.MustMarshalJSON(tx)))
	return nil
}
