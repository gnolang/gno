package auth

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentity(t *testing.T) {
	t.Parallel()

	t.Run("non-existent auth key file", func(t *testing.T) {
		t.Parallel()

		// Create the command flags with a non-existent auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filepath.Join(t.TempDir(), "non-existent"),
		}

		// Create the command.
		cmd := newAuthIdentityCmd(flags, commands.NewTestIO())

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
		buffer, io := createBufferedCmdOutput(t)

		// Create the command flags with the auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create the command.
		cmd := newAuthIdentityCmd(flags, io)

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command.
		cmdErr := cmd.ParseAndRun(ctx, []string{})
		require.NoError(t, cmdErr)

		// Check the command output.
		assert.Contains(t, buffer.String(), authKeysFile.ServerIdentity.PubKey)
	})

	t.Run("valid auth key file with raw flag", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file and a buffered command output.
		filePath, authKeysFile := createAuthKeysFile(t)
		buffer, io := createBufferedCmdOutput(t)

		// Create the command flags with the auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create the command.
		cmd := newAuthIdentityCmd(flags, io)

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command with the --raw flag.
		cmdErr := cmd.ParseAndRun(ctx, []string{"--raw"})
		require.NoError(t, cmdErr)

		// Check the command output contains only the raw public key without description.
		output := buffer.String()
		assert.Contains(t, output, authKeysFile.ServerIdentity.PubKey)
		assert.NotContains(t, output, "Server public key:")
	})
}
