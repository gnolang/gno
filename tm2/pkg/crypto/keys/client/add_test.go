package client

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd_Base_Add(t *testing.T) {
	t.Parallel()

	t.Run("valid key addition, generated mnemonic", func(t *testing.T) {
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

	t.Run("valid key addition, overwrite", func(t *testing.T) {
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

		// Password first, then override confirmation (collision check happens after mnemonic derivation)
		io.SetIn(strings.NewReader("test1234\ntest1234\ny\n"))

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		require.NoError(t, cmd.ParseAndRun(ctx, args))

		newKey, err := kb.GetByName(keyName)
		require.NoError(t, err)

		// Make sure the different key is generated and overwritten
		assert.NotEqual(t, original.GetAddress(), newKey.GetAddress())
	})

	t.Run("valid key addition, provided mnemonic", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			mnemonic    = generateTestMnemonic(t)
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			keyName = "key-name"
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("test1234" + "\n" + "test1234" + "\n" + mnemonic + "\n"))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			keyName,
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))
		// Check the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		key, err := kb.GetByName(keyName)
		require.NoError(t, err)
		require.NotNil(t, key)

		// Get the account
		accounts := generateAccounts(mnemonic, []string{"44'/118'/0'/0/0"})

		assert.Equal(t, accounts[0].String(), key.GetAddress().String())
	})

	t.Run("no overwrite permission", func(t *testing.T) {
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

		// Password first, then decline override
		io.SetIn(strings.NewReader("test1234\ntest1234\nn\n"))

		// Confirm overwrite
		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		require.ErrorIs(t, cmd.ParseAndRun(ctx, args), errOverwriteAborted)

		newKey, err := kb.GetByName(keyName)
		require.NoError(t, err)

		// Make sure the key is not overwritten
		assert.Equal(t, original.GetAddress(), newKey.GetAddress())
	})

	t.Run("valid key addition, provided mnemonic with masked flag", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: false, // Not using insecure mode
				Home:                  kbHome,
			}

			keyName  = "key-name"
			mnemonic = "equip will roof matter pink blind book anxiety banner elbow sun young"
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create base IO
		baseIO := commands.NewTestIO()
		var outBuf, errBuf bytes.Buffer
		baseIO.SetOut(commands.WriteNopCloser(&outBuf))
		baseIO.SetErr(commands.WriteNopCloser(&errBuf))

		// Create mock IO that handles all password calls
		// The order is: encryption password, repeat password, then mnemonic
		mockIO := &mockPasswordIO{
			IO:        baseIO,
			passwords: []string{"test1234", "test1234", mnemonic},
		}

		// Create the command
		cmd := NewRootCmdWithBaseConfig(mockIO, baseOptions)

		args := []string{
			"add",
			"--home",
			kbHome,
			"--recover",
			"--masked",
			keyName,
		}

		// This uses our mock GetPassword for all password inputs
		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Check the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		key, err := kb.GetByName(keyName)
		require.NoError(t, err)
		require.NotNil(t, key)

		// Verify the key was created correctly
		assert.NotNil(t, key)
	})

	t.Run("valid key addition, entropy with masked flag", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: false, // Not using insecure mode
				Home:                  kbHome,
			}

			keyName = "entropy-key"
			entropy = "this is test entropy that is long enough to meet the minimum requirement"
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Create base IO
		baseIO := commands.NewTestIO()
		var outBuf, errBuf bytes.Buffer
		baseIO.SetOut(commands.WriteNopCloser(&outBuf))
		baseIO.SetErr(commands.WriteNopCloser(&errBuf))

		// Create mock that handles password input for entropy
		// Order: encryption password, repeat password, entropy
		mockIO := &mockPasswordIO{
			IO:        baseIO,
			passwords: []string{"test1234", "test1234", entropy},
		}
		// For confirmation prompt after entropy
		mockIO.SetIn(strings.NewReader("y\n"))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(mockIO, baseOptions)

		args := []string{
			"add",
			"--home",
			kbHome,
			"--entropy",
			"--masked",
			keyName,
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Check the keybase
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		key, err := kb.GetByName(keyName)
		require.NoError(t, err)
		require.NotNil(t, key)
	})

	t.Run("name collision shows diff output", func(t *testing.T) {
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

		mockOut := bytes.NewBufferString("")

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("test1234\ntest1234\n"))
		io.SetOut(commands.WriteNopCloser(mockOut))

		// Create initial key
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)
		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			keyName,
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Get the original key info
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		original, err := kb.GetByName(keyName)
		require.NoError(t, err)

		// Try to add another key with the same name, confirm overwrite
		// Password first, then override confirmation
		mockOut.Reset()
		io.SetIn(strings.NewReader("test1234\ntest1234\ny\n"))

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Verify the output contains diff-style collision info
		output := mockOut.String()
		assert.Contains(t, output, "Key collision detected:")
		assert.Contains(t, output, original.GetAddress().String())
	})

	t.Run("address collision with recover, decline", func(t *testing.T) {
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

		io := commands.NewTestIO()

		// Create initial key with mnemonic
		io.SetIn(strings.NewReader("test1234\ntest1234\n" + mnemonic + "\n"))
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)
		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			"key1",
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Try to add another key with the same mnemonic (same address), decline rename
		io.SetIn(strings.NewReader("test1234\ntest1234\n" + mnemonic + "\nn\n"))
		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		args2 := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			"key2",
		}

		require.ErrorIs(t, cmd.ParseAndRun(ctx, args2), errOverwriteAborted)

		// Verify original key is untouched
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		original, err := kb.GetByName("key1")
		require.NoError(t, err)
		require.NotNil(t, original)
	})

	t.Run("address collision with recover, confirm rename", func(t *testing.T) {
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

		io := commands.NewTestIO()

		// Create initial key with mnemonic
		io.SetIn(strings.NewReader("test1234\ntest1234\n" + mnemonic + "\n"))
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)
		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			"key1",
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Verify key1 exists
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		original, err := kb.GetByName("key1")
		require.NoError(t, err)

		// Add key2 with same mnemonic, confirm rename
		// Same address + same type + different name → rename prompt
		io.SetIn(strings.NewReader("test1234\ntest1234\n" + mnemonic + "\ny\n"))
		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		args2 := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			"key2",
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args2))

		// Verify key1 was renamed to key2 (same address)
		newKey, err := kb.GetByName("key2")
		require.NoError(t, err)
		assert.Equal(t, original.GetAddress(), newKey.GetAddress())

		// key1 should no longer exist
		_, err = kb.GetByName("key1")
		require.Error(t, err)
	})

	t.Run("identical key recovery skips without prompt", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			mnemonic    = generateTestMnemonic(t)
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
		io.SetOut(commands.WriteNopCloser(mockOut))

		// Create initial key with mnemonic
		io.SetIn(strings.NewReader("test1234\ntest1234\n" + mnemonic + "\n"))
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)
		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			keyName,
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Recover with same name and same mnemonic → identical key, skip (no prompt)
		mockOut.Reset()
		io.SetIn(strings.NewReader("test1234\ntest1234\n" + mnemonic + "\n"))
		cmd = NewRootCmdWithBaseConfig(io, baseOptions)

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Verify skip message
		output := mockOut.String()
		assert.Contains(t, output, "Key is identical. Skipping.")

		// Verify key still exists
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		seed := bip39.NewSeed(mnemonic, "")
		expectedAddr := generateKeyFromSeed(seed, "44'/118'/0'/0/0").PubKey().Address()

		key, err := kb.GetByName(keyName)
		require.NoError(t, err)
		assert.Equal(t, expectedAddr, key.GetAddress())
	})

	t.Run("force flag overrides without prompt", func(t *testing.T) {
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

		// Create initial key
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)
		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			keyName,
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		original, err := kb.GetByName(keyName)
		require.NoError(t, err)

		// Add with --force (no confirmation prompt needed)
		io.SetIn(strings.NewReader("test1234\ntest1234\n"))

		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		forceArgs := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--force",
			keyName,
		}

		require.NoError(t, cmd.ParseAndRun(ctx, forceArgs))

		// Verify the key was overridden (different mnemonic → different address)
		newKey, err := kb.GetByName(keyName)
		require.NoError(t, err)
		assert.NotEqual(t, original.GetAddress(), newKey.GetAddress())
	})
}

