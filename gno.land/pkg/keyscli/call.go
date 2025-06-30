package keyscli

import (
	"context"
	"flag"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type MakeCallCfg struct {
	RootCfg *client.MakeTxCfg

	Send     string
	PkgPath  string
	FuncName string
	Args     commands.StringArr
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

	// construct msg & tx and marshal.
	msg := vm.MsgCall{
		Caller:  caller,
		Send:    send,
		PkgPath: cfg.PkgPath,
		Func:    fnc,
		Args:    cfg.Args,
	}

	// Create transaction
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.Fee{}, // Will be set by EstimateOrSetFee
		Signatures: nil,
		Memo:       cfg.RootCfg.Memo,
	}

	// Handle gas estimation or manual fee setting
	if err := client.EstimateOrSetFee(cfg.RootCfg, &tx); err != nil {
		return errors.Wrap(err, "setting transaction fee")
	}

	if cfg.RootCfg.Broadcast {
		err := client.ExecSignAndBroadcast(cfg.RootCfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		io.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}
