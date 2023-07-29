package client

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type execCfg struct {
	rootCfg *makeTxCfg
	send    string
	source  string
}

func newExecCmd(rootCfg *makeTxCfg) *commands.Command {
	cfg := &execCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "exec",
			ShortUsage: "exec [flags] <key-name or address>",
			ShortHelp:  "Executes arbitrary Gno code",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execExec(cfg, args, commands.NewDefaultIO())
		},
	)
}

func (c *execCfg) RegisterFlags(fs *flag.FlagSet) {}

func execExec(cfg *execCfg, args []string, io *commands.IO) error {
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
	// TODO: parse stdin
	source := cfg.source
	source = "package main\nfunc main() {println(\"42\")}"
	if source == "" {
		return errors.New("empty source")
	}
	// TODO: validate source

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

	// parse gas wanted & fee.
	gaswanted := cfg.rootCfg.gasWanted
	gasfee, err := std.ParseCoin(cfg.rootCfg.gasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// construct msg & tx and marshal.
	msg := vm.MsgExec{
		Caller: caller,
		Source: source,
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
