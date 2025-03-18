package gnokey

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGnokeyCmd(t *testing.T) {
	t.Parallel()

	t.Run("without keyname", func(t *testing.T) {
		t.Parallel()

		cmd := NewGnokeyCmd(commands.NewTestIO())
		require.NotNil(t, cmd)
		cmd.SetOutput(commands.WriteNopCloser(new(bytes.Buffer)))

		assert.Error(t, cmd.ParseAndRun(context.Background(), []string{}))
	})

	t.Run("unknown keyname", func(t *testing.T) {
		t.Parallel()

		cmd := NewGnokeyCmd(commands.NewTestIO())
		require.NotNil(t, cmd)

		assert.Error(t, cmd.ParseAndRun(context.Background(), []string{"unknown"}))
	})

	t.Run("valid keyname with wrong address", func(t *testing.T) {
		t.Parallel()

		// Generate a keybase with a key.
		filePath, keybase := generateKeyBaseWithKey(t)
		defer keybase.CloseDB()

		// Create a stdin with the password.
		io := commands.NewTestIO()
		io.SetIn(strings.NewReader(fmt.Sprintf("%s\n", keyPassword)))

		cmd := NewGnokeyCmd(io)
		assert.Error(
			t,
			cmd.ParseAndRun(
				context.Background(),
				[]string{
					"--log-level",
					"error",
					"--listeners",
					"wrong_address",
					"--home",
					filePath,
					"--insecure-password-stdin",
					keyName,
				},
			),
		)
	})
}
