package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
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
			closer := testutils.CaptureStdoutAndStderr()

			cfg := &gnolandCfg{}
			cmd := commands.NewCommand(
				commands.Metadata{},
				cfg,
				func(_ context.Context, _ []string) error {
					return exec(cfg)
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

func TestSortPkgs(t *testing.T) {
	for _, tc := range []struct {
		desc      string
		in        []pkg
		expected  []string
		shouldErr bool
	}{
		{
			desc:     "empty_input",
			in:       []pkg{},
			expected: make([]string, 0),
		}, {
			desc: "no_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{}},
				{name: "pkg2", path: "/path/to/pkg2", requires: []string{}},
				{name: "pkg3", path: "/path/to/pkg3", requires: []string{}},
			},
			expected: []string{"pkg1", "pkg2", "pkg3"},
		}, {
			desc: "circular_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{"pkg2"}},
				{name: "pkg2", path: "/path/to/pkg2", requires: []string{"pkg1"}},
			},
			shouldErr: true,
		}, {
			desc: "missing_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{"pkg2"}},
			},
			shouldErr: true,
		}, {
			desc: "valid_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{"pkg2"}},
				{name: "pkg2", path: "/path/to/pkg2", requires: []string{"pkg3"}},
				{name: "pkg3", path: "/path/to/pkg3", requires: []string{}},
			},
			expected: []string{"pkg3", "pkg2", "pkg1"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := sortPkgs(tc.in)
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for i := range tc.expected {
					assert.Equal(t, tc.expected[i], tc.in[i].name)
				}
			}
		})
	}
}

// TODO: test various configuration files?
