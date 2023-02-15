package main

import "testing"

func TestPrecompileApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:           []string{"precompile"},
			errShouldBe:    "invalid args",
			stderrShouldBe: "Usage: precompile [precompile flags] [packages]\n",
		}, {
			args:                []string{"precompile", "--help"},
			stdoutShouldContain: "# precompileFlags options\n-",
		},

		// {args: []string{"precompile", "..."}, stdoutShouldContain: "..."},
		// TODO: recursive
		// TODO: valid files
		// TODO: invalid files
	}
	testMainCaseRun(t, tc)
}
