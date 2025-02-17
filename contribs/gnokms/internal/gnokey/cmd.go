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
			ShortUsage: "gnokey [flags] <key-name or address>",
			ShortHelp:  "use gnokey as a remote signer",
			LongHelp:   "Run a gnokms remote signer server using gnokey as backend.",
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

	// Initialize the gnokey signer with the provided key name.
	gnokeySigner, err := newGnokeySigner(io, gnFlags, keyName)
	if err != nil {
		return err
	}

	// Run the remote signer server with the gnokey signer.
	return common.RunSignerServer(io, commonFlags, gnokeySigner)
}
