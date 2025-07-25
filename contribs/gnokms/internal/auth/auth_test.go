package auth

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createAuthKeysFile(t *testing.T) (string, *common.AuthKeysFile) {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), xid.New().String())
	authKeys, err := common.GeneratePersistedAuthKeysFile(filePath)
	require.NotNil(t, authKeys)
	require.NoError(t, err)

	return filePath, authKeys
}

func createBufferedCmdOutput(t *testing.T) (*bytes.Buffer, commands.IO) {
	t.Helper()

	buffer := new(bytes.Buffer)
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(buffer))

	return buffer, io
}

func TestLoadAuthKeysFile(t *testing.T) {
	t.Parallel()

	t.Run("non-existent auth key file", func(t *testing.T) {
		t.Parallel()

		// Create the command flags with a non-existent auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filepath.Join(t.TempDir(), "non-existent"),
		}

		// Try to load the auth keys file.
		authKeysFile, err := loadAuthKeysFile(flags)
		require.Nil(t, authKeysFile)
		assert.Error(t, err)
	})

	t.Run("invalid auth key file", func(t *testing.T) {
		t.Parallel()

		// Create the command flags with a non-existent auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filepath.Join(t.TempDir(), "invalid"),
		}

		// Create the invalid auth key file.
		os.WriteFile(flags.AuthKeysFile, []byte("invalid"), 0e600)

		// Try to load the auth keys file.
		authKeysFile, err := loadAuthKeysFile(flags)
		require.Nil(t, authKeysFile)
		assert.Error(t, err)
	})

	t.Run("valid auth key file", func(t *testing.T) {
		t.Parallel()

		// Create the command flags with a non-existent auth key file.
		flags := &common.AuthFlags{
			AuthKeysFile: filepath.Join(t.TempDir(), "invalid"),
		}

		// Create the valid auth key file.
		common.GeneratePersistedAuthKeysFile(flags.AuthKeysFile)

		// Try to load the auth keys file.
		authKeysFile, err := loadAuthKeysFile(flags)
		require.NotNil(t, authKeysFile)
		assert.NoError(t, err)
	})
}
