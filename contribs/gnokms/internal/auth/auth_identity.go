package auth

import (
	"context"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newAuthIdentity creates the auth identity subcommand.
func newAuthIdentityCmd(rootCfg *common.AuthFlags, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "identity",
			ShortUsage: "auth identity [flags]",
			ShortHelp:  "prints the identity public key of gnokms",
			LongHelp:   "Prints the identity public key of gnokms. This should be added to the authorized keys list of a client to allow it to authenticate with gnokms.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, _ []string) error {
			return execAuthIdentity(rootCfg, io)
		},
	)
}

func execAuthIdentity(rootCfg *common.AuthFlags, io commands.IO) error {
	// Load the auth keys file.
	authKeysFile, err := loadAuthKeysFile(rootCfg)
	if err != nil {
		return err
	}

	// Print the identity public key.
	io.Printfln("Server public key: %q", authKeysFile.ServerIdentity.PubKey)

	return nil
}
