package client

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd_Bech32(t *testing.T) {
	t.Parallel()

	t.Run("valid bech32 addition", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			seed    = bip39.NewSeed(generateTestMnemonic(t), "")
			account = generateKeyFromSeed(seed, "44'/118'/0'/0/0")

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
			"bech32",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--pubkey",
			account.PubKey().String(),
			keyName,
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Check the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		original, err := kb.GetByName(keyName)
		require.NoError(t, err)
		require.NotNil(t, original)

		assert.Equal(t, account.PubKey().Address().String(), original.GetAddress().String())
	})

	t.Run("valid bech32 addition, overwrite", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			seed            = bip39.NewSeed(generateTestMnemonic(t), "")
			originalAccount = generateKeyFromSeed(seed, "44'/118'/0'/0/0")
			copyAccount     = generateKeyFromSeed(seed, "44'/118'/0'/0/1")

			keyName = "key-name"
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("test1234\ntest1234\n"))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		baseArgs := []string{
			"add",
			"bech32",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			keyName,
		}

		initialArgs := append(baseArgs, []string{
			"--pubkey",
			originalAccount.PubKey().String(),
		}...)

		require.NoError(t, cmd.ParseAndRun(ctx, initialArgs))

		// Check the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		original, err := kb.GetByName(keyName)
		require.NoError(t, err)

		require.Equal(t, originalAccount.PubKey().Address().String(), original.GetAddress().String())

		// Overwrite the key
		io.SetIn(strings.NewReader("y\ntest1234\ntest1234\n"))

		secondaryArgs := append(baseArgs, []string{
			"--pubkey",
			copyAccount.PubKey().String(),
		}...)

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		require.NoError(t, cmd.ParseAndRun(ctx, secondaryArgs))

		newKey, err := kb.GetByName(keyName)
		require.NoError(t, err)

		require.Equal(t, copyAccount.PubKey().Address().String(), newKey.GetAddress().String())
	})

	t.Run("no overwrite permission", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			seed            = bip39.NewSeed(generateTestMnemonic(t), "")
			originalAccount = generateKeyFromSeed(seed, "44'/118'/0'/0/0")
			copyAccount     = generateKeyFromSeed(seed, "44'/118'/0'/0/1")

			keyName = "key-name"
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("test1234\ntest1234\n"))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		baseArgs := []string{
			"add",
			"bech32",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			keyName,
		}

		initialArgs := append(baseArgs, []string{
			"--pubkey",
			originalAccount.PubKey().String(),
		}...)

		require.NoError(t, cmd.ParseAndRun(ctx, initialArgs))

		// Check the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		original, err := kb.GetByName(keyName)
		require.NoError(t, err)

		io.SetIn(strings.NewReader("n\ntest1234\ntest1234\n"))

		// Confirm overwrite
		secondaryArgs := append(baseArgs, []string{
			"--pubkey",
			copyAccount.PubKey().String(),
		}...)

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		require.ErrorIs(t, cmd.ParseAndRun(ctx, secondaryArgs), errOverwriteAborted)

		newKey, err := kb.GetByName(keyName)
		require.NoError(t, err)

		// Make sure the key is not overwritten
		assert.Equal(t, original.GetAddress(), newKey.GetAddress())
	})
}
