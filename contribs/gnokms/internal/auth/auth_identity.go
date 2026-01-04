package auth

import (
	"context"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newAuthIdentity creates the auth identity subcommand.
func newAuthIdentityCmd(rootCfg *common.AuthFlags, io commands.IO) *commands.Command {
	cfg := &authRawFlags{
		auth: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "identity",
			ShortUsage: "auth identity [flags]",
			ShortHelp:  "prints the identity public key of gnokms",
			LongHelp:   "Prints the identity public key of gnokms. This should be added to the authorized keys list of a client to allow it to authenticate with gnokms.",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execAuthIdentity(cfg, io)
		},
	)
}

func execAuthIdentity(authRawCfg *authRawFlags, io commands.IO) error {
	// Load the auth keys file.
	authKeysFile, err := loadAuthKeysFile(authRawCfg.auth)
	if err != nil {
		return err
	}

	// Print the identity public key.
	if authRawCfg.raw {
		io.Println(authKeysFile.ServerIdentity.PubKey)
	} else {
		io.Printfln("Server public key: %q", authKeysFile.ServerIdentity.PubKey)
	}

	return nil
}
