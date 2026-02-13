package client

import (
	"bytes"
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

	t.Run("skip when local key exists at same address", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			mnemonic    = generateTestMnemonic(t)
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// First create a local key using `add --recover`
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("test1234\ntest1234\n" + mnemonic + "\n"))

		cmd := NewRootCmdWithBaseConfig(io, baseOptions)
		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			"local-key",
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Get the public key from the local key
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		localKey, err := kb.GetByName("local-key")
		require.NoError(t, err)

		// Try to add a bech32 key with the same public key
		mockOut := bytes.NewBufferString("")
		io = commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOut))

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		bech32Args := []string{
			"add",
			"bech32",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--pubkey",
			localKey.GetPubKey().String(),
			"pubonly-key",
		}

		// Should succeed but skip adding
		require.NoError(t, cmd.ParseAndRun(ctx, bech32Args))

		output := mockOut.String()
		assert.Contains(t, output, "A key with signing capability already exists at this address:")
		assert.Contains(t, output, "redundant")

		// Verify the pub-only key was NOT created
		_, err = kb.GetByName("pubonly-key")
		require.Error(t, err)
	})

	t.Run("address collision with different name", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			seed    = bip39.NewSeed(generateTestMnemonic(t), "")
			account = generateKeyFromSeed(seed, "44'/118'/0'/0/0")
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("test1234\ntest1234\n"))

		// Add first offline key
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)
		args := []string{
			"add",
			"bech32",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--pubkey",
			account.PubKey().String(),
			"offline-key-1",
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Try to add second offline key with same pubkey but different name, decline
		io.SetIn(strings.NewReader("n\n"))

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		args2 := []string{
			"add",
			"bech32",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--pubkey",
			account.PubKey().String(),
			"offline-key-2",
		}

		require.ErrorIs(t, cmd.ParseAndRun(ctx, args2), errOverwriteAborted)
	})
}
