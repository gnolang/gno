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

func TestAdd_Multisig(t *testing.T) {
	t.Parallel()

	t.Run("invalid multisig threshold", func(t *testing.T) {
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
			"multisig",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--multisig",
			"example",
			"--threshold",
			"2",
			keyName,
		}

		require.ErrorIs(t, cmd.ParseAndRun(ctx, args), errUnableToVerifyMultisig)
	})

	t.Run("valid multisig reference added", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}
			mnemonic = generateTestMnemonic(t)

			keyNames = []string{
				"key-1",
				"key-2",
			}
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("y\ntest1234\ntest1234\n"))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"add",
			"multisig",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--multisig",
			keyNames[0],
			"--multisig",
			keyNames[1],
			keyNames[0],
		}

		// Prepare the multisig keys
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		for index, keyName := range keyNames {
			_, err = kb.CreateAccount(
				keyName,
				mnemonic,
				"",
				"123",
				0,
				uint32(index),
			)

			require.NoError(t, err)
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Verify the key is multisig
		original, err := kb.GetByName(keyNames[0])
		require.NoError(t, err)
		require.NotNil(t, original)

		assert.Equal(t, original.GetType(), keys.TypeMulti)
	})

	t.Run("multisig address collision, decline overwrite", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}
			mnemonic = generateTestMnemonic(t)

			keyNames = []string{
				"key-1",
				"key-2",
			}
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Prepare the multisig keys
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		for index, keyName := range keyNames {
			_, err = kb.CreateAccount(
				keyName,
				mnemonic,
				"",
				"123",
				0,
				uint32(index),
			)

			require.NoError(t, err)
		}

		// Create first multisig key
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("y\n"))

		cmd := NewRootCmdWithBaseConfig(io, baseOptions)
		args := []string{
			"add",
			"multisig",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--multisig",
			keyNames[0],
			"--multisig",
			keyNames[1],
			"multi-1",
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Try to create second multisig with same keys (same address), different name
		// Decline address collision overwrite
		io.SetIn(strings.NewReader("n\n"))

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		args2 := []string{
			"add",
			"multisig",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--multisig",
			keyNames[0],
			"--multisig",
			keyNames[1],
			"multi-2",
		}

		require.ErrorIs(t, cmd.ParseAndRun(ctx, args2), errOverwriteAborted)
	})
}
