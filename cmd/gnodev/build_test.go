package main

import "testing"

func TestBuildApp(t *testing.T) {
	tc := []testMainCase{
		{args: []string{"build"}, errShouldBe: "invalid args", stderrShouldBe: "Usage: build [build flags] [packages]\n"},
		{args: []string{"build", "--help"}, stdoutShouldContain: "# buildOptions options\n-"},

		// args
		// {args: []string{"build", "..."}, stdoutShouldContain: "..."},
	}
	testMainCaseRun(t, tc)
}
