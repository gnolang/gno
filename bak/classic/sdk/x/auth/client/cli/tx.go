package cli

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/classic/sdk/client"
	"github.com/tendermint/classic/sdk/x/auth/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Auth transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetMultiSignCommand(),
		GetSignCommand(),
	)
	return txCmd
}
