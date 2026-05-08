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

type SessionCreateCfg struct {
	RootCfg *SessionCfg

	PublicKey   string
	AllowPaths  commands.StringArr
	SpendLimit  string
	SpendPeriod int64
}

// NewSessionCreateCmd creates a gnokey session create command
func NewSessionCreateCmd(rootCfg *SessionCfg, io commands.IO) *commands.Command {
	cfg := &SessionCreateCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "create",
			ShortUsage: "session create [flags] <master-key-name or address>",
			ShortHelp:  "create a session account",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSessionCreate(cfg, args, io)
		},
	)
}

func (c *SessionCreateCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.PublicKey,
		"pubkey",
		"",
		"the subaccount public key in bech32 format",
	)

	fs.Var(
		&c.AllowPaths,
		"allow-paths",
		"realm path prefixes (optional; omitted = unrestricted)",
	)

	fs.StringVar(
		&c.SpendLimit,
		"spend-limit",
		"",
		"max spend per period (optional; omitted = no spending)",
	)

	fs.Int64Var(
		&c.SpendPeriod,
		"spend-period",
		0,
		"seconds; 0 = lifetime cap",
	)
}

func execSessionCreate(cfg *SessionCreateCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
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
	if cfg.SpendPeriod < 0 {
		return errors.New("spend-period must be non-negative")
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
	msg := auth.MsgCreateSession{
		Creator:     masterAddr,
		SessionKey:  sessionPub,
		AllowPaths:  cfg.AllowPaths,
		SpendPeriod: cfg.SpendPeriod,
	}
	if cfg.SpendLimit != "" {
		// Parse send amount.
		spendLimit, err := std.ParseCoins(cfg.SpendLimit)
		if err != nil {
			return errors.Wrap(err, "parsing spend limit coins")
		}
		msg.SpendLimit = spendLimit
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