func generateDerivationPaths(count int) []string {
	paths := make([]string, count)

	for i := range count {
		paths[i] = fmt.Sprintf("44'/118'/0'/0/%d", i)
	}

	return paths
}

func TestAdd_Derive(t *testing.T) {
	t.Parallel()

	t.Run("valid address derivation", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome   = t.TempDir()
			mnemonic = generateTestMnemonic(t)
			paths    = generateDerivationPaths(10)

			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			dummyPass = "dummy-pass"
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		mockOut := bytes.NewBufferString("")

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(dummyPass + "\n" + dummyPass + "\n" + mnemonic + "\n"))
		io.SetOut(commands.WriteNopCloser(mockOut))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			"example-key",
		}

		for _, path := range paths {
			args = append(
				args, []string{
					"--derivation-path",
					path,
				}...,
			)
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		// Verify the addresses are derived correctly
		expectedAccounts := generateAccounts(
			mnemonic,
			paths,
		)

		// Grab the output
		deriveOutput := mockOut.String()

		for _, expectedAccount := range expectedAccounts {
			assert.Contains(t, deriveOutput, expectedAccount.String())
		}
	})

	t.Run("malformed derivation path", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			mnemonic    = generateTestMnemonic(t)
			dummyPass   = "dummy-pass"
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		mockOut := bytes.NewBufferString("")

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(dummyPass + "\n" + dummyPass + "\n" + mnemonic + "\n"))
		io.SetOut(commands.WriteNopCloser(mockOut))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			"example-key",
			"--derivation-path",
			"malformed path",
		}

		require.ErrorIs(t, cmd.ParseAndRun(ctx, args), errInvalidDerivationPath)
	})

	t.Run("invalid derivation path", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			mnemonic    = generateTestMnemonic(t)
			dummyPass   = "dummy-pass"
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		mockOut := bytes.NewBufferString("")

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(dummyPass + "\n" + dummyPass + "\n" + mnemonic + "\n"))
		io.SetOut(commands.WriteNopCloser(mockOut))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			"example-key",
			"--derivation-path",
			"44'/500'/0'/0/0", // invalid coin type
		}

		require.ErrorIs(t, cmd.ParseAndRun(ctx, args), errInvalidDerivationPath)
	})
}
