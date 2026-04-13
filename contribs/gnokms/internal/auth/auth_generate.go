package auth

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

type authGenerateFlags struct {
	auth      *common.AuthFlags
	overwrite bool
}

var defaultAuthGenerateFlags = &authGenerateFlags{
	overwrite: false,
}

// newAuthGenerateCmd creates the auth generate subcommand.
func newAuthGenerateCmd(rootCfg *common.AuthFlags, io commands.IO) *commands.Command {
	cfg := &authGenerateFlags{
		auth: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "generate",
			ShortUsage: "auth generate [flags]",
			ShortHelp:  "generates a new file with mutual authentication keys",
			LongHelp:   "Generates a new file with mutual authentication keys including gnokms' own private key and an empty list of client authorized public keys.",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execAuthGenerate(cfg, io)
		},
	)
}

func (f *authGenerateFlags) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&f.overwrite,
		"overwrite",
		defaultAuthGenerateFlags.overwrite,
		"overwrite the keys file if it already exists",
	)
}

func execAuthGenerate(authGenCfg *authGenerateFlags, io commands.IO) error {
	// Check if the file already exists.
	if osm.FileExists(authGenCfg.auth.AuthKeysFile) && !authGenCfg.overwrite {
		return fmt.Errorf("%s: %s\n%s",
			"error: auth keys file already exists at path", authGenCfg.auth.AuthKeysFile,
			"use 'gnokms auth generate -overwrite' to overwrite it",
		)
	}

	// Generate a new auth keys file.
	if _, err := common.GeneratePersistedAuthKeysFile(authGenCfg.auth.AuthKeysFile); err != nil {
		return fmt.Errorf("error generating auth keys file: %w", err)
	}

	// Print the path to the generated file.
	io.Printfln("Generated auth keys file at path: %q", authGenCfg.auth.AuthKeysFile)

	return nil
}
