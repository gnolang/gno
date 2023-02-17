// Dedicated to my love, Lexi.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/cmd/common"
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/errors"
	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
)

const (
	mnemonicEntropySize = 256
)

type baseCfg struct {
	common.BaseOptions
}

func main() {
	cfg := &baseCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Manages private keys for the node",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		// TODO add
		newAddCmd(cfg),
		newDeleteCmd(cfg),
		newGenerateCmd(cfg),
		newExportCmd(cfg),
		newImportCmd(cfg),
		newListCmd(cfg),
		newSignCmd(cfg),
		newVerifyCmd(cfg),
		newQueryCmd(cfg),
		newBroadcastCmd(cfg),
	)

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}

func (c *baseCfg) RegisterFlags(fs *flag.FlagSet) {
	// Base options
	fs.StringVar(
		&c.Home,
		"home",
		common.DefaultBaseOptions.Home,
		"home directory",
	)

	fs.StringVar(
		&c.Remote,
		"remote",
		common.DefaultBaseOptions.Remote,
		"remote node URL",
	)

	fs.BoolVar(
		&c.Quiet,
		"quiet",
		common.DefaultBaseOptions.Quiet,
		"suppress output during execution",
	)

	fs.BoolVar(
		&c.InsecurePasswordStdin,
		"insecure-password-stdin",
		common.DefaultBaseOptions.Quiet,
		"WARNING! take password from stdin",
	)
}

// var makeTxApps client.AppList = []client.AppItem{
// 	{
// 		makeAddPackageTxApp,
// 		"addpkg", "upload new package",
// 		defaultMakeAddPackageTxOptions,
// 	},
// 	{
// 		makeCallTxApp,
// 		"call", "call public function",
// 		defaultMakeCallTxOptions,
// 	},
// 	{
// 		makeSendTxApp,
// 		"send", "send coins",
// 		defaultMakeSendTxOptions,
// 	},
// }

type SignBroadcastOptions struct {
	GasWanted int64
	GasFee    string
	Memo      string

	Broadcast bool
	ChainID   string
}

// ----------------------------------------
// makeAddPackageTx

type makeAddPackageTxOptions struct {
	common.BaseOptions          // home,...
	SignBroadcastOptions        // gas-wanted, gas-fee, memo, ...
	PkgPath              string `flag:"pkgpath" help:"package path (required)"`
	PkgDir               string `flag:"pkgdir" help:"path to package files (required)"`
	Deposit              string `flag:"deposit" help:"deposit coins"`
}

var defaultMakeAddPackageTxOptions = makeAddPackageTxOptions{
	BaseOptions: common.DefaultBaseOptions,
	// SignBroadcastOptions: defaultSignBroadcastOptions,
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
		// err := signAndBroadcast(cmd, args, tx, opts.BaseOptions, opts.SignBroadcastOptions)
		// if err != nil {
		// 	return err
		// }
	} else {
		fmt.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}

// ----------------------------------------
// makeCallTxApp

type makeCallTxOptions struct {
	common.BaseOptions            // home,...
	SignBroadcastOptions          // gas-wanted, gas-fee, memo, ...
	Send                 string   `flag:"send" help:"send coins"`
	PkgPath              string   `flag:"pkgpath" help:"package path (required)"`
	Func                 string   `flag:"func" help:"contract to call (required)"`
	Args                 []string `flag:"args" help:"arguments to contract"`
}

var defaultMakeCallTxOptions = makeCallTxOptions{
	BaseOptions: common.DefaultBaseOptions,
	// SignBroadcastOptions: defaultSignBroadcastOptions,
	PkgPath: "", // must override
	Func:    "", // must override
	Args:    nil,
	Send:    "",
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
		// err := signAndBroadcast(cmd, args, tx, opts.BaseOptions, opts.SignBroadcastOptions)
		// if err != nil {
		// 	return err
		// }
	} else {
		fmt.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}

