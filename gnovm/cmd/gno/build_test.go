package main

import "testing"

func TestBuildApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"build"},
			errShouldBe: "flag: help requested",
		},

		// {args: []string{"build", "..."}, stdoutShouldContain: "..."},
		// TODO: auto precompilation
		// TODO: error handling
	}
	testMainCaseRun(t, tc)
}
