package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizedAdd(t *testing.T) {
	t.Parallel()

	t.Run("non-existent auth key file", func(t *testing.T) {
		t.Parallel()

		// Create the command flags with a non-existent auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filepath.Join(t.TempDir(), "non-existent"),
		}

		// Create the command.
		cmd := newAuthAuthorizedAddCmd(flags, commands.NewTestIO())

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command.
		cmdErr := cmd.ParseAndRun(ctx, []string{})
		assert.Error(t, cmdErr)
	})

	t.Run("valid auth key file", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file and a buffered command output.
		filePath, authKeysFile := createAuthKeysFile(t)
		require.Equal(t, 0, len(authKeysFile.ClientAuthorizedKeys))

		// Create the command flags with the auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create a random authorized key.
		authorizedKey := ed25519.GenPrivKey().PubKey().String()

		// Run exec authorized key argument.
		buffer, io := createBufferedCmdOutput(t)
		err := execAuthAuthorizedAdd(flags, []string{authorizedKey}, io)
		require.NoError(t, err)

		// Check the first command output.
		require.Contains(t, buffer.String(), "added to the authorized keys list.")
		require.Contains(t, buffer.String(), authorizedKey)
		authKeysFile, err = common.LoadAuthKeysFile(flags.AuthKeysFile)
		require.NotNil(t, authKeysFile)
		require.NoError(t, err)
		require.Equal(t, 1, len(authKeysFile.ClientAuthorizedKeys))

		// Rerun exec authorized key argument.
		buffer, io = createBufferedCmdOutput(t)
		err = execAuthAuthorizedAdd(flags, []string{authorizedKey}, io)
		require.NoError(t, err)

		// Check the second command output.
		require.Contains(t, buffer.String(), "already in the authorized keys list.")
		require.Contains(t, buffer.String(), authorizedKey)
		authKeysFile, err = common.LoadAuthKeysFile(flags.AuthKeysFile)
		require.NotNil(t, authKeysFile)
		require.NoError(t, err)
		assert.Equal(t, 1, len(authKeysFile.ClientAuthorizedKeys))
	})

	t.Run("valid auth key file with invalid argument", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file.
		filePath, _ := createAuthKeysFile(t)

		// Create the command flags with the auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create a random authorized key.
		authorizedKey := "invalid"

		// Run exec authorized key argument.
		err := execAuthAuthorizedAdd(flags, []string{authorizedKey}, commands.NewTestIO())
		assert.Error(t, err)
	})

	t.Run("valid auth key file stored in read-only directory", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file.
		filePath, _ := createAuthKeysFile(t)

		// Make it read-only.
		require.NoError(t, os.Chmod(filepath.Dir(filePath), 0o500))

		// Create the command flags with a read-only file path.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create a random authorized key.
		authorizedKey := ed25519.GenPrivKey().PubKey().String()

		// Run exec authorized key argument.
		err := execAuthAuthorizedAdd(flags, []string{authorizedKey}, commands.NewTestIO())
		require.Error(t, err)

		// Turn it back to read-write for cleanup.
		assert.NoError(t, os.Chmod(filepath.Dir(filePath), 0o700))
	})
}

func TestAuthorizedRemove(t *testing.T) {
	t.Parallel()

	t.Run("non-existent auth key file", func(t *testing.T) {
		t.Parallel()

		// Create the command flags with a non-existent auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filepath.Join(t.TempDir(), "non-existent"),
		}

		// Create the command.
		cmd := newAuthAuthorizedRemoveCmd(flags, commands.NewTestIO())

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command.
		cmdErr := cmd.ParseAndRun(ctx, []string{})
		assert.Error(t, cmdErr)
	})

	t.Run("valid auth key file", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file and a buffered command output.
		filePath, authKeysFile := createAuthKeysFile(t)
		require.Equal(t, 0, len(authKeysFile.ClientAuthorizedKeys))

		// Create the command flags with the auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create a random authorized key.
		authorizedKey := ed25519.GenPrivKey().PubKey().String()

		// Add authorized key to the auth key file.
		authKeysFile.ClientAuthorizedKeys = []string{authorizedKey}

		// Save the auth key file.
		require.NoError(t, authKeysFile.Save(filePath))
		require.Equal(t, 1, len(authKeysFile.ClientAuthorizedKeys))

		// Run exec authorized key argument.
		buffer, io := createBufferedCmdOutput(t)
		err := execAuthAuthorizedRemove(flags, []string{authorizedKey}, io)
		require.NoError(t, err)

		// Check the first command output.
		require.Contains(t, buffer.String(), "removed from the authorized keys list.")
		require.Contains(t, buffer.String(), authorizedKey)
		authKeysFile, err = common.LoadAuthKeysFile(flags.AuthKeysFile)
		require.NotNil(t, authKeysFile)
		require.NoError(t, err)
		require.Equal(t, 0, len(authKeysFile.ClientAuthorizedKeys))

		// Rerun exec authorized key argument.
		buffer, io = createBufferedCmdOutput(t)
		err = execAuthAuthorizedRemove(flags, []string{authorizedKey}, io)
		require.NoError(t, err)

		// Check the second command output.
		require.Contains(t, buffer.String(), "not found in the authorized keys list.")
		require.Contains(t, buffer.String(), authorizedKey)
		authKeysFile, err = common.LoadAuthKeysFile(flags.AuthKeysFile)
		require.NotNil(t, authKeysFile)
		require.NoError(t, err)
		assert.Equal(t, 0, len(authKeysFile.ClientAuthorizedKeys))
	})

	t.Run("valid auth key file with invalid argument", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file.
		filePath, _ := createAuthKeysFile(t)

		// Create the command flags with the auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create a random authorized key.
		authorizedKey := "invalid"

		// Run exec authorized key argument.
		err := execAuthAuthorizedRemove(flags, []string{authorizedKey}, commands.NewTestIO())
		assert.Error(t, err)
	})

	t.Run("valid auth key file stored in read-only directory", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file.
		filePath, _ := createAuthKeysFile(t)

		// Make it read-only.
		require.NoError(t, os.Chmod(filepath.Dir(filePath), 0o500))

		// Create the command flags with a read-only file path.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create a random authorized key.
		authorizedKey := ed25519.GenPrivKey().PubKey().String()

		// Run exec authorized key argument.
		err := execAuthAuthorizedRemove(flags, []string{authorizedKey}, commands.NewTestIO())
		require.Error(t, err)

		// Turn it back to read-write for cleanup.
		assert.NoError(t, os.Chmod(filepath.Dir(filePath), 0o700))
	})
}

