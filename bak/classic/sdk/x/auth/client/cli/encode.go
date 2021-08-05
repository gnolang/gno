package cli

import (
	"encoding/base64"

	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino-x"

	"github.com/tendermint/classic/sdk/client/context"
	"github.com/tendermint/classic/sdk/client/flags"
	"github.com/tendermint/classic/sdk/x/auth/client/utils"
)

// txEncodeRespStr implements a simple Stringer wrapper for a encoded tx.
type txEncodeRespStr string

func (txr txEncodeRespStr) String() string {
	return string(txr)
}

// GetEncodeCommand returns the encode command to take a JSONified transaction and turn it into
// Amino-serialized bytes
func GetEncodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encode [file]",
		Short: "Encode transactions generated offline",
		Long: `Encode transactions created with the --generate-only flag and signed with the sign command.
Read a transaction from <file>, serialize it to the Amino wire protocol, and output it as base64.
If you supply a dash (-) argument in place of an input filename, the command reads from standard input.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliCtx := context.NewCLIContext()

			stdTx, err := utils.ReadStdTxFromFile(cliCtx.Codec, args[0])
			if err != nil {
				return
			}

			// re-encode it via the Amino wire protocol
			txBytes, err := cliCtx.Codec.MarshalSized(stdTx)
			if err != nil {
				return err
			}

			// base64 encode the encoded tx bytes
			txBytesBase64 := base64.StdEncoding.EncodeToString(txBytes)

			response := txEncodeRespStr(txBytesBase64)
			cliCtx.PrintOutput(response) // nolint:errcheck

			return nil
		},
	}

	return flags.PostCommands(cmd)[0]
}
