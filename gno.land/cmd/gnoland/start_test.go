package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func TestStartInitialize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		args []string
	}{
		{[]string{"start", "--skip-start", "--skip-failing-genesis-txs"}},
		// {[]string{"--skip-start"}},
		// FIXME: test seems flappy as soon as we have multiple cases.
	}
	os.Chdir(filepath.Join("..", "..")) // go to repo's root dir

	for _, tc := range cases {
		tc := tc
		name := strings.Join(tc.args, " ")
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockOut := new(bytes.Buffer)
			mockErr := new(bytes.Buffer)
			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOut))
			io.SetErr(commands.WriteNopCloser(mockErr))
			cmd := newRootCmd(io)

			t.Logf(`Running "gnoland %s"`, strings.Join(tc.args, " "))
			err := cmd.ParseAndRun(context.Background(), tc.args)
			require.NoError(t, err)

			stdout := mockOut.String()
			stderr := mockErr.String()

			require.Contains(t, stderr, "Node created.", "failed to create node")
			require.Contains(t, stderr, "'--skip-start' is set. Exiting.", "not exited with skip-start")
			require.NotContains(t, stdout, "panic:")
		})
	}
}

// TODO: test various configuration files?
