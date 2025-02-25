package gnokey

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func TestNewGnokeyCmd(t *testing.T) {
	t.Parallel()

	t.Run("without keyname", func(t *testing.T) {
		t.Parallel()

		cmd := NewGnokeyCmd(commands.NewTestIO())
		require.NotNil(t, cmd)

		require.Error(t, cmd.ParseAndRun(context.Background(), []string{}))
	})

	t.Run("unknown keyname", func(t *testing.T) {
		t.Parallel()

		cmd := NewGnokeyCmd(commands.NewTestIO())
		require.NotNil(t, cmd)

		require.Error(t, cmd.ParseAndRun(context.Background(), []string{"unknown"}))
	})

	t.Run("valid keyname", func(t *testing.T) {
		t.Parallel()

		// Generate a keybase with a key.
		filePath, keybase := generateKeyBaseWithKey(t)
		defer keybase.CloseDB()

		// Create a stdin with the password.
		io := commands.NewDefaultIO()
		io.SetIn(strings.NewReader(fmt.Sprintf("%s\n", keyPassword)))

		cmd := NewGnokeyCmd(io)
		require.Error(
			t,
			cmd.ParseAndRun(
				context.Background(),
				[]string{
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
