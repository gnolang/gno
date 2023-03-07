package main

import "testing"

func TestPrecompileApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"precompile"},
			errShouldBe: "flag: help requested",
		},

		// {args: []string{"precompile", "..."}, stdoutShouldContain: "..."},
		// TODO: recursive
		// TODO: valid files
		// TODO: invalid files
	}
	testMainCaseRun(t, tc)
}
