package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStartInitialize(t *testing.T) {
	cases := []struct {
		args []string
	}{
		{[]string{"start", "--skip-start", "--skip-failing-genesis-txs"}},
		// {[]string{"--skip-start"}},
		// FIXME: test seems flappy as soon as we have multiple cases.
	}
	os.Chdir(filepath.Join("..", "..")) // go to repo's root dir

	for _, tc := range cases {
		name := strings.Join(tc.args, " ")
		in, err := NewMockStdin(name)
		if err != nil {
			t.Fatal("failed creating test io pipe")
		}

		t.Run(name, func(t *testing.T) {
			cmd := newRootCmd()
			t.Logf(`Running "gnoland %s"`, strings.Join(tc.args, " "))
			err := cmd.ParseAndRun(context.Background(), tc.args)
			require.NoError(t, err)

			bz, err := in.ReadAndClose()
			if err != nil {
				t.Fatal("failed reading test io pipe")
			}

			out := string(bz)

			require.Contains(t, out, "Node created.", "failed to create node")

			require.Contains(t, out, "'--skip-start' is set. Exiting.", "not exited with skip-start")
			require.NotContains(t, out, "panic:")
		})
	}
}

// TODO: test various configuration files?
