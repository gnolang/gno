package client

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
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

	// Both shapes below used to panic out of the command with a stack trace.
	// execAddMultisig pre-checked the threshold and nothing else, so every
	// further condition validate grew reached the constructor unguarded; the
	// checked constructor is what keeps them errors without that pre-check
	// having to mirror them one at a time.
	t.Run("invalid multisig key reported as an error", func(t *testing.T) {
		t.Parallel()

		var (
			kbHome      = t.TempDir()
			baseOptions = BaseOptions{
				InsecurePasswordStdin: true,
				Home:                  kbHome,
			}
		)

		kb, err := keys.NewKeyBaseFromDir(kbHome)
		require.NoError(t, err)

		// MaxTotalKeys counts the multisig key itself alongside its
		// constituents, so MaxTotalKeys constituents is one key past the bound.
		// These are pubkey-only references: --multisig reads nothing but
		// GetPubKey, and CreateOffline skips the armor encryption a full
		// account would pay for on each of them.
		keyNames := make([]string, multisig.MaxTotalKeys)
		for index := range keyNames {
			keyNames[index] = fmt.Sprintf("key-%d", index)

			_, err = kb.CreateOffline(keyNames[index], secp256k1.GenPrivKey().PubKey())
			require.NoError(t, err)
		}

		tests := []struct {
			name    string
			signers []string
		}{
			{"more keys in total than MaxTotalKeys", keyNames},
			{"duplicate constituent key", []string{keyNames[0], keyNames[0]}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancelFn()

				args := []string{"add", "multisig", "--insecure-password-stdin", "--home", kbHome}
				for _, signer := range tt.signers {
					args = append(args, "--multisig", signer)
				}
				args = append(args, "multi")

				io := commands.NewTestIO()
				io.SetIn(strings.NewReader("y\n"))
				cmd := NewRootCmdWithBaseConfig(io, baseOptions)

				require.NotPanics(t, func() {
					require.ErrorContains(t, cmd.ParseAndRun(ctx, args),
						"unable to construct multisig public key")
				})
			})
		}
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

		// Name collision: multisig key uses same name "key-1" as existing account
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader("y\n"))

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

	t.Run("multisig address collision, decline", func(t *testing.T) {
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
		// Same address + same type (multi) + different name → rename prompt, decline
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

	t.Run("multisig address collision, confirm rename", func(t *testing.T) {
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

		// Get the original multisig key address
		multi1, err := kb.GetByName("multi-1")
		require.NoError(t, err)

		// Create second multisig with same keys, different name → rename prompt, confirm
		io.SetIn(strings.NewReader("y\n"))

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

		require.NoError(t, cmd.ParseAndRun(ctx, args2))

		// Verify multi-1 was renamed to multi-2
		renamedKey, err := kb.GetByName("multi-2")
		require.NoError(t, err)
		assert.Equal(t, multi1.GetAddress(), renamedKey.GetAddress())
		assert.Equal(t, keys.TypeMulti, renamedKey.GetType())

		// multi-1 should no longer exist
		_, err = kb.GetByName("multi-1")
		require.Error(t, err)
	})
}
