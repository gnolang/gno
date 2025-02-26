package common

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func getFreePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	return port
}

func getTCPAddress(t *testing.T) string {
	t.Helper()

	return fmt.Sprintf("tcp://127.0.0.1:%d", getFreePort(t))
}

func TestNewSignerServer(t *testing.T) {
	t.Parallel()

	t.Run("nil signer", func(t *testing.T) {
		serverFlags := &ServerFlags{
			ListenAddresses: getTCPAddress(t),
		}

		signerServer, err := NewSignerServer(
			serverFlags,
			nil,
			log.NewNoopLogger(),
		)
		require.Nil(t, signerServer)
		require.Error(t, err)
	})

	t.Run("invalid auth keys file", func(t *testing.T) {
		t.Parallel()

		filePath := filepath.Join(t.TempDir(), "invalid")

		os.WriteFile(filePath, []byte("invalid"), 0o600)

		serverFlags := &ServerFlags{
			ListenAddresses: getTCPAddress(t),
			AuthFlags: AuthFlags{
				AuthKeysFile: filePath,
			},
		}

		signerServer, err := NewSignerServer(
			serverFlags,
			types.NewMockSigner(),
			log.NewNoopLogger(),
		)
		require.Nil(t, signerServer)
		require.Error(t, err)
	})

	t.Run("valid auth keys file with valid authorized keys", func(t *testing.T) {
		t.Parallel()

		filePath := filepath.Join(t.TempDir(), "invalid")

		privKey := ed25519.GenPrivKey()

		akf := AuthKeysFile{
			ServerIdentity: ServerIdentity{
				PrivKey: privKey,
				PubKey:  privKey.PubKey().String(),
			},
			ClientAuthorizedKeys: []string{ed25519.GenPrivKey().PubKey().String()},
			filePath:             filePath,
		}

		jsonBytes, err := amino.MarshalJSONIndent(akf, "", "  ")
		require.NoError(t, err)
		os.WriteFile(akf.filePath, jsonBytes, 0o600)

		serverFlags := &ServerFlags{
			ListenAddresses: getTCPAddress(t),
			AuthFlags: AuthFlags{
				AuthKeysFile: filePath,
			},
		}

		signerServer, err := NewSignerServer(
			serverFlags,
			types.NewMockSigner(),
			log.NewNoopLogger(),
		)
		require.NotNil(t, signerServer)
		require.NoError(t, err)
	})
}

func TestGenesisValidatorInfoFromSigner(t *testing.T) {
	t.Parallel()

	t.Run("nil signer", func(t *testing.T) {
		t.Parallel()

		require.Error(t, printValidatorInfo(nil, log.NewNoopLogger()))
	})

	t.Run("erroring signer", func(t *testing.T) {
		t.Parallel()

		require.Error(t, printValidatorInfo(types.NewErroringMockSigner(), log.NewNoopLogger()))
	})

	t.Run("valid signer", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, printValidatorInfo(types.NewMockSigner(), log.NewNoopLogger()))
	})
}

// mockSignerCloseFail is a mock signer that fails on close.
type mockSignerCloseFail struct {
	privKey ed25519.PrivKeyEd25519
}

// Error returned when the signer fails on close.
var errMockSignerCloseFail = errors.New("close error")

// mockSignerCloseFail implements the Signer interface.
var _ types.Signer = (*mockSignerCloseFail)(nil)

// PubKey implements the Signer interface.
func (ms *mockSignerCloseFail) PubKey() (crypto.PubKey, error) {
	return ms.privKey.PubKey(), nil
}

// Sign implements the Signer interface.
func (ms *mockSignerCloseFail) Sign(signBytes []byte) ([]byte, error) {
	return ms.privKey.Sign(signBytes)
}

// Close implements the Signer interface.
func (ms *mockSignerCloseFail) Close() error {
	return errMockSignerCloseFail
}

func TestRunSignerServer(t *testing.T) {
	t.Parallel()

	t.Run("invalid logger level", func(t *testing.T) {
		t.Parallel()

		serverFlags := &ServerFlags{
			LogLevel: "invalid",
		}

		require.Error(t, RunSignerServer(
			serverFlags,
			types.NewMockSigner(),
			commands.NewTestIO(),
		))
	})

	t.Run("nil signer", func(t *testing.T) {
		t.Parallel()

		serverFlags := &ServerFlags{}

		require.Error(t, RunSignerServer(
			serverFlags,
			nil,
			commands.NewTestIO(),
		))
	})

	t.Run("invalid auth keys file", func(t *testing.T) {
		t.Parallel()

		filePath := filepath.Join(t.TempDir(), "invalid")
		os.WriteFile(filePath, []byte("invalid"), 0o600)

		serverFlags := &ServerFlags{
			ListenAddresses: getTCPAddress(t),
			LogLevel:        zapcore.ErrorLevel.String(),
			AuthFlags: AuthFlags{
				AuthKeysFile: filePath,
			},
		}

		require.Error(t, RunSignerServer(
			serverFlags,
			types.NewMockSigner(),
			commands.NewDefaultIO(),
		))
	})

	t.Run("listener not free", func(t *testing.T) {
		t.Parallel()

		serverFlags := &ServerFlags{
			ListenAddresses: getTCPAddress(t),
			LogLevel:        zapcore.ErrorLevel.String(),
		}

		// Listen on the address to make it unavailable.
		protocol, address := osm.ProtocolAndAddress(serverFlags.ListenAddresses)
		net.Listen(protocol, address)

		require.Error(t, RunSignerServer(
			serverFlags,
			types.NewMockSigner(),
			commands.NewDefaultIO(),
		))
	})

	t.Run("signer fail on close", func(t *testing.T) {
		t.Parallel()

		serverFlags := &ServerFlags{
			ListenAddresses: getTCPAddress(t),
			LogLevel:        zapcore.ErrorLevel.String(),
		}

		require.ErrorIs(t, RunSignerServer(
			serverFlags,
			&mockSignerCloseFail{privKey: ed25519.GenPrivKey()},
			commands.NewDefaultIO(),
		),
			errMockSignerCloseFail,
		)
	})

	t.Run("valid server params", func(t *testing.T) {
		t.Parallel()

		serverFlags := &ServerFlags{
			ListenAddresses: getTCPAddress(t),
			LogLevel:        zapcore.ErrorLevel.String(),
		}

		// Simulate a SIGINT signal after 50 milliseconds.
		go func() {
			time.Sleep(50 * time.Millisecond)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()

		require.NoError(t, RunSignerServer(
			serverFlags,
			types.NewMockSigner(),
			commands.NewDefaultIO(),
		))
	})
}
