package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestServerInitialize(t *testing.T) {
	cases := []struct {
		args []string
	}{
		{[]string{"--skip-start", "--skip-failing-genesis-txs"}},
		// {[]string{"--skip-start"}},
		// FIXME: test seems flappy as soon as we have multiple cases.
	}
	os.Chdir(filepath.Join("..", "..")) // go to repo's root dir

	for _, tc := range cases {
		name := strings.Join(tc.args, " ")
		t.Run(name, func(t *testing.T) {
			mockOut := bytes.NewBufferString("")
			mockErr := bytes.NewBufferString("")
			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOut))
			io.SetErr(commands.WriteNopCloser(mockErr))

			closer := testutils.CaptureStdoutAndStderr()

			cfg := &serverCfg{}
			cmd := commands.NewCommand(
				commands.Metadata{},
				cfg,
				func(_ context.Context, args []string) error {
					return execServer(cfg, args, io)
				},
			)

			t.Logf(`Running "gnoland %s"`, strings.Join(tc.args, " "))
			err := cmd.ParseAndRun(context.Background(), tc.args)
			require.NoError(t, err)

			stdouterr, bufErr := closer()
			require.NoError(t, bufErr)
			require.NoError(t, err)

			require.Contains(t, stdouterr, "Node created.", "failed to create node")
			require.Contains(t, stdouterr, "'--skip-start' is set. Exiting.", "not exited with skip-start")
			require.NotContains(t, stdouterr, "panic:")
		})
	}
}

// TODO: test various configuration files?
