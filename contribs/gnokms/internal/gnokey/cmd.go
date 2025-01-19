package gnokey

import (
	"context"
	"flag"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func NewGnokeyCmd(io commands.IO, commonFlags *common.Flags) *commands.Command {
	gnFlags := &gnokeyFlags{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "gnokey",
			ShortUsage: "gnokms gnokey <key-name> [flags]",
			ShortHelp:  "use gnokey as a remote signer",
			LongHelp:   "This tool allows to use gnokey as a remote signer for gnokms.",
		},
		gnFlags,
		func(_ context.Context, args []string) error {
			return exec(io, args, gnFlags, commonFlags)
		},
	)
}

func exec(io commands.IO, args []string, gnFlags *gnokeyFlags, commonFlags *common.Flags) error {
	// Key name must be provided.
	if len(args) != 1 {
		io.ErrPrintln("error: a key name must be provided\n")
		return flag.ErrHelp
	}
	keyName := args[0]

	// Initialize the gnokey private validator with the provided key name.
	privVal, err := newGnokeyPrivVal(io, gnFlags, keyName)
	if err != nil {
		return err
	}

	// Run the remote signer server with the gnokey private validator.
	return common.RunSignerServer(io, commonFlags, privVal)
}
