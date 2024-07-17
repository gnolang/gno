package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

type secretsGetCfg struct {
	commonAllCfg

	raw bool
}

// newSecretsGetCmd creates the secrets get command
func newSecretsGetCmd(io commands.IO) *commands.Command {
	cfg := &secretsGetCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "get",
			ShortUsage: "secrets get [flags] [<key>]",
			ShortHelp:  "shows the Gno secrets present in a common directory",
			LongHelp: "shows the validator private key, the node p2p key and the validator's last sign state at the given path " +
				"by fetching the option specified at <key>",
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
		&c.raw,
		"raw",
		false,
		"output raw string values, rather than as JSON strings",
	)
}

func execSecretsGet(cfg *secretsGetCfg, args []string, io commands.IO) error {
	// Load the secrets from the dir
	loadedSecrets, err := loadSecrets(cfg.homeDir)
	if err != nil {
		return err
	}

	// Find and print the secrets value, if any
	if err := printKeyValue(loadedSecrets, cfg.raw, io, args...); err != nil {
		return fmt.Errorf("unable to get secrets value, %w", err)
	}

	return nil
}

// loadSecrets loads the secrets from the specified data directory
func loadSecrets(homeDir homeDirectory) (*secrets, error) {
	var (
		s   *secrets = &secrets{}
		err error
	)

	if osm.FileExists(homeDir.SecretsValidatorKey()) {
		s.ValidatorKeyInfo, err = readValidatorKey(homeDir.SecretsValidatorKey())
		if err != nil {
			return nil, fmt.Errorf("unable to load secrets, %w", err)
		}
	}

	if osm.FileExists(homeDir.SecretsValidatorState()) {
		s.ValidatorStateInfo, err = readValidatorState(homeDir.SecretsValidatorState())
		if err != nil {
			return nil, fmt.Errorf("unable to load secrets, %w", err)
		}
	}

	if osm.FileExists(homeDir.SecretsNodeKey()) {
		s.NodeIDInfo, err = readNodeID(homeDir)
		if err != nil {
			return nil, fmt.Errorf("unable to load secrets, %w", err)
		}
	}

	return s, nil
}

// readValidatorKey reads the validator key from the given path
func readValidatorKey(path string) (*validatorKeyInfo, error) {
	validatorKey, err := readSecretData[privval.FilePVKey](path)
	if err != nil {
		return nil, fmt.Errorf("unable to read validator key, %w", err)
	}

	return &validatorKeyInfo{
		Address: validatorKey.Address.String(),
		PubKey:  validatorKey.PubKey.String(),
	}, nil
}

// readValidatorState reads the validator state from the given path
func readValidatorState(path string) (*validatorStateInfo, error) {
	validatorState, err := readSecretData[privval.FilePVLastSignState](path)
	if err != nil {
		return nil, fmt.Errorf("unable to read validator state, %w", err)
	}

	return &validatorStateInfo{
		Height:    validatorState.Height,
		Round:     validatorState.Round,
		Step:      validatorState.Step,
		Signature: validatorState.Signature,
		SignBytes: validatorState.SignBytes,
	}, nil
}

// readNodeID reads the node p2p info from the given path
func readNodeID(homeDir homeDirectory) (*nodeIDInfo, error) {
	cfg := config.DefaultConfig()

	nodeKey, err := readSecretData[p2p.NodeKey](homeDir.SecretsNodeKey())
	if err != nil {
		return nil, fmt.Errorf("unable to read node key, %w", err)
	}

	if osm.FileExists(homeDir.ConfigFile()) {
		cfg, err = config.LoadConfig(homeDir.Path())
		if err != nil {
			return nil, fmt.Errorf("unable to load config file, %w", err)
		}
	}

	return &nodeIDInfo{
		ID:         nodeKey.ID().String(),
		P2PAddress: constructP2PAddress(nodeKey.ID(), cfg.P2P.ListenAddress),
	}, nil
}

// constructP2PAddress constructs the P2P address other nodes can use
// to connect directly
func constructP2PAddress(nodeID p2p.ID, listenAddress string) string {
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
