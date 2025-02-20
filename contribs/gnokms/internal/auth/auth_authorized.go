package auth

import (
	"context"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newAuthAuthorizedCmd creates the auth authorized subcommand.
func newAuthAuthorizedCmd(rootCfg *common.AuthFlags, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "authorized",
			ShortUsage: "auth authorized [flags] [<arg>...]",
			ShortHelp:  "manipulates the list of authorized client public keys",
			LongHelp:   "Manipulates the list of authorized client public keys by adding, removing, or listing them.",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newAuthAuthorizedAddCmd(rootCfg, io),
		newAuthAuthorizedRemoveCmd(rootCfg, io),
		newAuthAuthorizedListCmd(rootCfg, io),
	)

	return cmd
}

// newAuthAuthorizedAddCmd creates the auth authorized add subcommand.
func newAuthAuthorizedAddCmd(rootCfg *common.AuthFlags, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "add",
			ShortUsage: "auth authorized add [flags] <public-key> [<public-key>...]",
			ShortHelp:  "adds client public keys to the list of authorized keys",
			LongHelp:   "Adds public keys to the list of authorized keys to authenticate clients with gnokms. The public keys must be of type ed25519 and encoded in bech32 format.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execAuthAuthorizedAdd(args, rootCfg, io)
		},
	)
}

func execAuthAuthorizedAdd(args []string, rootCfg *common.AuthFlags, io commands.IO) error {
	// Load the auth keys file.
	authKeysFile, err := loadAuthKeysFile(rootCfg)
	if err != nil {
		return err
	}

	_ = authKeysFile

	return nil
}

// newAuthAuthorizedRemoveCmd creates the auth authorized remove subcommand.
func newAuthAuthorizedRemoveCmd(rootCfg *common.AuthFlags, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "remove",
			ShortUsage: "auth authorized remove [flags] <public-key> [<public-key>...]",
			ShortHelp:  "removes client public keys from the list of authorized keys",
			LongHelp:   "Removes public keys from the list of authorized keys.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execAuthAuthorizedRemove(args, rootCfg, io)
		},
	)
}

func execAuthAuthorizedRemove(args []string, rootCfg *common.AuthFlags, io commands.IO) error {
	// Load the auth keys file.
	authKeysFile, err := loadAuthKeysFile(rootCfg)
	if err != nil {
		return err
	}

	_ = authKeysFile

	return nil
}

// newAuthAuthorizedListCmd creates the auth authorized list subcommand.
func newAuthAuthorizedListCmd(rootCfg *common.AuthFlags, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "auth authorized list [flags]",
			ShortHelp:  "lists the client authorized keys",
			LongHelp:   "Lists the public keys of the clients that are authorized to authenticate with gnokms.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execAuthAuthorizedList(args, rootCfg, io)
		},
	)
}

func execAuthAuthorizedList(args []string, rootCfg *common.AuthFlags, io commands.IO) error {
	// Load the auth keys file.
	authKeysFile, err := loadAuthKeysFile(rootCfg)
	if err != nil {
		return err
	}

	_ = authKeysFile

	return nil
}
