package main

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newInitCmd creates the new secrets init command
func newInitCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [subcommand] [flags]",
			ShortHelp:  "Initializes the Gno node secrets",
			LongHelp:   "Initializes the Gno node secrets locally, including the validator key, validator state and node key",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newInitAllCmd(io),
		newInitSingleCmd(io),
	)

	return cmd
}

// initAndSaveValidatorKey generates a validator private key and saves it to the given path
func initAndSaveValidatorKey(path string, io commands.IO) error {
	// Initialize the validator's private key
	privateKey := generateValidatorPrivateKey()

	// Save the key
	if err := saveSecretData(privateKey, path); err != nil {
		return fmt.Errorf("unable to save validator key, %w", err)
	}

	io.Printfln("Validator private key saved at %s", path)

	return nil
}

// initAndSaveValidatorState generates an empty last validator sign state and saves it to the given path
func initAndSaveValidatorState(path string, io commands.IO) error {
	// Initialize the validator's last sign state
	validatorState := generateLastSignValidatorState()

	// Save the last sign state
	if err := saveSecretData(validatorState, path); err != nil {
		return fmt.Errorf("unable to save last validator sign state, %w", err)
	}

	io.Printfln("Validator last sign state saved at %s", path)

	return nil
}

// initAndSaveNodeKey generates a node p2p key and saves it to the given path
func initAndSaveNodeKey(path string, io commands.IO) error {
	// Initialize the node's p2p key
	nodeKey := generateNodeKey()

	// Save the node key
	if err := saveSecretData(nodeKey, path); err != nil {
		return fmt.Errorf("unable to save node p2p key, %w", err)
	}

	io.Printfln("Node key saved at %s", path)

	return nil
}
