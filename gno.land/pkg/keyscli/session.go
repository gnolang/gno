package keyscli

import (
	"context"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

// Note: If session accounts need to be used independent of gno.land, this may need to be moved to tm2/pkg/crypto/keys/client.
type SessionCfg struct {
	RootCfg *client.MakeTxCfg
}

func NewSessionCmd(rootCfg *client.MakeTxCfg, io commands.IO) *commands.Command {
	cfg := &SessionCfg{
		RootCfg: rootCfg,
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "session",
			ShortUsage: "session <subcommand> [flags]",
			ShortHelp:  "create or revoke session accounts",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSession(cfg, args, io)
		},
	)

	cmd.AddSubCommands(
		NewSessionCreateCmd(cfg, io),
		NewSessionRevokeCmd(cfg, io),
		NewSessionRevokeAllCmd(cfg, io),
	)

	return cmd
}

func (c *SessionCfg) RegisterFlags(fs *flag.FlagSet) {
}

func execSession(cfg *SessionCfg, args []string, io commands.IO) error {
	return nil
}
