package packages

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGnowork(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		overrides map[string]string
		wantErr   bool
	}{
		{
			name:      "empty",
			body:      "",
			overrides: map[string]string{},
		},
		{
			name: "single domain override",
			body: `[domains."gno.land"]
rpc = "http://localhost:26657"`,
			overrides: map[string]string{"gno.land": "http://localhost:26657"},
		},
		{
			name: "multiple domains",
			body: `[domains."gno.land"]
rpc = "http://localhost:26657"

[domains."example.com"]
rpc = "http://localhost:8080"`,
			overrides: map[string]string{
				"gno.land":    "http://localhost:26657",
				"example.com": "http://localhost:8080",
			},
		},
		{
			name: "empty rpc is skipped",
			body: `[domains."gno.land"]
rpc = ""`,
			overrides: map[string]string{},
		},
		{
			name:    "malformed toml",
			body:    `[domains."gno.land"`,
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gw, err := ParseGnowork([]byte(c.body))
			if c.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.overrides, gw.rpcOverrides())
		})
	}
}

func TestGnoworkRPCOverridesNil(t *testing.T) {
	// a nil *Gnowork (non-workspace context) must not panic and yields no overrides
	var gw *Gnowork
	require.Nil(t, gw.rpcOverrides())
}

func TestReadGnowork(t *testing.T) {
	dir := t.TempDir()

	t.Run("valid file", func(t *testing.T) {
		file := filepath.Join(dir, "gnowork.toml")
		require.NoError(t, os.WriteFile(file, []byte(`[domains."gno.land"]
rpc = "http://localhost:26657"`), 0o600))

		gw, err := ReadGnowork(file)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"gno.land": "http://localhost:26657"}, gw.rpcOverrides())
	})

	t.Run("missing file wraps path", func(t *testing.T) {
		file := filepath.Join(dir, "does-not-exist.toml")
		_, err := ReadGnowork(file)
		require.ErrorContains(t, err, file)
	})

	t.Run("malformed file wraps path", func(t *testing.T) {
		file := filepath.Join(dir, "bad.toml")
		require.NoError(t, os.WriteFile(file, []byte(`[domains."gno.land"`), 0o600))
		_, err := ReadGnowork(file)
		require.ErrorContains(t, err, file)
	})
}
