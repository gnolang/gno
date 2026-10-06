package client

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
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
		commands.HelpExec,
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

// rejectSessionMasterFlag errors when --master is set on a session-lifecycle
// command. Sessions can never sign create/revoke/revokeall msgs (the gno.land
// ante allowlist permanently blocks them to prevent privilege escalation), so
// the CLI catches the misuse before the user pays gas on a doomed broadcast.
func rejectSessionMasterFlag(cfg *MakeTxCfg) error {
	if cfg.Master != "" {
		return errors.New("--master cannot be used with session create/revoke/revokeall: session-lifecycle messages must be signed by the master key directly")
	}
	return nil
}
