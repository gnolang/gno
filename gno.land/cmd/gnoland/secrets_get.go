package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

type secretsGetCfg struct {
	commonAllCfg

	json bool
}

// newSecretsGetCmd creates the secrets get command
func newSecretsGetCmd(io commands.IO) *commands.Command {
	cfg := &secretsGetCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "get",
			ShortUsage: "secrets get [flags] [<key>]",
			ShortHelp:  "shows all Gno secrets present in a common directory",
			LongHelp: fmt.Sprintf(
				"shows the validator private key, the node p2p key and the validator's last sign state. "+
					"If a key is provided, it shows the specified key value. Available keys: %s",
				getAvailableSecretsKeys(),
			),
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSecretsGet(cfg, args, io)
		},
	)

	return cmd
}

func (c *secretsGetCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonAllCfg.RegisterFlags(fs)

	fs.BoolVar(
		&c.json,
		"json",
		false,
		"flag indicating if the secret output should be in JSON",
	)
}

func execSecretsGet(cfg *secretsGetCfg, args []string, io commands.IO) error {
	// Make sure the directory is there
	if cfg.dataDir == "" || !isValidDirectory(cfg.dataDir) {
		return errInvalidDataDir
	}

	// Verify the secrets key
	if err := verifySecretsKey(args); err != nil {
		return err
	}

	var key string

	if len(args) > 0 {
		key = args[0]
	}

	// Construct the paths
	var (
		validatorKeyPath   = filepath.Join(cfg.dataDir, defaultValidatorKeyName)
		validatorStatePath = filepath.Join(cfg.dataDir, defaultValidatorStateName)
		nodeKeyPath        = filepath.Join(cfg.dataDir, defaultNodeKeyName)
	)

	// Construct the display methods
	var (
		displayVK = wrapDisplayFn(io, outputTerminalVK)
		displayVS = wrapDisplayFn(io, outputTerminalVS)
		displayNK = wrapDisplayFn(io, outputTerminalNK)
	)

	if cfg.json {
		displayVK = wrapDisplayFn(io, outputJSONCommon[validatorKeyInfo])
		displayVS = wrapDisplayFn(io, outputJSONCommon[validatorStateInfo])
		displayNK = wrapDisplayFn(io, outputJSONCommon[nodeKeyInfo])
	}

	switch key {
	case validatorPrivateKeyKey:
		// Show the validator's key info
		return readAndShowValidatorKey(validatorKeyPath, displayVK)
	case validatorStateKey:
		// Show the validator's last sign state
		return readAndShowValidatorState(validatorStatePath, displayVS)
	case nodeIDKey:
		// Show the node's p2p info
		return readAndShowNodeKey(nodeKeyPath, displayNK)
	default:
		// Show the node's p2p info
		if err := readAndShowNodeKey(nodeKeyPath, displayNK); err != nil {
			return err
		}

		// Show the validator's key info
		if err := readAndShowValidatorKey(validatorKeyPath, displayVK); err != nil {
			return err
		}

		// Show the validator's last sign state
		return readAndShowValidatorState(validatorStatePath, displayVS)
	}
}

type (
	validatorKeyInfo struct {
		Address string `json:"address"`
		PubKey  string `json:"pub_key"`
	}

	validatorStateInfo struct {
		Height int64 `json:"height"`
		Round  int   `json:"round"`
		Step   int8  `json:"step"`

		Signature []byte `json:"signature"`
		SignBytes []byte `json:"sign_bytes"`
	}

	nodeKeyInfo struct {
		NodeID     string `json:"node_id"`
		P2PAddress string `json:"p2p_address"`
	}
)

func (v validatorStateInfo) MarshalJSON() ([]byte, error) {
	type original validatorStateInfo

	return json.Marshal(&struct {
		Signature string `json:"signature,omitempty"`
		SignBytes string `json:"sign_bytes,omitempty"`
		original
	}{
		Signature: fmt.Sprintf("%X", v.Signature),
		SignBytes: fmt.Sprintf("%X", v.SignBytes),
		original:  (original)(v),
	})
}

type (
	secretDisplayType interface {
		validatorKeyInfo | validatorStateInfo | nodeKeyInfo
	}

	displayFn[T secretDisplayType] func(input T) error
	outputFn[T secretDisplayType]  func(input T, io commands.IO) error
)

// wrapDisplayFn wraps the display function to output to the specific IO
func wrapDisplayFn[T secretDisplayType](
	io commands.IO,
	outputFn outputFn[T],
) displayFn[T] {
	return func(input T) error {
		return outputFn(input, io)
	}
}

// readAndShowValidatorKey reads and shows the validator key from the given path
func readAndShowValidatorKey(
	path string,
	displayFn displayFn[validatorKeyInfo],
) error {
	validatorKey, err := readSecretData[privval.FilePVKey](path)
	if err != nil {
		return fmt.Errorf("unable to read validator key, %w", err)
	}

	info := validatorKeyInfo{
		Address: validatorKey.Address.String(),
		PubKey:  validatorKey.PubKey.String(),
	}

	// Print the output
	return displayFn(info)
}

