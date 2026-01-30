package keyscli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeCLAHash(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "simple content",
			content:  "I agree to the CLA.\n",
			expected: "b6faa56f8eec79eb", // printf 'I agree to the CLA.\n' | sha256sum | cut -c1-16
		},
		{
			name:     "test CLA content",
			content:  "Test CLA content\n",
			expected: "a3d74e2544d091e8", // echo "Test CLA content" | sha256sum | cut -c1-16
		},
		{
			name:     "empty content",
			content:  "",
			expected: "e3b0c44298fc1c14", // printf '' | sha256sum | cut -c1-16
		},
		{
			name:     "multiline content",
			content:  "Line 1\nLine 2\nLine 3\n",
			expected: "ec1f9796b88620f6", // printf 'Line 1\nLine 2\nLine 3\n' | sha256sum | cut -c1-16
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeCLAHash(tt.content)
			assert.Equal(t, tt.expected, result)
			assert.Len(t, result, 16, "hash should be 16 hex characters")
		})
	}
}

func TestFetchCLAContent(t *testing.T) {
	// Create temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_cla.txt")
	testContent := "This is test CLA content.\n"
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0o644))

	tests := []struct {
		name        string
		urlOrPath   string
		expected    string
		expectError bool
	}{
		{
			name:      "local file path",
			urlOrPath: testFile,
			expected:  testContent,
		},
		{
			name:      "file:// URL",
			urlOrPath: "file://" + testFile,
			expected:  testContent,
		},
		{
			name:        "nonexistent file",
			urlOrPath:   "/nonexistent/path/to/file.txt",
			expectError: true,
		},
		{
			name:        "nonexistent file:// URL",
			urlOrPath:   "file:///nonexistent/path/to/file.txt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FetchCLAContent(tt.urlOrPath)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("load empty config", func(t *testing.T) {
		cfg, err := LoadConfig(tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, cfg.Zones)
		assert.Empty(t, cfg.Zones)
	})

	t.Run("save and load config", func(t *testing.T) {
		cfg := &Config{
			Zones: map[string]ZoneConfig{
				"https://rpc.gno.land:443": {
					CLAHash: "b6faa56f8eec79eb",
				},
				"http://localhost:26657": {
					CLAHash: "a3d74e2544d091e8",
				},
			},
		}

		require.NoError(t, SaveConfig(tmpDir, cfg))

		// Verify file was created with correct permissions
		configPath := filepath.Join(tmpDir, configFile)
		info, err := os.Stat(configPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

		// Load and verify
		loaded, err := LoadConfig(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "b6faa56f8eec79eb", loaded.GetCLAHash("https://rpc.gno.land:443"))
		assert.Equal(t, "a3d74e2544d091e8", loaded.GetCLAHash("http://localhost:26657"))
		assert.Equal(t, "", loaded.GetCLAHash("https://unknown.chain:443"))
	})

	t.Run("set CLA hash", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "set_hash")
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		cfg, err := LoadConfig(subDir)
		require.NoError(t, err)

		cfg.SetCLAHash("https://rpc.test.gno.land:443", "1234567890abcdef")
		require.NoError(t, SaveConfig(subDir, cfg))

		loaded, err := LoadConfig(subDir)
		require.NoError(t, err)
		assert.Equal(t, "1234567890abcdef", loaded.GetCLAHash("https://rpc.test.gno.land:443"))
	})
}

func TestLoadCLAHashForRemote(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up config with multiple remotes
	cfg := &Config{
		Zones: map[string]ZoneConfig{
			"https://rpc.gno.land:443": {
				CLAHash: "b6faa56f8eec79eb",
			},
			"http://localhost:26657": {
				CLAHash: "a3d74e2544d091e8",
			},
		},
	}
	require.NoError(t, SaveConfig(tmpDir, cfg))

	t.Run("load existing remote", func(t *testing.T) {
		hash := LoadCLAHashForRemote(tmpDir, "https://rpc.gno.land:443")
		assert.Equal(t, "b6faa56f8eec79eb", hash)
	})

	t.Run("load another remote", func(t *testing.T) {
		hash := LoadCLAHashForRemote(tmpDir, "http://localhost:26657")
		assert.Equal(t, "a3d74e2544d091e8", hash)
	})

	t.Run("load unknown remote", func(t *testing.T) {
		hash := LoadCLAHashForRemote(tmpDir, "https://unknown:443")
		assert.Equal(t, "", hash)
	})

	t.Run("load from nonexistent dir", func(t *testing.T) {
		hash := LoadCLAHashForRemote("/nonexistent/dir", "https://rpc.gno.land:443")
		assert.Equal(t, "", hash)
	})
}
