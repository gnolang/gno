package client

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/internal/ledger"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd_Ledger(t *testing.T) {
	ledger.Discover = ledger.DiscoverMock
	t.Cleanup(func() { ledger.Discover = ledger.DiscoverDefault })

	t.Run("valid ledger reference added", func(t *testing.T) {
		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			keyName = "key-name"
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("test1234\ntest1234\n"))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"add",
			"ledger",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			keyName,
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Check the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		original, err := kb.GetByName(keyName)
		require.NoError(t, err)
		require.NotNil(t, original)
	})

	t.Run("valid ledger reference added, overwrite", func(t *testing.T) {
		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			keyName = "key-name"
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		mockOut := bytes.NewBufferString("")

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("test1234\ntest1234\n"))
		io.SetOut(commands.WriteNopCloser(mockOut))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"add",
			"ledger",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			keyName,
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Check the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		original, err := kb.GetByName(keyName)
		require.NoError(t, err)
		require.NotNil(t, original)

		mockOut.Reset()
		io.SetIn(strings.NewReader("y\ntest1234\ntest1234\n"))

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Verify the output contains diff-style collision info
		output := mockOut.String()
		assert.Contains(t, output, "Key collision detected:")
		assert.Contains(t, output, original.GetAddress().String())

		newKey, err := kb.GetByName(keyName)
		require.NoError(t, err)

		// Make sure the different key is generated and overwritten
		assert.NotEqual(t, original.GetAddress(), newKey.GetAddress())
	})

	t.Run("valid ledger reference added, no overwrite permission", func(t *testing.T) {
		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			keyName = "key-name"
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("test1234\ntest1234\n"))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"add",
			"ledger",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			keyName,
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Check the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		original, err := kb.GetByName(keyName)
		require.NoError(t, err)
		require.NotNil(t, original)

		io.SetIn(strings.NewReader("n\ntest1234\ntest1234\n"))

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		require.ErrorIs(t, cmd.ParseAndRun(ctx, args), errOverwriteAborted)

		newKey, err := kb.GetByName(keyName)
		require.NoError(t, err)

		// Make sure the key is not overwritten
		assert.Equal(t, original.GetAddress(), newKey.GetAddress())
	})
}
