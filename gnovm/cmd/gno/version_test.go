package main

import (
	"testing"
)

func TestVersionApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:                []string{"version"},
			testDir:             "testdata/gno_version/gopath_version.txtar",
			stdoutShouldContain: "gno version: v0.2.0",
		},
		{
			args:                []string{"version"},
			testDir:             "testdata/gno_version/git_version.txtar",
			stdoutShouldContain: "gno version: v0.2.0",
		},
	}

	testMainCaseRun(t, tc)
}
