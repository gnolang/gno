package main

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

// newVerifyCmd creates the new secrets verify command
func newVerifyCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "verify",
			ShortUsage: "verify [subcommand] [flags]",
			ShortHelp:  "Verifies the Gno node secrets",
			LongHelp:   "Verifies the Gno node secrets locally, including the validator key, validator state and node key",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newVerifyAllCmd(io),
		newVerifySingleCmd(io),
	)

	return cmd
}

// readAndVerifyValidatorKey reads the validator key from the given path and verifies it
func readAndVerifyValidatorKey(path string, io commands.IO) (*privval.FilePVKey, error) {
	validatorKey, err := readSecretData[privval.FilePVKey](path)
	if err != nil {
		return nil, fmt.Errorf("unable to read validator key, %w", err)
	}

	if err := validateValidatorKey(validatorKey); err != nil {
		return nil, err
	}

	io.Printfln("Validator Private Key at %s is valid", path)

	return validatorKey, nil
}

// readAndVerifyValidatorState reads the validator state from the given path and verifies it
func readAndVerifyValidatorState(path string, io commands.IO) (*privval.FilePVLastSignState, error) {
	validatorState, err := readSecretData[privval.FilePVLastSignState](path)
	if err != nil {
		return nil, fmt.Errorf("unable to read last validator sign state, %w", err)
	}

	if err := validateValidatorState(validatorState); err != nil {
		return nil, err
	}

	io.Printfln("Last Validator Sign state at %s is valid", path)

	return validatorState, nil
}

// readAndVerifyNodeKey reads the node p2p key from the given path and verifies it
func readAndVerifyNodeKey(path string, io commands.IO) error {
	nodeKey, err := readSecretData[p2p.NodeKey](path)
	if err != nil {
		return fmt.Errorf("unable to read node p2p key, %w", err)
	}

	if err := validateNodeKey(nodeKey); err != nil {
		return err
	}

	io.Printfln("Node P2P key at %s is valid", path)

	return nil
}
