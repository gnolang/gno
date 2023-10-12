package config

import (
	"os"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ensureFiles(t *testing.T, rootDir string, files ...string) {
	t.Helper()

	for _, f := range files {
		p := join(f, rootDir)
		_, err := os.Stat(p)
		assert.Nil(t, err, p)
	}
}

func TestEnsureRoot(t *testing.T) {
	require := require.New(t)

	// setup temp dir for test
	tmpDir := t.TempDir()

	// create root dir
	throwaway := DefaultConfig()
	throwaway.SetRootDir(tmpDir)
	throwaway.EnsureDirs()
	configPath := join(tmpDir, defaultConfigFilePath)
	WriteConfigFile(configPath, throwaway)

	// make sure config is set properly
	data, err := os.ReadFile(join(tmpDir, defaultConfigFilePath))
	require.Nil(err)

	if !checkConfig(string(data)) {
		t.Fatalf("config file missing some information")
	}

	ensureFiles(t, tmpDir, "data")
}

func TestEnsureTestRoot(t *testing.T) {
	require := require.New(t)

	testName := "ensureTestRoot"

	// create root dir
	cfg := ResetTestRoot(testName)
	defer os.RemoveAll(cfg.RootDir)
	rootDir := cfg.RootDir

	// make sure config is set properly
	data, err := os.ReadFile(join(rootDir, defaultConfigFilePath))
	require.Nil(err)

	if !checkConfig(string(data)) {
		t.Fatalf("config file missing some information")
	}

	// TODO: make sure the cfg returned and testconfig are the same!
	baseConfig := DefaultBaseConfig()
	ensureFiles(t, rootDir, defaultDataDir, baseConfig.Genesis, baseConfig.PrivValidatorKey, baseConfig.PrivValidatorState)
}

func checkConfig(configFile string) bool {
	var valid bool

	// list of words we expect in the config
	elems := []string{
		"moniker",
		"seeds",
		"proxy_app",
		"fast_sync",
		"create_empty_blocks",
		"peer",
		"timeout",
		"broadcast",
		"send",
		"addr",
		"wal",
		"propose",
		"max",
		"genesis",
	}
	for _, e := range elems {
		if !strings.Contains(configFile, e) {
			valid = false
		} else {
			valid = true
		}
	}
	return valid
}

func TestTOML_LoadConfig(t *testing.T) {
	t.Parallel()

	t.Run("config does not exist", func(t *testing.T) {
		t.Parallel()

		cfg, loadErr := LoadConfigFile("dummy-path")

		assert.Error(t, loadErr)
		assert.Nil(t, cfg)
	})

	t.Run("config is not valid toml", func(t *testing.T) {
		t.Parallel()

		// Create config file
		configFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		// Write invalid TOML
		_, writeErr := configFile.WriteString("invalid TOML")
		require.NoError(t, writeErr)

		cfg, loadErr := LoadConfigFile(configFile.Name())

		assert.Error(t, loadErr)
		assert.Nil(t, cfg)
	})

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		// Create config file
		configFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		// Create the default config
		defaultConfig := DefaultConfig()

		// Marshal the default config
		defaultConfigRaw, marshalErr := toml.Marshal(defaultConfig)
		require.NoError(t, marshalErr)

		// Write valid TOML
		_, writeErr := configFile.Write(defaultConfigRaw)
		require.NoError(t, writeErr)

		cfg, loadErr := LoadConfigFile(configFile.Name())

		require.NoError(t, loadErr)
		assert.Equal(t, defaultConfig, cfg)
	})
}
