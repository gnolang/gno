// Package ledger implements a remote signer server using a Ledger Tendermint
// validator app as backend.
package ledger

import (
	"context"
	"flag"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type ledgerFlags struct {
	common.ServerFlags
}

// NewLedgerCmd creates the gnokms ledger subcommand.
func NewLedgerCmd(io commands.IO) *commands.Command {
	ledgerFlags := &ledgerFlags{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "ledger",
			ShortUsage: "ledger [flags]",
			ShortHelp:  "uses a Ledger Tendermint validator app as a remote signer",
			LongHelp:   "Runs a gnokms remote signer server using a Ledger Tendermint validator app as backend.",
		},
		ledgerFlags,
		func(ctx context.Context, args []string) error {
			return execLedger(ctx, args, ledgerFlags, io)
		},
	)
}

func (f *ledgerFlags) RegisterFlags(fs *flag.FlagSet) {
	f.ServerFlags.RegisterFlags(fs)
}

func execLedger(ctx context.Context, args []string, ledgerFlags *ledgerFlags, io commands.IO) error {
	if len(args) != 0 {
		io.ErrPrintln("error: unexpected arguments\n")
		return flag.ErrHelp
	}

	ledgerSigner, err := newLedgerSigner()
	if err != nil {
		return err
	}

	return common.RunSignerServer(ctx, &ledgerFlags.ServerFlags, ledgerSigner, io)
}
