// Package gnokey implements a remote signer server using gnokey as the backend.
package gnokey

import (
	"context"
	"flag"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type gnokeyFlags struct {
	common.ServerFlags

	home                  string
	insecurePasswordStdin bool
}

var defaultGnokeyFlags = &gnokeyFlags{
	home:                  gnoenv.HomeDir(),
	insecurePasswordStdin: false,
}

// NewGnokeyCmd creates the gnokms gnokey subcommand.
func NewGnokeyCmd(io commands.IO) *commands.Command {
	gnFlags := &gnokeyFlags{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "gnokey",
			ShortUsage: "gnokey [flags] <key-name or address>",
			ShortHelp:  "uses gnokey as a remote signer",
			LongHelp:   "Runs a gnokms remote signer server using gnokey as backend.",
		},
		gnFlags,
		func(ctx context.Context, args []string) error {
			return execGnokey(ctx, args, gnFlags, io)
		},
	)
}

func (f *gnokeyFlags) RegisterFlags(fs *flag.FlagSet) {
	f.ServerFlags.RegisterFlags(fs)

	fs.StringVar(
		&f.home,
		"home",
		defaultGnokeyFlags.home,
		"gnokey home directory",
	)

	fs.BoolVar(
		&f.insecurePasswordStdin,
		"insecure-password-stdin",
		defaultGnokeyFlags.insecurePasswordStdin,
		"WARNING! take password from stdin",
	)
}

func execGnokey(ctx context.Context, args []string, gnFlags *gnokeyFlags, io commands.IO) error {
	// Key name must be provided.
	if len(args) != 1 {
		io.ErrPrintln("error: a key name must be provided\n")
		return flag.ErrHelp
	}
	keyName := args[0]

	// Initialize the gnokey signer with the provided key name.
	gnokeySigner, err := newGnokeySigner(gnFlags, keyName, io)
	if err != nil {
		return err
	}

	// Run the remote signer server with the gnokey signer.
	return common.RunSignerServer(ctx, &gnFlags.ServerFlags, gnokeySigner, io)
}
