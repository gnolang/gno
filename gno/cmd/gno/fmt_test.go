package main

import "testing"

func TestFmtApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"fmt"},
			errShouldBe: "flag: help requested",
		}, {
			args:                []string{"fmt", "../../tests/integ/unformated/missing_import.gno"},
			stdoutShouldContain: "strconv",
		},

		// XXX: more complex output are tested in `testdata/gno_test/fmt_*.txtar`.
	}
	testMainCaseRun(t, tc)
}
