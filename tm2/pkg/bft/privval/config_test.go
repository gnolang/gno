package privval

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

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
		// Local/remote-client paths return the concrete *PrivValidator;
		// type-assert to inspect the inner signer.
		pv, ok := privval.(*PrivValidator)
		require.True(t, ok, "expected *PrivValidator, got %T", privval)
		assert.IsType(t, &local.LocalSigner{}, pv.signer)
		pv.Close()
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
		pv, ok := privval.(*PrivValidator)
		require.True(t, ok, "expected *PrivValidator, got %T", privval)
		assert.IsType(t, &rsclient.RemoteSignerClient{}, pv.signer)
		pv.Close()
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

	t.Run("both external signers configured rejected", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.RootDir = t.TempDir()
		cfg.RemoteSigner.ServerAddress = "unix:///tmp/x.sock"
		cfg.TmkmsListener.ListenAddr = "tcp://127.0.0.1:0"
		cfg.TmkmsListener.ChainID = "test"

		_, err := NewPrivValidatorFromConfig(cfg, privKey, logger)
		require.ErrorIs(t, err, errBothExternalSignersEnabled)
	})

	t.Run("tmkms listener wait-timeout surfaces as error", func(t *testing.T) {
		t.Parallel()

		// Configure a TCP listener at an OS-chosen port (no signer will
		// dial in within the test budget — Init must surface the
		// connection-wait timeout as an error).
		probe, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := probe.Addr().String()
		require.NoError(t, probe.Close())

		cfg := DefaultPrivValidatorConfig()
		cfg.RootDir = t.TempDir()
		cfg.RemoteSigner.ServerAddress = "" // disable gnokms path
		cfg.TmkmsListener.ListenAddr = "tcp://" + addr
		cfg.TmkmsListener.ChainID = "test-chain"
		cfg.TmkmsListener.WaitForConnectionTimeout = 50 * time.Millisecond

		privVal, err := NewPrivValidatorFromConfig(cfg, privKey, logger)
		require.Nil(t, privVal)
		require.Error(t, err)
	})
}
