package privval

import (
	"fmt"
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	rsclient "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/client"
	rsserver "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/server"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateBasic(t *testing.T) {
	t.Parallel()

	t.Run("default config", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, DefaultPrivValidatorConfig().ValidateBasic())
	})

	t.Run("test config", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, TestPrivValidatorConfig().ValidateBasic())
	})

	t.Run("sign state file path is not set", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.SignState = ""

		assert.ErrorIs(t, cfg.ValidateBasic(), errInvalidSignStatePath)
	})

	t.Run("local signer file path is not set", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.LocalSigner = ""

		assert.ErrorIs(t, cfg.ValidateBasic(), errInvalidLocalSignerPath)
	})

	t.Run("remote signer config is nil", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.RemoteSigner = nil

		assert.ErrorIs(t, cfg.ValidateBasic(), errNilRemoteSignerConfig)
	})

	t.Run("remote signer config with invalid key", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.RemoteSigner.AuthorizedKeys = []string{"invalid"}

		assert.Error(t, cfg.ValidateBasic())
	})
}

func TestPathGetters(t *testing.T) {
	t.Parallel()

	const rootDir = "/root/dir"

	cfg := DefaultPrivValidatorConfig()

	require.NotContains(t, cfg.LocalSignerPath(), rootDir)
	require.NotContains(t, cfg.SignStatePath(), rootDir)

	cfg.RootDir = rootDir

	require.Contains(t, cfg.LocalSignerPath(), rootDir)
	assert.Contains(t, cfg.SignStatePath(), rootDir)
}

func TestNewPrivValidatorFromConfig(t *testing.T) {
	t.Parallel()

	var (
		privKey = ed25519.GenPrivKey()
		logger  = log.NewNoopLogger()
	)

	t.Run("valid local signer", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.RootDir = t.TempDir()

		privval, err := NewPrivValidatorFromConfig(cfg, privKey, logger)
		require.NotNil(t, privval)
		require.NoError(t, err)
		assert.IsType(t, &local.LocalSigner{}, privval.signer)
		privval.Close()
	})

	t.Run("valid remote signer", func(t *testing.T) {
		t.Parallel()

		// Setup a Unix socket address for the remote signer.
		unixSocketPath := "test_tm2_remote_signer"
		addr := fmt.Sprintf("unix://%s/%s.sock", unixSocketPath, xid.New().String())

		// Create the directory for the Unix socket then remove it after the test.
		os.MkdirAll(unixSocketPath, 0o755)
		t.Cleanup(func() {
			os.Remove(unixSocketPath)
		})

		// Init a remote signer server to fetch the public key on client init.
		rss, err := rsserver.NewRemoteSignerServer(types.NewMockSigner(), addr, log.NewNoopLogger())
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())

		cfg := DefaultPrivValidatorConfig()
		cfg.RootDir = t.TempDir()
		cfg.RemoteSigner.ServerAddress = addr

		privval, err := NewPrivValidatorFromConfig(cfg, privKey, logger)
		require.NotNil(t, privval)
		require.NoError(t, err)
		assert.IsType(t, &rsclient.RemoteSignerClient{}, privval.signer)
		privval.Close()
		rss.Stop()
	})

	t.Run("invalid authorized keys", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.RootDir = t.TempDir()
		cfg.RemoteSigner.ServerAddress = "unix:///tmp/remote_signer.sock"
		cfg.RemoteSigner.AuthorizedKeys = []string{"invalid"}

		privval, err := NewPrivValidatorFromConfig(cfg, privKey, logger)
		require.Nil(t, privval)
		assert.Error(t, err)
	})
}
