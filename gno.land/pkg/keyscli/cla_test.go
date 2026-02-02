package keyscli

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCLACmd(t *testing.T) {
	cfg := &client.BaseCfg{}
	io := commands.NewTestIO()

	cmd := NewCLACmd(cfg, io)
	assert.NotNil(t, cmd)
}

func TestCLASignCfg_RegisterFlags(t *testing.T) {
	cfg := &CLASignCfg{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg.RegisterFlags(fs)

	// Check that --url flag is registered
	urlFlag := fs.Lookup("url")
	assert.NotNil(t, urlFlag)
	assert.Equal(t, "", urlFlag.DefValue)
}

func TestCLAStatusCfg_RegisterFlags(t *testing.T) {
	cfg := &CLAStatusCfg{
		RootCfg: &client.BaseCfg{},
	}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg.RegisterFlags(fs)
	// CLAStatusCfg has no flags, just verify it doesn't panic
}

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

	// Create test HTTP server
	httpContent := "HTTP CLA content\n"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cla.txt" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(httpContent))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

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
		{
			name:      "HTTP URL success",
			urlOrPath: ts.URL + "/cla.txt",
			expected:  httpContent,
		},
		{
			name:        "HTTP URL not found",
			urlOrPath:   ts.URL + "/notfound.txt",
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

	t.Run("load invalid TOML", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "invalid_toml")
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		// Write invalid TOML
		configPath := filepath.Join(subDir, configFile)
		require.NoError(t, os.WriteFile(configPath, []byte("invalid { toml ["), 0o600))

		_, err := LoadConfig(subDir)
		assert.Error(t, err)
	})

	t.Run("set CLA hash on nil zones", func(t *testing.T) {
		// Test SetCLAHash initializing nil Zones map
		cfg := &Config{Zones: nil}
		cfg.SetCLAHash("https://test:443", "hash123")
		assert.Equal(t, "hash123", cfg.GetCLAHash("https://test:443"))
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

func TestExecCLAStatus(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("no CLAs signed", func(t *testing.T) {
		cfg := &CLAStatusCfg{
			RootCfg: &client.BaseCfg{
				BaseOptions: client.BaseOptions{Home: tmpDir},
			},
		}
		io := commands.NewTestIO()

		err := execCLAStatus(cfg, io)
		assert.NoError(t, err)
	})

	t.Run("with specific remote - not signed", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "not_signed")
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		cfg := &CLAStatusCfg{
			RootCfg: &client.BaseCfg{
				BaseOptions: client.BaseOptions{
					Home:   subDir,
					Remote: "https://rpc.gno.land:443",
				},
			},
		}
		io := commands.NewTestIO()

		err := execCLAStatus(cfg, io)
		assert.NoError(t, err)
	})

	t.Run("with specific remote - signed", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "signed")
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		// Save a config with CLA hash
		config := &Config{
			Zones: map[string]ZoneConfig{
				"https://rpc.gno.land:443": {CLAHash: "testhash12345678"},
			},
		}
		require.NoError(t, SaveConfig(subDir, config))

		cfg := &CLAStatusCfg{
			RootCfg: &client.BaseCfg{
				BaseOptions: client.BaseOptions{
					Home:   subDir,
					Remote: "https://rpc.gno.land:443",
				},
			},
		}
		io := commands.NewTestIO()

		err := execCLAStatus(cfg, io)
		assert.NoError(t, err)
	})

	t.Run("list all signed CLAs", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "list_all")
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		// Save a config with multiple CLA hashes
		config := &Config{
			Zones: map[string]ZoneConfig{
				"https://rpc.gno.land:443":  {CLAHash: "hash1"},
				"http://localhost:26657":    {CLAHash: "hash2"},
				"https://rpc.test.gno.land": {CLAHash: ""}, // empty hash should be skipped
			},
		}
		require.NoError(t, SaveConfig(subDir, config))

		cfg := &CLAStatusCfg{
			RootCfg: &client.BaseCfg{
				BaseOptions: client.BaseOptions{
					Home:   subDir,
					Remote: "", // empty = list all
				},
			},
		}
		io := commands.NewTestIO()

		err := execCLAStatus(cfg, io)
		assert.NoError(t, err)
	})

	t.Run("load config error", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "invalid_config")
		require.NoError(t, os.MkdirAll(subDir, 0o755))

		// Write invalid TOML
		configPath := filepath.Join(subDir, configFile)
		require.NoError(t, os.WriteFile(configPath, []byte("invalid { toml"), 0o600))

		cfg := &CLAStatusCfg{
			RootCfg: &client.BaseCfg{
				BaseOptions: client.BaseOptions{Home: subDir},
			},
		}
		io := commands.NewTestIO()

		err := execCLAStatus(cfg, io)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})
}
