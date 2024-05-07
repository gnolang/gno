package client

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
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

		io.SetIn(strings.NewReader("y\ntest1234\ntest1234\n"))

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

		io.SetIn(strings.NewReader("n\ntest1234\ntest1234\n"))

		// Confirm overwrite
		cmd = NewRootCmdWithBaseConfig(io, baseOptions)
		require.ErrorIs(t, cmd.ParseAndRun(ctx, args), errOverwriteAborted)

		newKey, err := kb.GetByName(keyName)
		require.NoError(t, err)

		// Make sure the key is not overwritten
		assert.Equal(t, original.GetAddress(), newKey.GetAddress())
	})
}

func generateDerivationPaths(count int) []string {
	paths := make([]string, count)

	for i := 0; i < count; i++ {
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
