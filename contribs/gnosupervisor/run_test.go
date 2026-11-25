package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func TestRunFlags(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		args []string
		want string
	}{
		"no ledger":            {[]string{}, "--ledger is required"},
		"owned data-dir":       {[]string{"--ledger", "x", "--", "--data-dir", "/tmp/x"}, "--data-dir is set by the supervisor"},
		"owned genesis, =":     {[]string{"--ledger", "x", "--", "-genesis=g.json"}, "--genesis is set by the supervisor"},
		"owned gnoroot-dir":    {[]string{"--ledger", "x", "--", "--log-level", "info", "--gnoroot-dir", "/g"}, "--gnoroot-dir is set by the supervisor"},
		"unknown log format":   {[]string{"--ledger", "x", "--log-format", "xml"}, `--log-format "xml"`},
		"platform with a path": {[]string{"--ledger", "x", "--platform", "linux/../../x"}, "is not os/arch"},
		"ledger that is gone":  {[]string{"--ledger", "/nonexistent/upgrades.json", "--data-dir", t.TempDir()}, "read ledger"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			io := commands.NewTestIO()
			err := newRunCmd(io).ParseAndRun(context.Background(), tc.args)
			assert.ErrorContains(t, err, tc.want)
		})
	}
}
