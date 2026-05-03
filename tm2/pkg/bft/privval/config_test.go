package privval

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
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

	t.Run("tmkms listener enabled with empty allowlist rejected", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.RemoteSigner.ServerAddress = "" // disable gnokms path
		cfg.TmkmsListener.ListenAddr = "tcp://127.0.0.1:0"
		cfg.TmkmsListener.ChainID = "test-chain"
		// Leave AllowedKMSPubKeys empty — production-mode footgun must
		// be refused at ValidateBasic time so operators don't ship a
		// validator that accepts any peer who completes the SecretConn
		// handshake.

		assert.ErrorIs(t, cfg.ValidateBasic(), errEmptyTmkmsAllowedPubkeys)
	})

	t.Run("tmkms listener with unsupported protocol_version rejected", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.RemoteSigner.ServerAddress = ""
		cfg.TmkmsListener.ListenAddr = "tcp://127.0.0.1:0"
		cfg.TmkmsListener.ChainID = "test-chain"
		cfg.TmkmsListener.AllowedKMSPubKeys = []string{
			// Any 64-hex-char ed25519 pubkey suffices to clear the
			// allowlist and parse checks; we only want to reach the
			// protocol-version check after them.
			"0000000000000000000000000000000000000000000000000000000000000000",
		}
		cfg.TmkmsListener.ProtocolVersion = "v0.99"

		assert.ErrorIs(t, cfg.ValidateBasic(), errUnsupportedProtocolVersion)
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

	t.Run("tmkms listener UDS socket has 0600 perms", func(t *testing.T) {
		t.Parallel()
		// Default umask on most distros leaves UDS sockets group/world
		// readable+writable. The factory chmods 0600 so a local
		// non-root user can't enter the SecretConn handshake and
		// attempt to become the signer.
		if runtime.GOOS == "windows" {
			t.Skip("UDS perm semantics differ on windows")
		}

		// macOS limits AF_UNIX paths to 104 chars; t.TempDir's path is
		// long enough to overshoot, so use a short /tmp directory.
		dir, err := os.MkdirTemp("/tmp", "pv")
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.RemoveAll(dir) })
		sockPath := filepath.Join(dir, "p.sock")

		cfg := DefaultPrivValidatorConfig()
		cfg.RootDir = t.TempDir()
		cfg.RemoteSigner.ServerAddress = ""
		cfg.TmkmsListener.ListenAddr = "unix://" + sockPath
		cfg.TmkmsListener.ChainID = "test-chain"
		// Long enough that the socket exists while we stat it; short
		// enough that the test exits quickly.
		cfg.TmkmsListener.WaitForConnectionTimeout = 800 * time.Millisecond

		// Run the factory in a goroutine — Init will block waiting for
		// a signer that never arrives, then time out. Meanwhile we stat
		// the socket from the main goroutine.
		done := make(chan error, 1)
		go func() {
			_, err := NewPrivValidatorFromConfig(cfg, privKey, logger)
			done <- err
		}()

		// Poll for the socket to exist (Listen is fast; chmod runs
		// immediately after).
		var info os.FileInfo
		var statErr error
		deadline := time.Now().Add(500 * time.Millisecond)
		for time.Now().Before(deadline) {
			info, statErr = os.Stat(sockPath)
			if statErr == nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		require.NoError(t, statErr, "socket must exist before Init times out")

		// Mode bits: only owner-RW. Higher bits (S_IFSOCK, etc.) are
		// fine; mask to the perm bits we care about.
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(),
			"UDS socket must be 0600 to keep non-owner users from connecting")

		// Drain the goroutine.
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("NewPrivValidatorFromConfig did not return")
		}
	})

	t.Run("tmkms listener Init failure releases the listener", func(t *testing.T) {
		t.Parallel()

		// Get an OS-chosen port, release it.
		probe, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := probe.Addr().String()
		require.NoError(t, probe.Close())

		cfg := DefaultPrivValidatorConfig()
		cfg.RootDir = t.TempDir()
		cfg.RemoteSigner.ServerAddress = ""
		cfg.TmkmsListener.ListenAddr = "tcp://" + addr
		cfg.TmkmsListener.ChainID = "test-chain"
		cfg.TmkmsListener.WaitForConnectionTimeout = 30 * time.Millisecond

		// First call: Init times out waiting for a signer to dial in.
		// Pre-fix this leaked the bound port — the second Listen below
		// would then fail with EADDRINUSE.
		_, err = NewPrivValidatorFromConfig(cfg, privKey, logger)
		require.Error(t, err)

		// Listener must be released so we can re-bind the same port.
		// Loop briefly because port release is async on some kernels.
		var rebound net.Listener
		for i := 0; i < 50; i++ {
			rebound, err = net.Listen("tcp", addr)
			if err == nil {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		require.NoError(t, err, "port must be released after Init failure")
		require.NoError(t, rebound.Close())
	})
}
