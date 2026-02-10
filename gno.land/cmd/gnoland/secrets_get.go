package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

var errInvalidSecretsGetArgs = errors.New("invalid number of secrets get arguments provided")

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

	// Add subcommand helpers
	gen := commands.FieldsGenerator{
		MetaUpdate: func(meta *commands.Metadata, inputType string) {
			meta.ShortUsage = fmt.Sprintf("secrets get %s <%s>", meta.Name, inputType)
		},
		TagNameSelector: "json",
		TreeDisplay:     false,
	}
	cmd.AddSubCommands(gen.GenerateFrom(secrets{}, func(_ context.Context, args []string) error {
		return execSecretsGet(cfg, args, io)
	})...)

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
	// Make sure the directory is there
	if cfg.dataDir == "" || !isValidDirectory(cfg.dataDir) {
		return errInvalidDataDir
	}

	// Make sure the get arguments are valid
	if len(args) > 1 {
		return errInvalidSecretsGetArgs
	}

	// Load the secrets from the dir
	loadedSecrets, err := loadSecrets(cfg.dataDir)
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
func loadSecrets(dirPath string) (*secrets, error) {
	// Construct the file paths
	var (
		validatorKeyPath   = filepath.Join(dirPath, defaultValidatorKeyName)
		validatorStatePath = filepath.Join(dirPath, defaultValidatorStateName)
		nodeKeyPath        = filepath.Join(dirPath, defaultNodeKeyName)
	)

	var (
		vkInfo *validatorKeyInfo
		vsInfo *validatorStateInfo
		niInfo *nodeIDInfo

		err error
	)

	// Load the secrets
	if osm.FileExists(validatorKeyPath) {
		vkInfo, err = readValidatorKey(validatorKeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load secrets, %w", err)
		}
	}

	if osm.FileExists(validatorStatePath) {
		vsInfo, err = readValidatorState(validatorStatePath)
		if err != nil {
			return nil, fmt.Errorf("unable to load secrets, %w", err)
		}
	}

	if osm.FileExists(nodeKeyPath) {
		niInfo, err = readNodeID(nodeKeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load secrets, %w", err)
		}
	}

	return &secrets{
		ValidatorKeyInfo:   vkInfo,
		ValidatorStateInfo: vsInfo,
		NodeIDInfo:         niInfo,
	}, nil
}

// readValidatorKey reads the validator key from the given path
func readValidatorKey(path string) (*validatorKeyInfo, error) {
	validatorKey, err := signer.LoadFileKey(path)
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
	validatorState, err := fstate.LoadFileState(path)
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
func readNodeID(path string) (*nodeIDInfo, error) {
	nodeKey, err := types.LoadNodeKey(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read node key, %w", err)
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
			return nil, fmt.Errorf("unable to load config file, %w", err)
		}
	}

	return &nodeIDInfo{
		ID:         nodeKey.ID().String(),
		P2PAddress: constructP2PAddress(nodeKey.ID(), cfg.P2P.ListenAddress),
		PubKey:     nodeKey.PrivKey.PubKey().String(),
	}, nil
}

// constructP2PAddress constructs the P2P address other nodes can use
// to connect directly
func constructP2PAddress(nodeID types.ID, listenAddress string) string {
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
