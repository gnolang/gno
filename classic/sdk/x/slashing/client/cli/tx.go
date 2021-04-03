package cli

import (
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino-x"

	"github.com/tendermint/classic/sdk/client"
	"github.com/tendermint/classic/sdk/client/context"
	sdk "github.com/tendermint/classic/sdk/types"
	"github.com/tendermint/classic/sdk/x/auth"
	"github.com/tendermint/classic/sdk/x/auth/client/utils"
	"github.com/tendermint/classic/sdk/x/slashing/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	slashingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Slashing transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	slashingTxCmd.AddCommand(client.PostCommands(
		GetCmdUnjail(),
	)...)

	return slashingTxCmd
}

// GetCmdUnjail implements the create unjail validator command.
func GetCmdUnjail() *cobra.Command {
	return &cobra.Command{
		Use:   "unjail",
		Args:  cobra.NoArgs,
		Short: "unjail validator previously jailed for downtime",
		Long: `unjail a jailed validator:

$ <appcli> tx slashing unjail --from mykey
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder())
			cliCtx := context.NewCLIContext()

			valAddr := cliCtx.GetFromAddress()

			msg := types.NewMsgUnjail(sdk.ValAddress(valAddr))
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
