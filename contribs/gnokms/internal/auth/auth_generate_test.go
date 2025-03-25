package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	t.Run("non-existent auth key file", func(t *testing.T) {
		t.Parallel()

		// Create the buffered command output.
		buffer, io := createBufferedCmdOutput(t)

		// Create the command flags with a non-existent auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filepath.Join(t.TempDir(), "non-existent"),
		}

		// Create the command.
		cmd := newAuthGenerateCmd(flags, io)

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command.
		cmdErr := cmd.ParseAndRun(ctx, []string{})
		require.NoError(t, cmdErr)

		// Check the command output.
		assert.Contains(t, buffer.String(), flags.AuthKeysFile)
	})

	t.Run("read-only auth key file path", func(t *testing.T) {
		t.Parallel()

		// Create a read-only directory.
		readOnlyDir := filepath.Join(t.TempDir(), "read-only")
		require.NoError(t, os.MkdirAll(readOnlyDir, 0o500))

		// Create the command flags with a read-only file path.
		flags := &common.AuthFlags{
			AuthKeysFile: filepath.Join(readOnlyDir, "non-existent"),
		}

		// Create the command.
		cmd := newAuthGenerateCmd(flags, commands.NewTestIO())

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command.
		cmdErr := cmd.ParseAndRun(ctx, []string{})
		assert.Error(t, cmdErr)
	})

	t.Run("existent auth key file path", func(t *testing.T) {
		t.Parallel()

		// Create the auth key file and a buffered command output.
		filePath, _ := createAuthKeysFile(t)

		// Create the command flags with a read-only file path.
		flags := &common.AuthFlags{
			AuthKeysFile: filePath,
		}

		// Create the command.
		cmd := newAuthGenerateCmd(flags, commands.NewTestIO())

		// Create a context with a 5s timeout.
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Run the command.
		cmdErr := cmd.ParseAndRun(ctx, []string{})
		require.Error(t, cmdErr)

		// Run exec with overwrite flag.
		cmdErr = execAuthGenerate(
			&authGenerateFlags{auth: flags, overwrite: true},
			commands.NewTestIO(),
		)
		assert.NoError(t, cmdErr)
	})
}
