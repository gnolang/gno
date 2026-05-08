package client

import (
	"context"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type SessionCfg struct {
	RootCfg *MakeTxCfg
}

func NewSessionCmd(rootCfg *MakeTxCfg, io commands.IO) *commands.Command {
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
