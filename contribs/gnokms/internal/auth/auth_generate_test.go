package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
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

		// Run the command.
		cmdErr := cmd.ParseAndRun(context.Background(), []string{})
		require.NoError(t, cmdErr)

		// Check the command output.
		require.Contains(t, buffer.String(), flags.AuthKeysFile)
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

		// Run the command.
		cmdErr := cmd.ParseAndRun(context.Background(), []string{})
		require.Error(t, cmdErr)
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

		// Run the command.
		cmdErr := cmd.ParseAndRun(context.Background(), []string{})
		require.Error(t, cmdErr)

		// Run exec with overwrite flag.
		cmdErr = execAuthGenerate(
			&authGenerateFlags{auth: flags, overwrite: true},
			commands.NewDefaultIO(),
		)
		require.NoError(t, cmdErr)
	})
}
