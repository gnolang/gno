// Package auth provides the 'gnokms auth' command and its subcommands for managing mutual authentication keys.
package auth

import (
	"fmt"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// loadAuthKeysFile loads the auth keys file and checks if it is valid.
func loadAuthKeysFile(rootCfg *common.AuthFlags) (*common.AuthKeysFile, error) {
	// Check if the file exists.
	if !osm.FileExists(rootCfg.AuthKeysFile) {
		return nil, fmt.Errorf("%s: %s\n%s",
			"error: auth keys file does not exist at path", rootCfg.AuthKeysFile,
			"use 'gnokms auth generate' to create a new one",
		)
	}

	// Check if the file is valid.
	authKeysFile, err := common.LoadAuthKeysFile(rootCfg.AuthKeysFile)
	if err != nil {
		return nil, fmt.Errorf("%s: %s\n%s: %w\n%s",
			"error: auth keys file is invalid at path", rootCfg.AuthKeysFile,
			"unable to load", err,
			"use 'gnokms auth generate -overwrite' to create a new one",
		)
	}

	return authKeysFile, nil
}

// NewAuthCmd creates the gnokms auth subcommand.
func NewAuthCmd(io commands.IO) *commands.Command {
	rootCfg := &common.AuthFlags{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "auth",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "manages mutual authentication keys",
			LongHelp:   "Manages the mutual authentication keys used by gnokms including gnokms' own private key and client authorized public keys.",
		},
		rootCfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newAuthAuthorizedCmd(rootCfg, io),
		newAuthGenerateCmd(rootCfg, io),
		newAuthIdentityCmd(rootCfg, io),
	)

	return cmd
}
