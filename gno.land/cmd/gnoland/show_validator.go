package main

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// Display a node's validator info.
func newValidatorCmd(bc baseCfg) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "validator",
			ShortUsage: "validator",
			ShortHelp:  "Show node's validator info",
		},
		nil,
		func(_ context.Context, args []string) error {
			return execShowValidator(bc)
		},
	)
	return cmd
}

func execShowValidator(bc baseCfg) error {
	config := bc.tmConfig
	keyFilePath := config.PrivValidatorKeyFile()
	stateFilePath := config.PrivValidatorStateFile()
	pv := privval.LoadFilePV(keyFilePath, stateFilePath)
	pubKey := pv.GetPubKey()
	bz := amino.MustMarshalJSON(pubKey)

	fmt.Printf("Address: \"%v\"\n", pubKey.Address())
	fmt.Printf(" Pubkey: %v\n", string(bz))
	return nil
}
