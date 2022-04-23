// Dedicated to my love, Lexi.
package main

import (
	"fmt"
	"os"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/crypto/keys/client"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
)

func main() {
	cmd := command.NewStdCommand()
	exec := os.Args[0]
	args := os.Args[1:]
	// extend default crypto/keys/client with maketx.
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
	{makeCallTxApp,
		"call", "call public function",
		defaultMakeCallTxOptions},
	{makeSendTxApp,
		"send", "send coins",
		defaultMakeSendTxOptions},
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

type SignBroadcastOptions struct {
	GasWanted int64  `flag:"gas-wanted" help:"gas requested for tx"`
	GasFee    string `flag:"gas-fee" help:"gas payment fee"`
	Memo      string `flag:"memo" help:"any descriptive text"`

	Broadcast bool   `flag:"broadcast" help:"sign and broadcast"`
	ChainID   string `flag:"chainid" help:"chainid to sign for (only useful if --broadcast)"`
}

//----------------------------------------
// makeAddPackageTx

type makeAddPackageTxOptions struct {
	client.BaseOptions          // home,...
	SignBroadcastOptions        // gas-wanted, gas-fee, memo, ...
	PkgPath              string `flag:"pkgpath" help:"package path (required)"`
	PkgDir               string `flag:"pkgdir" help:"path to package files (required)"`
	Deposit              string `flag:"deposit" help:"deposit coins"`
}

var defaultMakeAddPackageTxOptions = makeAddPackageTxOptions{
	BaseOptions: client.DefaultBaseOptions,
	PkgPath:     "", // must override
	PkgDir:      "", // must override
	Deposit:     "",
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
		cmd.ErrPrintfln("Usage: addpkg <keyname or address>")
		return errors.New("invalid args")
	}

	// read account pubkey.
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
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

	// open files in directory as MemPackage.
	memPkg := gno.ReadMemPackage(opts.PkgDir, opts.PkgPath)

	// precompile and validate syntax
	err = gno.PrecompileAndCheckMempkg(memPkg)
	if err != nil {
		panic(err)
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
		Package: memPkg,
		Deposit: deposit,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       opts.Memo,
	}

	if opts.Broadcast {
		err := signAndBroadcast(cmd, args, tx, opts.BaseOptions, opts.SignBroadcastOptions)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}

//----------------------------------------
// makeCallTxApp

type makeCallTxOptions struct {
	client.BaseOptions            // home,...
	SignBroadcastOptions          // gas-wanted, gas-fee, memo, ...
	Send                 string   `flag:"send" help:"send coins"`
	PkgPath              string   `flag:"pkgpath" help:"package path (required)"`
	Func                 string   `flag:"func" help:"contract to call" (required)"`
	Args                 []string `flag:"args" help:"arguments to contract"`
}

var defaultMakeCallTxOptions = makeCallTxOptions{
	BaseOptions: client.DefaultBaseOptions,
	PkgPath:     "", // must override
	Func:        "", // must override
	Args:        nil,
	Send:        "",
}

func makeCallTxApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(makeCallTxOptions)
	if opts.PkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	if opts.Func == "" {
		return errors.New("func not specified")
	}
	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: call <keyname or address>")
		return errors.New("invalid args")
	}
	if opts.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if opts.GasFee == "" {
		return errors.New("gas-fee not specified")
	}

	// read statement.
	fnc := opts.Func

	// read account pubkey.
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	caller := info.GetAddress()
	// info.GetPubKey()

	// Parse send amount.
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
	msg := vm.MsgCall{
		Caller:  caller,
		Send:    send,
		PkgPath: opts.PkgPath,
		Func:    fnc,
		Args:    opts.Args,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       opts.Memo,
	}

	if opts.Broadcast {
		err := signAndBroadcast(cmd, args, tx, opts.BaseOptions, opts.SignBroadcastOptions)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}

func signAndBroadcast(cmd *command.Command, args []string, tx std.Tx, baseopts client.BaseOptions, txopts SignBroadcastOptions) error {
	// query account
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(baseopts.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	accountAddr := info.GetAddress()

	qopts := client.QueryOptions{
		Path: fmt.Sprintf("auth/accounts/%s", accountAddr),
	}
	qopts.Remote = baseopts.Remote
	qres, err := client.QueryHandler(qopts)
	if err != nil {
		return errors.Wrap(err, "query account")
	}
	var qret struct{ BaseAccount std.BaseAccount }
	err = amino.UnmarshalJSON(qres.Response.Data, &qret)
	if err != nil {
		return err
	}

	// sign tx
	var accountNumber uint64 = qret.BaseAccount.AccountNumber
	var sequence uint64 = qret.BaseAccount.Sequence
	sopts := client.SignOptions{
		Sequence:      &sequence,
		AccountNumber: &accountNumber,
		ChainID:       txopts.ChainID,
		NameOrBech32:  nameOrBech32,
		TxJson:        amino.MustMarshalJSON(tx),
	}
	sopts.Home = baseopts.Home
	if baseopts.Quiet {
		sopts.Pass, err = cmd.GetPassword("")
	} else {
		sopts.Pass, err = cmd.GetPassword("Enter password.")
	}
	if err != nil {
		return err
	}

	signedTx, err := client.SignHandler(sopts)
	if err != nil {
		return errors.Wrap(err, "sign tx")
	}

	// broadcast signed tx
	bopts := client.BroadcastOptions{
		Tx: signedTx,
	}
	bopts.Remote = baseopts.Remote
	bres, err := client.BroadcastHandler(bopts)
	if err != nil {
		return errors.Wrap(err, "broadcast tx")
	}
	if bres.CheckTx.IsErr() {
		return errors.New("transaction failed %#v\nlog %s", bres, bres.CheckTx.Log)
	} else if bres.DeliverTx.IsErr() {
		return errors.New("transaction failed %#v\nlog %s", bres, bres.DeliverTx.Log)
	} else {
		cmd.Println(string(bres.DeliverTx.Data))
		cmd.Println("OK!")
		cmd.Println("GAS WANTED:", bres.DeliverTx.GasWanted)
		cmd.Println("GAS USED:  ", bres.DeliverTx.GasUsed)
	}
	return nil
}

//----------------------------------------
// makeSendTxApp

type makeSendTxOptions struct {
	client.BaseOptions          // home,...
	SignBroadcastOptions        // gas-wanted, gas-fee, memo, ...
	Send                 string `flag:"send" help:"send coins"`
	To                   string `flag:"to" help:"destination address"`
}

var defaultMakeSendTxOptions = makeSendTxOptions{
	BaseOptions: client.DefaultBaseOptions,
	Send:        "", // must override
	To:          "", // must override
}

func makeSendTxApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(makeSendTxOptions)
	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: send <keyname or address>")
		return errors.New("invalid args")
	}
	if opts.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if opts.GasFee == "" {
		return errors.New("gas-fee not specified")
	}
	if opts.Send == "" {
		return errors.New("send (amount) must be specified")
	}
	if opts.To == "" {
		return errors.New("to (destination address) must be specified")
	}

	// read account pubkey.
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(opts.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	fromAddr := info.GetAddress()
	// info.GetPubKey()

	// Parse to address.
	toAddr, err := crypto.AddressFromBech32(opts.To)
	if err != nil {
		return err
	}

	// Parse send amount.
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
	msg := bank.MsgSend{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      send,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       opts.Memo,
	}

	if opts.Broadcast {
		err := signAndBroadcast(cmd, args, tx, opts.BaseOptions, opts.SignBroadcastOptions)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}