// outputTerminalVK outputs the validator key info raw to the terminal.
// TODO we should consider ditching this "structured" terminal output, and having
// similar key-value output we have for 'gnoland config get'
func outputTerminalVK(info validatorKeyInfo, io commands.IO) error {
	w := tabwriter.NewWriter(io.Out(), 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintf(w, "[Validator Key Info]\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Address:\t%s\n", info.Address); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Public Key:\t%s\n", info.PubKey); err != nil {
		return err
	}

	return w.Flush()
}

// readAndShowValidatorState reads and shows the validator state from the given path
func readAndShowValidatorState(path string, displayFn displayFn[validatorStateInfo]) error {
	validatorState, err := readSecretData[privval.FilePVLastSignState](path)
	if err != nil {
		return fmt.Errorf("unable to read validator state, %w", err)
	}

	info := validatorStateInfo{
		Height:    validatorState.Height,
		Round:     validatorState.Round,
		Step:      validatorState.Step,
		Signature: validatorState.Signature,
		SignBytes: validatorState.SignBytes,
	}

	// Print the output
	return displayFn(info)
}

// outputTerminalVS outputs the validator state info raw to the terminal
// TODO we should consider ditching this "structured" terminal output, and having
// similar key-value output we have for 'gnoland config get'
func outputTerminalVS(info validatorStateInfo, io commands.IO) error {
	w := tabwriter.NewWriter(io.Out(), 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintf(w, "[Last Validator Sign State Info]\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Height:\t%d\n",
		info.Height,
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Round:\t%d\n",
		info.Round,
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		w,
		"Step:\t%d\n",
		info.Step,
	); err != nil {
		return err
	}

	if info.Signature != nil {
		if _, err := fmt.Fprintf(
			w,
			"Signature:\t%X\n",
			info.Signature,
		); err != nil {
			return err
		}
	}

	if info.SignBytes != nil {
		if _, err := fmt.Fprintf(
			w,
			"Sign Bytes:\t%X\n",
			info.SignBytes,
		); err != nil {
			return err
		}
	}

	return w.Flush()
}

// readAndShowNodeKey reads and shows the node p2p key from the given path
func readAndShowNodeKey(path string, displayFn displayFn[nodeKeyInfo]) error {
	nodeKey, err := readSecretData[p2p.NodeKey](path)
	if err != nil {
		return fmt.Errorf("unable to read node key, %w", err)
	}

	// Construct the config path
	var (
		nodeDir    = filepath.Join(filepath.Dir(path), "..")
		configPath = constructConfigPath(nodeDir)

		cfg = config.DefaultConfig()
	)

	// Check if there is an existing config file
	if osm.FileExists(configPath) {
		// Attempt to grab the config from disk
		cfg, err = config.LoadConfig(nodeDir)
		if err != nil {
			return fmt.Errorf("unable to load config file, %w", err)
		}
	}

	info := nodeKeyInfo{
		NodeID:     nodeKey.ID().String(),
		P2PAddress: cfg.P2P.ListenAddress,
	}

	// Print the output
	return displayFn(info)
}

// outputTerminalNK outputs the node key info raw to the terminal
// TODO we should consider ditching this "structured" terminal output, and having
// similar key-value output we have for 'gnoland config get'
func outputTerminalNK(info nodeKeyInfo, io commands.IO) error {
	w := tabwriter.NewWriter(io.Out(), 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintf(w, "[Node P2P Info]\n\n"); err != nil {
		return err
	}

	// Print the ID info
	if _, err := fmt.Fprintf(
		w,
		"Node ID:\t%s\n",
		info.NodeID,
	); err != nil {
		return err
	}

	// Print the P2P address info
	if _, err := fmt.Fprintf(
		w,
		"P2P Address:\t%s\n",
		constructP2PAddress(
			info.NodeID,
			info.P2PAddress,
		),
	); err != nil {
		return err
	}

	return w.Flush()
}

// outputJSONCommon outputs the given secrets to JSON
func outputJSONCommon[T secretDisplayType](
	input T,
	io commands.IO,
) error {
	encoded, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("unable to marshal JSON, %w", err)
	}

	io.Println(string(encoded))

	return nil
}

// constructP2PAddress constructs the P2P address other nodes can use
// to connect directly
func constructP2PAddress(nodeID, listenAddress string) string {
	var (
		address string
		parts   = strings.SplitN(listenAddress, "://", 2)
	)

	switch len(parts) {
	case 2:
		address = parts[1]
	default:
		address = listenAddress
	}

	return fmt.Sprintf("%s@%s", nodeID, address)
}