func TestAuthorizedList(t *testing.T) {
	t.Parallel()

	t.Run("non-existent auth key file", func(t *testing.T) {
		t.Parallel()

		// Create the command flags with a non-existent auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filepath.Join(t.TempDir(), "non-existent"),
		}

		// Create the command.
		cmd := newAuthAuthorizedListCmd(flags, commands.NewTestIO())

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command.
		cmdErr := cmd.ParseAndRun(ctx, []string{})
		assert.Error(t, cmdErr)
	})

	t.Run("valid auth key file without authorized keys", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file and a buffered command output.
		filePath, _ := createAuthKeysFile(t)
		buffer, io := createBufferedCmdOutput(t)

		// Create the command flags with the auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create the command.
		cmd := newAuthAuthorizedListCmd(flags, io)

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command.
		cmdErr := cmd.ParseAndRun(ctx, []string{})
		require.NoError(t, cmdErr)

		// Check the command output.
		assert.Contains(t, buffer.String(), "No authorized keys found.")
	})

	t.Run("valid auth key file with authorized keys", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file and a buffered command output.
		filePath, authKeysFile := createAuthKeysFile(t)
		buffer, io := createBufferedCmdOutput(t)

		// Add authorized keys to the auth key file.
		for range 3 {
			authorizedKey := ed25519.GenPrivKey().PubKey().String()
			authKeysFile.ClientAuthorizedKeys = append(authKeysFile.ClientAuthorizedKeys, authorizedKey)
		}

		// Save the auth key file.
		require.NoError(t, authKeysFile.Save(filePath))

		// Create the command flags with the auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create the command.
		cmd := newAuthAuthorizedListCmd(flags, io)

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command.
		cmdErr := cmd.ParseAndRun(ctx, []string{})
		require.NoError(t, cmdErr)

		// Check the command output.
		require.Contains(t, buffer.String(), "Authorized keys:")
		for _, authorizedKey := range authKeysFile.ClientAuthorizedKeys {
			require.Contains(t, buffer.String(), authorizedKey)
		}
	})

	t.Run("valid auth key file without authorized keys with raw flag", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file and a buffered command output.
		filePath, _ := createAuthKeysFile(t)
		buffer, io := createBufferedCmdOutput(t)

		// Create the command flags with the auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create the command.
		cmd := newAuthAuthorizedListCmd(flags, io)

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command with the --raw flag.
		cmdErr := cmd.ParseAndRun(ctx, []string{"--raw"})
		require.NoError(t, cmdErr)

		// Check that the command output is empty (no "No authorized keys found." message).
		assert.Empty(t, buffer.String())
	})

	t.Run("valid auth key file with authorized keys with raw flag", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file and a buffered command output.
		filePath, authKeysFile := createAuthKeysFile(t)
		buffer, io := createBufferedCmdOutput(t)

		// Add authorized keys to the auth key file.
		for range 3 {
			authorizedKey := ed25519.GenPrivKey().PubKey().String()
			authKeysFile.ClientAuthorizedKeys = append(authKeysFile.ClientAuthorizedKeys, authorizedKey)
		}

		// Save the auth key file.
		require.NoError(t, authKeysFile.Save(filePath))

		// Create the command flags with the auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create the command.
		cmd := newAuthAuthorizedListCmd(flags, io)

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command with the --raw flag.
		cmdErr := cmd.ParseAndRun(ctx, []string{"--raw"})
		require.NoError(t, cmdErr)

		// Check the command output contains raw keys without formatting.
		output := buffer.String()
		assert.NotContains(t, output, "Authorized keys:")
		assert.NotContains(t, output, "-")
		for _, authorizedKey := range authKeysFile.ClientAuthorizedKeys {
			require.Contains(t, output, authorizedKey)
		}
	})
}

func TestAuthorizedCmd(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, newAuthAuthorizedCmd(nil, commands.NewTestIO()))
}
