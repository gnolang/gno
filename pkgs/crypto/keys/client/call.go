package client

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/sdk/vm"
	"github.com/gnolang/gno/pkgs/std"
)

type callCfg struct {
	rootCfg *makeTxCfg

	send     string
	pkgPath  string
	funcName string
	args     commands.StringArr
}

func newCallCmd(rootCfg *makeTxCfg) *commands.Command {
	cfg := &callCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "call",
			ShortUsage: "call [flags] <key-name or address>",
			ShortHelp:  "Executes a Realm function call",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execCall(cfg, args, commands.NewDefaultIO())
		},
	)
}

func (c *callCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.send,
		"send",
		"",
		"send amount",
	)

	fs.StringVar(
		&c.pkgPath,
		"pkgpath",
		"",
		"package path (required)",
	)

	fs.StringVar(
		&c.funcName,
		"func",
		"",
		"contract to call (required)",
	)

	fs.Var(
		&c.args,
		"args",
		"arguments to contract",
	)
}

func execCall(cfg *callCfg, args []string, io *commands.IO) error {
	if cfg.pkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	if cfg.funcName == "" {
		return errors.New("func not specified")
	}
	if len(args) != 1 {
		return flag.ErrHelp
	}
	if cfg.rootCfg.gasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.rootCfg.gasFee == "" {
		return errors.New("gas-fee not specified")
	}

	// read statement.
	fnc := cfg.funcName

	// read account pubkey.
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.rootCfg.rootCfg.Home)
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
	send, err := std.ParseCoins(cfg.send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}

	// parse gas wanted & fee.
	gaswanted := cfg.rootCfg.gasWanted
	gasfee, err := std.ParseCoin(cfg.rootCfg.gasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// construct msg & tx and marshal.
	msg := vm.MsgCall{
		Caller:  caller,
		Send:    send,
		PkgPath: cfg.pkgPath,
		Func:    fnc,
		Args:    cfg.args,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.rootCfg.memo,
	}

	if cfg.rootCfg.broadcast {
		err := signAndBroadcast(cfg.rootCfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}
