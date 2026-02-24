package auth

import (
	"context"
	"flag"
	"fmt"
	"slices"

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

// manipulatesAuthorizedKeys manipulates the list of authorized client public keys
// using the given process function.
func manipulatesAuthorizedKeys(
	rootCfg *common.AuthFlags,
	args []string,
	processKeys func([]string, []string) []string,
) error {
	// Load the auth keys file.
	authKeysFile, err := loadAuthKeysFile(rootCfg)
	if err != nil {
		return err
	}

	// Validate the public keys passed as arguments.
	for _, publicKey := range args {
		if _, err := common.Bech32ToEd25519PubKey(publicKey); err != nil {
			return fmt.Errorf("invalid public key %q: %w", publicKey, err)
		}
	}

	// Sort and deduplicate the keys.
	publicKeys := common.SortAndDeduplicate(args)

	// Process the keys.
	authKeysFile.ClientAuthorizedKeys = processKeys(authKeysFile.ClientAuthorizedKeys, publicKeys)

	// Save the auth keys file.
	if err := authKeysFile.Save(rootCfg.AuthKeysFile); err != nil {
		return fmt.Errorf("unable to save the auth keys file: %w", err)
	}

	return nil
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
			return execAuthAuthorizedAdd(rootCfg, args, io)
		},
	)
}

func execAuthAuthorizedAdd(rootCfg *common.AuthFlags, args []string, io commands.IO) error {
	// Check that at least one public key is provided.
	if len(args) == 0 {
		io.ErrPrintln("error: at least one public key must be provided\n")
		return flag.ErrHelp
	}

	// Add the public keys to the authorized keys list.
	return manipulatesAuthorizedKeys(rootCfg, args, func(current []string, updates []string) []string {
		for _, publicKey := range updates {
			if _, found := slices.BinarySearch(current, publicKey); found {
				io.Printfln("Public key %q already in the authorized keys list.", publicKey)
			} else {
				current = append(current, publicKey)
				io.Printfln("Public key %q added to the authorized keys list.", publicKey)
			}
		}
		return current
	})
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
			return execAuthAuthorizedRemove(rootCfg, args, io)
		},
	)
}

func execAuthAuthorizedRemove(rootCfg *common.AuthFlags, args []string, io commands.IO) error {
	// Check that at least one public key is provided.
	if len(args) == 0 {
		io.ErrPrintln("error: at least one public key must be provided\n")
		return flag.ErrHelp
	}

	// Remove the public keys from the authorized keys list.
	return manipulatesAuthorizedKeys(rootCfg, args, func(current []string, updates []string) []string {
		for _, publicKey := range updates {
			if index, found := slices.BinarySearch(current, publicKey); found {
				current = slices.Delete(current, index, index+1)
				io.Printfln("Public key %q removed from the authorized keys list.", publicKey)
			} else {
				io.Printfln("Public key %q not found in the authorized keys list.", publicKey)
			}
		}
		return current
	})
}

// newAuthAuthorizedListCmd creates the auth authorized list subcommand.
func newAuthAuthorizedListCmd(rootCfg *common.AuthFlags, io commands.IO) *commands.Command {
	cfg := &authRawFlags{
		auth: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "auth authorized list [flags]",
			ShortHelp:  "lists the client authorized keys",
			LongHelp:   "Lists the public keys of the clients that are authorized to authenticate with gnokms.",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execAuthAuthorizedList(cfg, io)
		},
	)
}

func execAuthAuthorizedList(authRawCfg *authRawFlags, io commands.IO) error {
	// Load the auth keys file.
	authKeysFile, err := loadAuthKeysFile(authRawCfg.auth)
	if err != nil {
		return err
	}

	// Print the authorized keys.
	keys := authKeysFile.ClientAuthorizedKeys

	// Handle empty list.
	if len(authKeysFile.ClientAuthorizedKeys) == 0 {
		if !authRawCfg.raw {
			io.Printfln("No authorized keys found.")
		}
		return nil
	}

	// Print keys based on output mode.
	if authRawCfg.raw {
		for _, key := range keys {
			io.Printfln("%s", key)
		}
	} else {
		io.Printfln("Authorized keys:")
		for _, key := range keys {
			io.Printfln("- %s", key)
		}
	}

	return nil
}
