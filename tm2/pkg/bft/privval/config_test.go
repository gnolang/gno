package privval

import (
	"errors"
	"fmt"
	"io/fs"
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

	t.Run("tmkms listener on unix socket allows empty allowlist", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.RemoteSigner.ServerAddress = ""
		cfg.TmkmsListener.ListenAddr = "unix:///var/run/gnoland/privval.sock"
		cfg.TmkmsListener.ChainID = "test-chain"
		// AllowedKMSPubKeys intentionally empty: a unix socket has no
		// SecretConnection and the allowlist is not enforced, so it must
		// not be required (the boundary is filesystem permissions).

		assert.NoError(t, cfg.ValidateBasic())
	})

	t.Run("tmkms listener with unsupported scheme rejected", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.RemoteSigner.ServerAddress = ""
		cfg.TmkmsListener.ListenAddr = "udp://127.0.0.1:26659"
		cfg.TmkmsListener.ChainID = "test-chain"

		assert.ErrorIs(t, cfg.ValidateBasic(), errInvalidTmkmsListenAddr)
	})

	// On Linux an empty unix:// address or one starting with "@" is an
	// abstract socket: it has no file, so socket_mode cannot protect it.
	for _, path := range []string{"", "@privval"} {
		t.Run(fmt.Sprintf("tmkms listener on abstract unix socket %q rejected", path), func(t *testing.T) {
			t.Parallel()

			assert.ErrorIs(t, tmkmsUnixListenerConfig(path).ValidateBasic(), errInvalidTmkmsListenAddr)
		})
	}

	// validTmkmsTCPConfig returns a config that clears every check up to the
	// timeout/retry bounds, so a single tweaked field is what trips ValidateBasic.
	validTmkmsTCPConfig := func() *PrivValidatorConfig {
		cfg := DefaultPrivValidatorConfig()
		cfg.RemoteSigner.ServerAddress = ""
		cfg.TmkmsListener.ListenAddr = "tcp://127.0.0.1:0"
		cfg.TmkmsListener.ChainID = "test-chain"
		cfg.TmkmsListener.AllowedKMSPubKeys = []string{
			"0000000000000000000000000000000000000000000000000000000000000000",
		}
		return cfg
	}

	t.Run("tmkms listener with sub-minimum timeout_read_write rejected", func(t *testing.T) {
		t.Parallel()

		// 1ms makes pingInterval = 1ms*2/3 = 0, which panics time.NewTicker
		// at endpoint start; it must be refused at validation time.
		cfg := validTmkmsTCPConfig()
		cfg.TmkmsListener.TimeoutReadWrite = time.Millisecond

		assert.ErrorIs(t, cfg.ValidateBasic(), errInvalidTmkmsTimeoutReadWrite)
	})

	t.Run("tmkms listener with negative retries rejected", func(t *testing.T) {
		t.Parallel()

		cfg := validTmkmsTCPConfig()
		cfg.TmkmsListener.Retries = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), errNegativeTmkmsRetries)
	})

	t.Run("tmkms listener with negative retry_timeout rejected", func(t *testing.T) {
		t.Parallel()

		cfg := validTmkmsTCPConfig()
		cfg.TmkmsListener.RetryTimeout = -time.Second

		assert.ErrorIs(t, cfg.ValidateBasic(), errNegativeTmkmsRetryTimeout)
	})

	// validTmkmsUnixConfig returns a unix:// listener config that clears every
	// other check, so socket_mode is what ValidateBasic judges.
	validTmkmsUnixConfig := func(mode string) *PrivValidatorConfig {
		cfg := tmkmsUnixListenerConfig("/var/run/gnoland/privval.sock")
		cfg.TmkmsListener.SocketMode = mode
		return cfg
	}

	for _, mode := range []string{"0600", "0660", "600", "0770"} {
		t.Run(fmt.Sprintf("tmkms listener socket_mode %q accepted", mode), func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, validTmkmsUnixConfig(mode).ValidateBasic())
		})
	}

	// Not octal, prefixed, negative, any "other" bit (every local user could
	// connect and become the signer), or a setuid/setgid/sticky bit.
	for _, mode := range []string{"", "abc", "0999", "0o660", "0x1b0", "-600", "0606", "0666", "1660", "2660", "4600", "1777", "077777"} {
		t.Run(fmt.Sprintf("tmkms listener socket_mode %q rejected", mode), func(t *testing.T) {
			t.Parallel()

			assert.ErrorIs(t, validTmkmsUnixConfig(mode).ValidateBasic(), errInvalidTmkmsSocketMode)
		})
	}

	t.Run("tmkms listener on tcp ignores socket_mode", func(t *testing.T) {
		t.Parallel()

		cfg := validTmkmsTCPConfig()
		cfg.TmkmsListener.SocketMode = "not a mode"

		assert.NoError(t, cfg.ValidateBasic())
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

	t.Run("tmkms listener invalid config rejected by factory", func(t *testing.T) {
		t.Parallel()

		// A programmatic caller may invoke NewPrivValidatorFromConfig without
		// going through PrivValidatorConfig.ValidateBasic during config load.
		// The factory must still validate the tmkms config — here an empty
		// allowlist on a TCP listener must be refused, not built fail-open.
		cfg := DefaultPrivValidatorConfig()
		cfg.RootDir = t.TempDir()
		cfg.RemoteSigner.ServerAddress = ""
		cfg.TmkmsListener.ListenAddr = "tcp://127.0.0.1:0"
		cfg.TmkmsListener.ChainID = "test-chain"
		// AllowedKMSPubKeys left empty.

		privVal, err := NewPrivValidatorFromConfig(cfg, privKey, logger)
		require.Nil(t, privVal)
		require.ErrorIs(t, err, errEmptyTmkmsAllowedPubkeys)
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
		// Required on a TCP listener; value is irrelevant here since no
		// signer dials in.
		cfg.TmkmsListener.AllowedKMSPubKeys = []string{
			"0000000000000000000000000000000000000000000000000000000000000000",
		}
		cfg.TmkmsListener.WaitForConnectionTimeout = 50 * time.Millisecond

		privVal, err := NewPrivValidatorFromConfig(cfg, privKey, logger)
		require.Nil(t, privVal)
		require.Error(t, err)
	})

	// The socket's mode is the access control on a unix:// listener: a local
	// user who can write to the socket can connect and become the signer.
	for _, tc := range []struct {
		mode string // socket_mode; empty keeps the default
		want os.FileMode
	}{
		{mode: "", want: 0o600},
		{mode: "0660", want: 0o660},
	} {
		t.Run(fmt.Sprintf("tmkms listener UDS socket has %04o perms", tc.want), func(t *testing.T) {
			t.Parallel()
			if runtime.GOOS == "windows" {
				t.Skip("UDS perm semantics differ on windows")
			}

			sockPath := shortSocketPath(t)

			cfg := tmkmsUnixListenerConfig(sockPath)
			cfg.RootDir = t.TempDir()
			if tc.mode != "" {
				cfg.TmkmsListener.SocketMode = tc.mode
			}
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

			// Poll until the socket has the configured mode, before Init
			// times out: between bind and chmod it has another one. A
			// failure reports the last mode seen.
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				info, err := os.Stat(sockPath)
				if !assert.NoError(c, err, "socket must exist before Init times out") {
					return
				}
				// Higher bits (S_IFSOCK, etc.) are fine; mask to the perm
				// bits we care about.
				assert.Equal(c, tc.want, info.Mode().Perm(),
					"UDS socket must have the configured mode, whatever the umask")
			}, 500*time.Millisecond, 10*time.Millisecond)

			// Drain the goroutine.
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("NewPrivValidatorFromConfig did not return")
			}
		})
	}

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
		// Required on a TCP listener; value is irrelevant here since no
		// signer dials in.
		cfg.TmkmsListener.AllowedKMSPubKeys = []string{
			"0000000000000000000000000000000000000000000000000000000000000000",
		}
		cfg.TmkmsListener.WaitForConnectionTimeout = 30 * time.Millisecond

		// First call: Init times out waiting for a signer to dial in.
		// Pre-fix this leaked the bound port — the second Listen below
		// would then fail with EADDRINUSE.
		_, err = NewPrivValidatorFromConfig(cfg, privKey, logger)
		require.Error(t, err)

		// Listener must be released so we can re-bind the same port.
		// Loop briefly because port release is async on some kernels.
		var rebound net.Listener
		for range 50 {
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

// tmkmsUnixListenerConfig returns a config whose only signer is a tmkms
// listener on a unix:// socket at sockPath.
func tmkmsUnixListenerConfig(sockPath string) *PrivValidatorConfig {
	cfg := DefaultPrivValidatorConfig()
	cfg.RemoteSigner.ServerAddress = ""
	cfg.TmkmsListener.ListenAddr = "unix://" + sockPath
	cfg.TmkmsListener.ChainID = "test-chain"
	return cfg
}

// shortSocketPath returns a socket path in a fresh directory under /tmp:
// macOS limits AF_UNIX paths to 104 chars, and t.TempDir's path is long
// enough to overshoot.
func shortSocketPath(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("/tmp", "pv")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "p.sock")
}

// A unix listener whose socket mode cannot be set fails the start and leaves
// no socket behind. Not parallel: it swaps chmodSocket, and Go runs parallel
// top-level tests only once the sequential ones are done.
func TestTmkmsListenerFailedChmodFailsStart(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("UDS perm semantics differ on windows")
	}

	errChmod := errors.New("chmod refused")
	orig := chmodSocket
	chmodSocket = func(string, os.FileMode) error { return errChmod }
	t.Cleanup(func() { chmodSocket = orig })

	sockPath := shortSocketPath(t)
	cfg := tmkmsUnixListenerConfig(sockPath)
	cfg.RootDir = t.TempDir()
	cfg.TmkmsListener.WaitForConnectionTimeout = 100 * time.Millisecond

	privVal, err := NewPrivValidatorFromConfig(cfg, ed25519.GenPrivKey(), log.NewNoopLogger())
	require.Nil(t, privVal)
	require.ErrorIs(t, err, errChmod)
	_, statErr := os.Stat(sockPath)
	assert.ErrorIs(t, statErr, fs.ErrNotExist, "a socket whose mode could not be set must not be left behind")
}