// func signAndBroadcast(cmd *command.Command, args []string, tx std.Tx, baseopts client.BaseOptions, txopts SignBroadcastOptions) error {
// 	// query account
// 	nameOrBech32 := args[0]
// 	kb, err := keys.NewKeyBaseFromDir(baseopts.Home)
// 	if err != nil {
// 		return err
// 	}
// 	info, err := kb.GetByNameOrAddress(nameOrBech32)
// 	if err != nil {
// 		return err
// 	}
// 	accountAddr := info.GetAddress()
//
// 	qopts := client.QueryOptions{
// 		Path: fmt.Sprintf("auth/accounts/%s", accountAddr),
// 	}
// 	qopts.Remote = baseopts.Remote
// 	qres, err := client.QueryHandler(qopts)
// 	if err != nil {
// 		return errors.Wrap(err, "query account")
// 	}
// 	var qret struct{ BaseAccount std.BaseAccount }
// 	err = amino.UnmarshalJSON(qres.Response.Data, &qret)
// 	if err != nil {
// 		return err
// 	}
//
// 	// sign tx
// 	accountNumber := qret.BaseAccount.AccountNumber
// 	sequence := qret.BaseAccount.Sequence
// 	sopts := client.SignOptions{
// 		Sequence:      &sequence,
// 		AccountNumber: &accountNumber,
// 		ChainID:       txopts.ChainID,
// 		NameOrBech32:  nameOrBech32,
// 		TxJSON:        amino.MustMarshalJSON(tx),
// 	}
// 	sopts.Home = baseopts.Home
// 	if baseopts.Quiet {
// 		sopts.Pass, err = cmd.GetPassword("", baseopts.InsecurePasswordStdin)
// 	} else {
// 		sopts.Pass, err = cmd.GetPassword("Enter password.", baseopts.InsecurePasswordStdin)
// 	}
// 	if err != nil {
// 		return err
// 	}
//
// 	signedTx, err := client.SignHandler(sopts)
// 	if err != nil {
// 		return errors.Wrap(err, "sign tx")
// 	}
//
// 	// broadcast signed tx
// 	bopts := client.BroadcastOptions{
// 		Tx: signedTx,
// 	}
// 	bopts.Remote = baseopts.Remote
// 	bres, err := client.BroadcastHandler(bopts)
// 	if err != nil {
// 		return errors.Wrap(err, "broadcast tx")
// 	}
// 	if bres.CheckTx.IsErr() {
// 		return errors.Wrap(bres.CheckTx.Error, "check transaction failed: log:%s", bres.CheckTx.Log)
// 	}
// 	if bres.DeliverTx.IsErr() {
// 		return errors.Wrap(bres.DeliverTx.Error, "deliver transaction failed: log:%s", bres.DeliverTx.Log)
// 	}
// 	cmd.Println(string(bres.DeliverTx.Data))
// 	cmd.Println("OK!")
// 	cmd.Println("GAS WANTED:", bres.DeliverTx.GasWanted)
// 	cmd.Println("GAS USED:  ", bres.DeliverTx.GasUsed)
//
// 	return nil
// }

// ----------------------------------------
// makeSendTxApp

type makeSendTxOptions struct {
	common.BaseOptions          // home,...
	SignBroadcastOptions        // gas-wanted, gas-fee, memo, ...
	Send                 string `flag:"send" help:"send coins"`
	To                   string `flag:"to" help:"destination address"`
}

var defaultMakeSendTxOptions = makeSendTxOptions{
	BaseOptions: common.DefaultBaseOptions,
	// SignBroadcastOptions: defaultSignBroadcastOptions,
	Send: "", // must override
	To:   "", // must override
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
		// err := signAndBroadcast(cmd, args, tx, opts.BaseOptions, opts.SignBroadcastOptions)
		// if err != nil {
		// 	return err
		// }
	} else {
		fmt.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}
