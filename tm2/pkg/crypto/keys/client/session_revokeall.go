package client

import (
	"context"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type SessionRevokeAllCfg struct {
	RootCfg *SessionCfg
}

// NewSessionRevokeAllCmd creates a gnokey session revokeall command
func NewSessionRevokeAllCmd(rootCfg *SessionCfg, io commands.IO) *commands.Command {
	cfg := &SessionRevokeAllCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "revokeall",
			ShortUsage: "session revokeall [flags] <master-key-name or address>",
			ShortHelp:  "revoke all session accounts",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSessionRevokeAll(cfg, args, io)
		},
	)
}

func (c *SessionRevokeAllCfg) RegisterFlags(fs *flag.FlagSet) {
}

func execSessionRevokeAll(cfg *SessionRevokeAllCfg, args []string, io commands.IO) error {
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

	// parse gas wanted & fee.
	gaswanted := cfg.RootCfg.RootCfg.GasWanted
	gasfee, err := std.ParseCoin(cfg.RootCfg.RootCfg.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// construct msg & tx and marshal.
	msg := auth.MsgRevokeAllSessions{
		Creator: masterAddr,
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
