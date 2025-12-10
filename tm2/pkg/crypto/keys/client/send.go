package client

import (
	"context"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type MakeSendCfg struct {
	RootCfg *MakeTxCfg

	Send string
	To   string
}

func NewMakeSendCmd(rootCfg *MakeTxCfg, io commands.IO) *commands.Command {
	cfg := &MakeSendCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "send",
			ShortUsage: "send [flags] <key-name or address>",
			ShortHelp:  "sends native currency",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execMakeSend(cfg, args, io)
		},
	)
}

func (c *MakeSendCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.Send,
		"send",
		"",
		"send amount",
	)

	fs.StringVar(
		&c.To,
		"to",
		"",
		"destination address",
	)
}

func execMakeSend(cfg *MakeSendCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	if cfg.RootCfg.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.RootCfg.GasFee == "" {
		return errors.New("gas-fee not specified")
	}
	if cfg.Send == "" {
		return errors.New("send (amount) must be specified")
	}
	if cfg.To == "" {
		return errors.New("to (destination address) must be specified")
	}

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
	fromAddr := info.GetAddress()
	// info.GetPubKey()

	// Parse to address.
	toAddr, err := crypto.AddressFromBech32(cfg.To)
	if err != nil {
		return err
	}

	// Parse send amount.
	send, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}

	// parse gas wanted & fee.
	gaswanted := cfg.RootCfg.GasWanted
	gasfee, err := std.ParseCoin(cfg.RootCfg.GasFee)
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
		Memo:       cfg.RootCfg.Memo,
	}

	if cfg.RootCfg.Broadcast {
		return ExecSignAndBroadcast(cfg.RootCfg, args, tx, io)
	}

	io.Println(string(amino.MustMarshalJSON(tx)))

	return nil
}
