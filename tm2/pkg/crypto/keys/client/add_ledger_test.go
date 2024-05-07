//go:build ledger_suite
// +build ledger_suite

package client

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Make sure to run these tests with the following tag enabled:
// -tags='ledger_suite'
func TestAdd_Ledger(t *testing.T) {
	t.Parallel()

	t.Run("valid ledger reference added", func(t *testing.T) {
		t.Parallel()

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
		t.Parallel()

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

		io.SetIn(strings.NewReader("y\ntest1234\ntest1234\n"))

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		require.NoError(t, cmd.ParseAndRun(ctx, args))

		newKey, err := kb.GetByName(keyName)
		require.NoError(t, err)

		// Make sure the different key is generated and overwritten
		assert.NotEqual(t, original.GetAddress(), newKey.GetAddress())
	})

	t.Run("valid ledger reference added, no overwrite permission", func(t *testing.T) {
		t.Parallel()

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
