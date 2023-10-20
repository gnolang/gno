package client

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type sendCfg struct {
	rootCfg *makeTxCfg

	send string
	to   string
}

func newSendCmd(rootCfg *makeTxCfg, io commands.IO) *commands.Command {
	cfg := &sendCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "send",
			ShortUsage: "send [flags] <key-name or address>",
			ShortHelp:  "Sends native currency",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSend(cfg, args, io)
		},
	)
}

func (c *sendCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.send,
		"send",
		"",
		"send amount",
	)

	fs.StringVar(
		&c.to,
		"to",
		"",
		"destination address",
	)
}

func execSend(cfg *sendCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	if cfg.rootCfg.gasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.rootCfg.gasFee == "" {
		return errors.New("gas-fee not specified")
	}
	if cfg.send == "" {
		return errors.New("send (amount) must be specified")
	}
	if cfg.to == "" {
		return errors.New("to (destination address) must be specified")
	}

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
	fromAddr := info.GetAddress()
	// info.GetPubKey()

	// Parse to address.
	toAddr, err := crypto.AddressFromBech32(cfg.to)
	if err != nil {
		return err
	}

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
	msg := bank.MsgSend{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      send,
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
