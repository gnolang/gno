package main

import (
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/stretchr/testify/require"
)

const (
	withConfig int = iota
	withSecrets
)

func newTestHomeDirectory(t *testing.T, home string, args ...int) homeDirectory {
	t.Helper()

	homeDir := homeDirectory{
		homeDir:     home,
		genesisFile: home + "/genesis.json",
	}

	for _, arg := range args {
		switch arg {
		case withConfig:
			require.NoError(t, os.MkdirAll(homeDir.ConfigDir(), 0o755))
			require.NoError(t, config.WriteConfigFile(homeDir.ConfigFile(), config.DefaultConfig()))
		case withSecrets:
			require.NoError(t, os.MkdirAll(homeDir.SecretsDir(), 0o700))
		}
	}

	return homeDir
}

func TestHomeDirectoryPaths(t *testing.T) {
	dir := t.TempDir()

	h := homeDirectory{
		homeDir:     dir,
		genesisFile: dir + "/genesis.json",
	}

	tests := []struct {
		name     string
		f        func() string
		expected string
	}{
		{"Path", h.Path, dir},
		{"ConfigDir", h.ConfigDir, dir + "/config"},
		{"ConfigFile", h.ConfigFile, dir + "/config/config.toml"},
		{"GenesisFilePath", h.GenesisFilePath, h.genesisFile},
		{"SecretsDir", h.SecretsDir, dir + "/secrets"},
		{"SecretsNodeKey", h.SecretsNodeKey, dir + "/secrets/node_key.json"},
		{"SecretsValidatorKey", h.SecretsValidatorKey, dir + "/secrets/priv_validator_key.json"},
		{"SecretsValidatorState", h.SecretsValidatorState, dir + "/secrets/priv_validator_state.json"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.f())
		})
	}
}
