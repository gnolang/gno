package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

// newShowCmd creates the new secrets show command
func newShowCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "show",
			ShortUsage: "show [subcommand] [flags]",
			ShortHelp:  "Shows the Gno node secrets",
			LongHelp:   "Shows the Gno node secrets locally, including the validator key, validator state and node key",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newShowAllCmd(io),
		newShowSingleCmd(io),
	)

	return cmd
}

// readAndShowValidatorKey reads and shows the validator key from the given path
func readAndShowValidatorKey(path string, io commands.IO) error {
	validatorKey, err := readSecretData[privval.FilePVKey](path)
	if err != nil {
		return fmt.Errorf("unable to read validator key, %w", err)
	}

	w := tabwriter.NewWriter(io.Out(), 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintf(w, "[Validator Key Info]\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Address:\t%s\n", validatorKey.Address.String()); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Public Key:\t%s\n", validatorKey.PubKey.String()); err != nil {
		return err
	}

	return w.Flush()
}

// readAndShowValidatorState reads and shows the validator state from the given path
func readAndShowValidatorState(path string, io commands.IO) error {
	validatorState, err := readSecretData[privval.FilePVLastSignState](path)
	if err != nil {
		return fmt.Errorf("unable to read validator state, %w", err)
	}

	w := tabwriter.NewWriter(io.Out(), 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintf(w, "[Last Validator Sign State Info]\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Height:\t%d\n",
		validatorState.Height,
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Round:\t%d\n",
		validatorState.Round,
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Step:\t%d\n",
		validatorState.Step,
	); err != nil {
		return err
	}

	if validatorState.Signature != nil {
		if _, err := fmt.Fprintf(
			w,
			"Signature:\t%X\n",
			validatorState.Signature,
		); err != nil {
			return err
		}
	}

	if validatorState.SignBytes != nil {
		if _, err := fmt.Fprintf(
			w,
			"Sign Bytes:\t%X\n",
			validatorState.SignBytes,
		); err != nil {
			return err
		}
	}

	return w.Flush()
}

// readAndShowNodeKey reads and shows the node p2p key from the given path
func readAndShowNodeKey(path string, io commands.IO) error {
	nodeKey, err := readSecretData[p2p.NodeKey](path)
	if err != nil {
		return fmt.Errorf("unable to read node key, %w", err)
	}

	w := tabwriter.NewWriter(io.Out(), 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintf(w, "[Node P2P Info]\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Node ID:\t%s\n",
		nodeKey.ID(),
	); err != nil {
		return err
	}

	return w.Flush()
}
