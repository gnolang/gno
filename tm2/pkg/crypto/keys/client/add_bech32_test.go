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

		// Overwrite the key (name collision → override prompt)
		io.SetIn(strings.NewReader("y\n"))

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

		io.SetIn(strings.NewReader("n\n"))

		// Decline overwrite
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

	t.Run("local to offline same address different name, three-way prompt rename", func(t *testing.T) {
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
		io.SetIn(strings.NewReader(mnemonic + "\ntest1234\ntest1234\n"))

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

		// Try to add a bech32 key with same address but different name
		// This triggers case 3: same address, different name, local→offline → three-way prompt
		// Choose "r" to rename
		mockOut := bytes.NewBufferString("")
		io = commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOut))
		io.SetIn(strings.NewReader("r\n"))

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

		require.NoError(t, cmd.ParseAndRun(ctx, bech32Args))

		output := mockOut.String()
		assert.Contains(t, output, "Key collision detected:")

		// Verify the local key was renamed to "pubonly-key"
		kb, err = keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		renamedKey, err := kb.GetByName("pubonly-key")
		require.NoError(t, err)
		assert.Equal(t, localKey.GetAddress(), renamedKey.GetAddress())
		assert.Equal(t, keys.TypeLocal, renamedKey.GetType()) // still local, just renamed

		// "local-key" should no longer exist
		_, err = kb.GetByName("local-key")
		require.Error(t, err)
	})

	t.Run("local to offline same address different name, three-way prompt override", func(t *testing.T) {
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

		// First create a local key
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(mnemonic + "\ntest1234\ntest1234\n"))

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

		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		localKey, err := kb.GetByName("local-key")
		require.NoError(t, err)

		// Add bech32 with same address, different name → three-way prompt → choose override
		io = commands.NewTestIO()
		io.SetIn(strings.NewReader("o\n"))

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

		require.NoError(t, cmd.ParseAndRun(ctx, bech32Args))

		// Verify the offline key was created
		kb, err = keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		newKey, err := kb.GetByName("pubonly-key")
		require.NoError(t, err)
		assert.Equal(t, keys.TypeOffline, newKey.GetType())
		assert.Equal(t, localKey.GetAddress(), newKey.GetAddress())
	})

	t.Run("local to offline same address different name, three-way prompt cancel", func(t *testing.T) {
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
		io.SetIn(strings.NewReader(mnemonic + "\ntest1234\ntest1234\n"))

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

		// Try to add a bech32 key with same address but different name
		// This triggers case 3: same address, different name, local→offline → three-way prompt
		// Choose "c" to cancel
		io = commands.NewTestIO()
		io.SetIn(strings.NewReader("c\n"))

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

		require.ErrorIs(t, cmd.ParseAndRun(ctx, bech32Args), errOverwriteAborted)

		// Verify the local key is unchanged
		kb, err = keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		unchanged, err := kb.GetByName("local-key")
		require.NoError(t, err)
		assert.Equal(t, localKey.GetAddress(), unchanged.GetAddress())
		assert.Equal(t, keys.TypeLocal, unchanged.GetType())
	})

	t.Run("same address same name local to offline, override", func(t *testing.T) {
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

		// Create a local key
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(mnemonic + "\ntest1234\ntest1234\n"))

		cmd := NewRootCmdWithBaseConfig(io, baseOptions)
		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			"my-key",
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		localKey, err := kb.GetByName("my-key")
		require.NoError(t, err)

		// Add bech32 with same name and same address → same name + same addr + different type
		// Falls to the generic override prompt (no special warning in this path)
		mockOut := bytes.NewBufferString("")
		io = commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(mockOut))
		io.SetIn(strings.NewReader("y\n"))

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		bech32Args := []string{
			"add",
			"bech32",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--pubkey",
			localKey.GetPubKey().String(),
			"my-key",
		}

		require.NoError(t, cmd.ParseAndRun(ctx, bech32Args))

		output := mockOut.String()
		assert.Contains(t, output, "Key collision detected:")

		// Verify key is now offline
		kb, err = keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		newKey, err := kb.GetByName("my-key")
		require.NoError(t, err)
		assert.Equal(t, keys.TypeOffline, newKey.GetType())
	})

	t.Run("offline to offline same address different name, rename", func(t *testing.T) {
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

		// Try to add second offline key with same pubkey but different name
		// Same address + same type (offline) + different name → rename prompt, confirm
		io.SetIn(strings.NewReader("y\n"))

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

		require.NoError(t, cmd.ParseAndRun(ctx, args2))

		// Verify key was renamed
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		renamedKey, err := kb.GetByName("offline-key-2")
		require.NoError(t, err)
		assert.Equal(t, account.PubKey().Address(), renamedKey.GetAddress())

		// Old name should not exist
		_, err = kb.GetByName("offline-key-1")
		require.Error(t, err)
	})

	t.Run("address collision with different name, decline", func(t *testing.T) {
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
