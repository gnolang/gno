package client

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
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
		var sb strings.Builder
		sb.WriteString(mnemonic)
		sb.WriteString("\n")
		for range paths {
			sb.WriteString(dummyPass)
			sb.WriteString("\n")
			sb.WriteString(dummyPass)
			sb.WriteString("\n")
		}
		io.SetIn(strings.NewReader(sb.String()))
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

	t.Run("derivation paths create keybase entries", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome   = t.TempDir()
			mnemonic = generateTestMnemonic(t)
			paths    = generateDerivationPaths(3)

			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}

			dummyPass = "dummy-pass"
			keyName   = "example-key"
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		io := commands.NewTestIO()
		var sb strings.Builder
		sb.WriteString(mnemonic)
		sb.WriteString("\n")
		for range paths {
			sb.WriteString(dummyPass)
			sb.WriteString("\n")
			sb.WriteString(dummyPass)
			sb.WriteString("\n")
		}
		io.SetIn(strings.NewReader(sb.String()))

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

		for _, path := range paths {
			args = append(
				args, []string{
					"--derivation-path",
					path,
				}...,
			)
		}

		require.NoError(t, cmd.ParseAndRun(ctx, args))

		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		for _, path := range paths {
			params, err := hd.NewParamsFromPath(path)
			require.NoError(t, err)

			derivedName := deriveKeyName(keyName, params, len(paths))

			has, err := kb.HasByName(derivedName)
			require.NoError(t, err)
			require.True(t, has)
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

	t.Run("derivation path overwrite aborted", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			mnemonic    = generateTestMnemonic(t)
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}
			keyName = "example-key"
			path    = "44'/118'/0'/0/0"
		)

		// Pre-create a key that will collide with the derived name.
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		_, err = kb.CreateAccount(keyName, mnemonic, "", "encrypt", 0, 0)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("n\n"))

		// Create the command
		cmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"add",
			"--insecure-password-stdin",
			"--home",
			kbHome,
			"--recover",
			keyName,
			"--derivation-path",
			path,
		}

		require.ErrorIs(t, cmd.ParseAndRun(ctx, args), errOverwriteAborted)
	})

	t.Run("derivation path passphrase preflight", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			mnemonic    = generateTestMnemonic(t)
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}
			keyName = "example-key"
			paths   = []string{"44'/118'/0'/0/0", "44'/118'/0'/0/1"}
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		io := commands.NewTestIO()
		// First passphrase OK, second mismatch triggers errPassphraseMismatch.
		io.SetIn(strings.NewReader(mnemonic + "\npass1\npass1\npass2\npass3\n"))

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

		for _, path := range paths {
			args = append(
				args, []string{
					"--derivation-path",
					path,
				}...,
			)
		}

		require.ErrorIs(t, cmd.ParseAndRun(ctx, args), errPassphraseMismatch)

		// Ensure no derived keys were persisted.
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		for _, path := range paths {
			params, err := hd.NewParamsFromPath(path)
			require.NoError(t, err)

			derivedName := deriveKeyName(keyName, params, len(paths))
			has, err := kb.HasByName(derivedName)
			require.NoError(t, err)
			require.False(t, has)
		}
	})
}
