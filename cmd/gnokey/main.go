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
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/errors"
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
		newMakeTxCmd(cfg),
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

type SignBroadcastOptions struct {
	GasWanted int64
	GasFee    string
	Memo      string

	Broadcast bool
	ChainID   string
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
