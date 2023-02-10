package main

import "testing"

func TestRunApp(t *testing.T) {
	tc := []testMainCase{
		{args: []string{"run"}, errShouldBe: "invalid args", stderrShouldBe: "Usage: run [flags] file.gno [file2.gno...]\n"},
		{args: []string{"run", "--help"}, stdoutShouldContain: "# runOptions options\n-"},

		{args: []string{"run", "../../tests/integ/run-main/main.gno"}, stdoutShouldContain: "hello world!"},
	}
	testMainCaseRun(t, tc)
}
