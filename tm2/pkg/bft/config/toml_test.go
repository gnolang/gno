package config

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
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
		p := filepath.Join(rootDir, f)
		_, err := os.Stat(p)
		assert.Nil(t, err, p)
	}
}

func TestEnsureRoot(t *testing.T) {
	t.Parallel()

	// setup temp dir for test
	tmpDir := t.TempDir()

	// create root dir
	throwaway := DefaultConfig()
	throwaway.SetRootDir(tmpDir)
	require.NoError(t, throwaway.EnsureDirs())

	configPath := filepath.Join(tmpDir, defaultConfigPath)
	require.NoError(t, WriteConfigFile(configPath, throwaway))

	// make sure config is set properly
	data, err := os.ReadFile(filepath.Join(tmpDir, defaultConfigPath))
	require.Nil(t, err)

	require.True(t, checkConfig(string(data)))

	ensureFiles(t, tmpDir, DefaultDBDir)
}

func TestEnsureTestRoot(t *testing.T) {
	t.Parallel()

	testName := "ensureTestRoot"

	// create root dir
	cfg, _ := ResetTestRoot(testName)
	defer os.RemoveAll(cfg.RootDir)
	rootDir := cfg.RootDir

	// make sure config is set properly
	data, err := os.ReadFile(filepath.Join(rootDir, defaultConfigPath))
	require.Nil(t, err)

	require.True(t, checkConfig(string(data)))

	// TODO: make sure the cfg returned and testconfig are the same!
	ensureFiles(
		t,
		rootDir,
		"genesis.json",
		DefaultDBDir,
	)

	// Root dir was set directly in validator config along with the DefaultSecretsDir.
	ensureFiles(
		t,
		"",
		cfg.Consensus.PrivValidator.LocalSignerPath(),
		cfg.Consensus.PrivValidator.SignStatePath(),
	)
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

		assert.EqualValues(t, defaultConfig.BaseConfig, cfg.BaseConfig)
		assert.EqualValues(t, defaultConfig.RPC, cfg.RPC)
		assert.EqualValues(t, defaultConfig.P2P, cfg.P2P)
		assert.EqualValues(t, defaultConfig.Mempool, cfg.Mempool)
		assert.EqualValues(t, defaultConfig.Consensus, cfg.Consensus)
		assert.Equal(t, defaultConfig.TxEventStore.EventStoreType, cfg.TxEventStore.EventStoreType)
		assert.Empty(t, defaultConfig.TxEventStore.Params, cfg.TxEventStore.Params)
	})
}

func TestTOML_ConfigComments(t *testing.T) {
	t.Parallel()

	collectCommentTags := func(v reflect.Value) []string {
		var (
			comments    = make([]string, 0)
			structStack = []reflect.Value{v}
		)

		// Descend on and parse all child fields
		for len(structStack) > 0 {
			structVal := structStack[len(structStack)-1]
			structStack = structStack[:len(structStack)-1]

			// Process all fields of the struct
			for i := range structVal.NumField() {
				fieldVal := structVal.Field(i)
				fieldType := structVal.Type().Field(i)

				// If the field is a struct, push it onto the stack for further processing
				if fieldVal.Kind() == reflect.Struct {
					structStack = append(structStack, fieldVal)

					continue
				}

				// Collect the comment tag value from the field
				if commentTag := fieldType.Tag.Get("comment"); commentTag != "" {
					comments = append(comments, commentTag)
				}
			}
		}

		return comments
	}

	cleanComments := func(original string) string {
		return strings.ReplaceAll(original, "#", "")
	}

	// Create test config file
	configFile, cleanup := testutils.NewTestFile(t)
	t.Cleanup(cleanup)

	// Create the default config
	defaultConfig := DefaultConfig()

	// Write valid TOML
	require.NoError(t, WriteConfigFile(configFile.Name(), defaultConfig))

	// Collect config comments
	comments := collectCommentTags(reflect.ValueOf(*defaultConfig))
	require.NotEmpty(t, comments)

	// Read the entire config file
	rawConfig, err := io.ReadAll(configFile)
	require.NoError(t, err)

	// Verify TOML comments are present
	content := cleanComments(string(rawConfig))
	for _, comment := range comments {
		assert.Contains(t, content, cleanComments(comment))
	}
}
