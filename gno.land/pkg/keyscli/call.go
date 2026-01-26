package keyscli

import (
	"context"
	"flag"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mempool"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type MakeCallCfg struct {
	RootCfg *client.MakeTxCfg

	Send       string
	MaxDeposit string
	PkgPath    string
	FuncName   string
	Args       commands.StringArr
}

func NewMakeCallCmd(rootCfg *client.MakeTxCfg, io commands.IO) *commands.Command {
	cfg := &MakeCallCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "call",
			ShortUsage: "call [flags] <key-name or address>",
			ShortHelp:  "executes a realm function call",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMakeCall(cfg, args, io)
		},
	)
}

func (c *MakeCallCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.Send,
		"send",
		"",
		"send amount",
	)

	fs.StringVar(
		&c.MaxDeposit,
		"max-deposit",
		"",
		"max storage deposit",
	)

	fs.StringVar(
		&c.PkgPath,
		"pkgpath",
		"",
		"package path (required)",
	)

	fs.StringVar(
		&c.FuncName,
		"func",
		"",
		"contract to call (required)",
	)

	fs.Var(
		&c.Args,
		"args",
		"arguments to contract",
	)
}

func execMakeCall(cfg *MakeCallCfg, args []string, io commands.IO) error {
	if cfg.PkgPath == "" {
		return errors.New("pkgpath not specified")
	}
	if cfg.FuncName == "" {
		return errors.New("func not specified")
	}
	if len(args) != 1 {
		return flag.ErrHelp
	}
	if cfg.RootCfg.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.RootCfg.GasFee == "" {
		return errors.New("gas-fee not specified")
	}

	// read statement.
	fnc := cfg.FuncName

	// read account pubkey.
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.Home)
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
	send, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}

	// Parse deposit amount
	deposit, err := std.ParseCoins(cfg.MaxDeposit)
	if err != nil {
		return errors.Wrap(err, "parsing storage deposit coins")
	}

	// parse gas wanted & fee.
	gaswanted := cfg.RootCfg.GasWanted
	gasfee, err := std.ParseCoin(cfg.RootCfg.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// construct msg & tx and marshal.
	msg := vm.MsgCall{
		Caller:     caller,
		Send:       send,
		MaxDeposit: deposit,
		PkgPath:    cfg.PkgPath,
		Func:       fnc,
		Args:       cfg.Args,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.RootCfg.Memo,
	}

	if cfg.RootCfg.Broadcast {
		cfg.RootCfg.RootCfg.OnTxSuccess = func(tx std.Tx, res *mempool.ResultBroadcastTxCommit) {
			PrintTxInfo(tx, res, io)
		}
		err := client.ExecSignAndBroadcast(cfg.RootCfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		io.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}
