package common

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestNewSignerServer(t *testing.T) {
	t.Parallel()

	t.Run("nil signer", func(t *testing.T) {
		serverFlags := &ServerFlags{
			Listener: "tcp://127.0.0.1:0",
		}

		signerServer, err := NewSignerServer(
			serverFlags,
			nil,
			log.NewNoopLogger(),
		)
		require.Nil(t, signerServer)
		assert.Error(t, err)
	})

	t.Run("invalid auth keys file", func(t *testing.T) {
		t.Parallel()

		filePath := filepath.Join(t.TempDir(), "invalid")

		os.WriteFile(filePath, []byte("invalid"), 0o600)

		serverFlags := &ServerFlags{
			Listener: "tcp://127.0.0.1:0",
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
		assert.Error(t, err)
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
		}

		jsonBytes, err := amino.MarshalJSONIndent(akf, "", "  ")
		require.NoError(t, err)
		os.WriteFile(filePath, jsonBytes, 0o600)

		serverFlags := &ServerFlags{
			Listener: "tcp://127.0.0.1:0",
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
		assert.NoError(t, err)
	})
}

func TestGenesisValidatorInfoFromSigner(t *testing.T) {
	t.Parallel()

	t.Run("nil signer", func(t *testing.T) {
		t.Parallel()

		assert.Error(t, printValidatorInfo(nil, log.NewNoopLogger()))
	})

	t.Run("valid signer", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, printValidatorInfo(types.NewMockSigner(), log.NewNoopLogger()))
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
func (ms *mockSignerCloseFail) PubKey() crypto.PubKey {
	return ms.privKey.PubKey()
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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		serverFlags := &ServerFlags{
			LogLevel: "invalid",
		}

		assert.Error(t, RunSignerServer(
			ctx,
			serverFlags,
			types.NewMockSigner(),
			commands.NewTestIO(),
		))
	})

	t.Run("nil signer", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		serverFlags := &ServerFlags{}

		assert.Error(t, RunSignerServer(
			ctx,
			serverFlags,
			nil,
			commands.NewTestIO(),
		))
	})

	t.Run("invalid auth keys file", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filePath := filepath.Join(t.TempDir(), "invalid")
		os.WriteFile(filePath, []byte("invalid"), 0o600)

		serverFlags := &ServerFlags{
			Listener: "tcp://127.0.0.1:0",
			LogLevel: zapcore.ErrorLevel.String(),
			AuthFlags: AuthFlags{
				AuthKeysFile: filePath,
			},
		}

		assert.Error(t, RunSignerServer(
			ctx,
			serverFlags,
			types.NewMockSigner(),
			commands.NewDefaultIO(),
		))
	})

	t.Run("listener not free", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Listen on the address to make it unavailable.
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		// Use the address:port for the server flags.
		serverFlags := &ServerFlags{
			Listener: fmt.Sprintf("tcp://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port),
			LogLevel: zapcore.ErrorLevel.String(),
		}

		assert.Error(t, RunSignerServer(
			ctx,
			serverFlags,
			types.NewMockSigner(),
			commands.NewDefaultIO(),
		))
	})

	t.Run("signer fail on close", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		serverFlags := &ServerFlags{
			Listener: "tcp://127.0.0.1:0",
			LogLevel: zapcore.ErrorLevel.String(),
		}

		assert.ErrorIs(t, RunSignerServer(
			ctx,
			serverFlags,
			&mockSignerCloseFail{privKey: ed25519.GenPrivKey()},
			commands.NewDefaultIO(),
		),
			errMockSignerCloseFail,
		)
	})

	t.Run("valid server params", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		serverFlags := &ServerFlags{
			Listener: "tcp://127.0.0.1:0",
			LogLevel: zapcore.ErrorLevel.String(),
		}

		assert.NoError(t, RunSignerServer(
			ctx,
			serverFlags,
			types.NewMockSigner(),
			commands.NewDefaultIO(),
		))
	})
}
