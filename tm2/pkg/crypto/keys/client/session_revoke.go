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
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type SessionRevokeCfg struct {
	RootCfg *SessionCfg

	PublicKey string
}

// NewSessionRevokeCmd creates a gnokey session revoke command
func NewSessionRevokeCmd(rootCfg *SessionCfg, io commands.IO) *commands.Command {
	cfg := &SessionRevokeCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "revoke",
			ShortUsage: "session revoke [flags] <master-key-name or address>",
			ShortHelp:  "revoke a session account",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSessionRevoke(cfg, args, io)
		},
	)
}

func (c *SessionRevokeCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.PublicKey,
		"pubkey",
		"",
		"the subaccount public key in bech32 format",
	)
}

func execSessionRevoke(cfg *SessionRevokeCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	if err := rejectSessionMasterFlag(cfg.RootCfg.RootCfg); err != nil {
		return err
	}
	if cfg.RootCfg.RootCfg.GasWanted == 0 {
		return errors.New("gas-wanted not specified")
	}
	if cfg.RootCfg.RootCfg.GasFee == "" {
		return errors.New("gas-fee not specified")
	}
	if cfg.PublicKey == "" {
		return errors.New("pubkey must be specified")
	}

	// read account pubkey.
	nameOrBech32 := args[0]
	kb, err := keys.NewKeyBaseFromDir(cfg.RootCfg.RootCfg.RootCfg.Home)
	if err != nil {
		return err
	}
	info, err := kb.GetByNameOrAddress(nameOrBech32)
	if err != nil {
		return err
	}
	masterAddr := info.GetAddress()

	// Parse the public key
	sessionPub, err := crypto.PubKeyFromBech32(cfg.PublicKey)
	if err != nil {
		return fmt.Errorf("unable to parse public key from bech32, %w", err)
	}

	// parse gas wanted & fee.
	gaswanted := cfg.RootCfg.RootCfg.GasWanted
	gasfee, err := std.ParseCoin(cfg.RootCfg.RootCfg.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// construct msg & tx and marshal.
	msg := auth.MsgRevokeSession{
		Creator:    masterAddr,
		SessionKey: sessionPub,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gaswanted, gasfee),
		Signatures: nil,
		Memo:       cfg.RootCfg.RootCfg.Memo,
	}

	if cfg.RootCfg.RootCfg.Broadcast {
		err := ExecSignAndBroadcast(cfg.RootCfg.RootCfg, args, tx, io)
		if err != nil {
			return err
		}
	} else {
		io.Println(string(amino.MustMarshalJSON(tx)))
	}
	return nil
}
